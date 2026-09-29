package crawler

import (
	"context"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"web-crawler/internal/document"
	"web-crawler/internal/fetcher"
	"web-crawler/internal/frontier"
	"web-crawler/internal/parser"
	htmlparser "web-crawler/internal/parser/html"
	pdfparser "web-crawler/internal/parser/pdf"
	"web-crawler/internal/resource"
	"web-crawler/internal/storage"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Start    url.URL
	MaxPages int
	Timeout  time.Duration
	Delay    time.Duration
}

type DocumentStore interface {
	Save(context.Context, *document.Document) error
}

// ParseConfig validates command-line arguments without opening external connections.
func ParseConfig(args []string, output io.Writer) (Config, error) {
	var cfg Config
	flags := flag.NewFlagSet("crawler", flag.ContinueOnError)
	flags.SetOutput(output)
	flags.IntVar(&cfg.MaxPages, "max-pages", 100, "maximum number of queued URLs")
	flags.DurationVar(&cfg.Timeout, "timeout", 15*time.Second, "timeout per HTTP request and database operation")
	flags.DurationVar(&cfg.Delay, "delay", 200*time.Millisecond, "delay between pages")
	flags.Usage = func() {
		fmt.Fprintln(output, "Usage: crawler [flags] <http(s)://start-url>\nRequires DATABASE_URL for PostgreSQL. Flags must precede the URL.")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		return cfg, err
	}
	if flags.NArg() != 1 {
		flags.Usage()
		return cfg, fmt.Errorf("provide exactly one starting URL")
	}
	if cfg.MaxPages <= 0 || cfg.Timeout <= 0 || cfg.Delay < 0 {
		return cfg, fmt.Errorf("max-pages and timeout must be positive; delay must be non-negative")
	}
	start, err := url.Parse(flags.Arg(0))
	if err != nil {
		return cfg, fmt.Errorf("parse starting URL: %w", err)
	}
	cfg.Start, err = resource.NormalizeURL(*start)
	if err != nil {
		return cfg, fmt.Errorf("invalid starting URL: %w", err)
	}
	return cfg, nil
}

func Run(ctx context.Context, args []string, databaseURL string, output io.Writer) error {
	cfg, err := ParseConfig(args, output)
	if err != nil {
		return err
	}
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	initCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()
	pool, err := pgxpool.New(initCtx, databaseURL)
	if err != nil {
		return fmt.Errorf("invalid PostgreSQL connection configuration")
	}
	defer pool.Close()
	if err := pool.Ping(initCtx); err != nil {
		return fmt.Errorf("connect to PostgreSQL: %w", err)
	}
	store := storage.NewStorage(pool)
	if err := store.Init(initCtx); err != nil {
		return err
	}
	cancel()
	return Crawl(ctx, cfg, store)
}

func Crawl(ctx context.Context, cfg Config, store DocumentStore) error {
	client := &http.Client{
		Timeout: cfg.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			normalized, err := resource.NormalizeURL(*req.URL)
			if err != nil {
				return err
			}
			if normalized.Host != cfg.Start.Host {
				return fmt.Errorf("redirect outside starting host")
			}
			return nil
		},
	}
	defer client.CloseIdleConnections()
	loader := fetcher.NewFetcher(client)
	parsers := []parser.Parser{htmlparser.HTMLParser{}, pdfparser.PDFParser{}}
	queue := frontier.Frontier{MaxPages: cfg.MaxPages, Delay: cfg.Delay}
	attempted, saved, skipped := 0, 0, 0
	err := queue.Crawl(ctx, cfg.Start, func(target *url.URL) (*document.Document, error) {
		attempted++
		res, err := loader.FetchResource(ctx, *target)
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			skipped++
			slog.Warn("fetch failed", "url", target.String(), "error", err)
			return nil, nil
		}
		for _, p := range parsers {
			if !p.CanParse(*res) {
				continue
			}
			doc, err := p.Parse(ctx, *res)
			if err != nil {
				if ctx.Err() != nil {
					return nil, ctx.Err()
				}
				skipped++
				slog.Warn("parse failed", "url", target.String(), "error", err)
				return nil, nil
			}
			doc.ContentHash = fmt.Sprintf("%x", sha256.Sum256(res.Data))
			doc.CrawledAt = time.Now().UTC()
			saveCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
			err = store.Save(saveCtx, doc)
			cancel()
			if err != nil {
				return nil, err
			}
			saved++
			slog.Info("document saved", "url", doc.URL.String(), "title", doc.Title)
			return doc, nil
		}
		skipped++
		slog.Info("unsupported content type", "url", target.String(), "content_type", res.ContentType)
		return nil, nil
	})
	slog.Info("crawl finished", "attempted", attempted, "saved", saved, "skipped", skipped)
	if err != nil {
		return err
	}
	if saved == 0 {
		return fmt.Errorf("no documents saved")
	}
	return nil
}
