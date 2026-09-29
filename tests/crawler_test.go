package tests

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"web-crawler/internal/crawler"
	"web-crawler/internal/document"
)

type memoryStore struct {
	docs []*document.Document
	err  error
}

func (s *memoryStore) Save(_ context.Context, doc *document.Document) error {
	if s.err != nil {
		return s.err
	}
	s.docs = append(s.docs, doc)
	return nil
}

func TestCrawlPipeline(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.URL.Path)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		switch r.URL.Path {
		case "/":
			io.WriteString(w, `<title>Home</title><p>Hello</p><a href="/bad">Bad</a><a href="/next#one">Next</a><a href="/next#two">Again</a><a href="mailto:x@example.com">Mail</a><a href="https://other.invalid/">External</a><a href="/image">Image</a>`)
		case "/bad":
			w.WriteHeader(http.StatusNotFound)
		case "/next":
			io.WriteString(w, `<title>Next</title><p>World</p><a href="/">Home</a>`)
		case "/image":
			w.Header().Set("Content-Type", "image/png")
			w.Write([]byte("image"))
		default:
			t.Errorf("unexpected request: %s", r.URL)
		}
	}))
	defer server.Close()
	start, _ := url.Parse(server.URL)
	store := &memoryStore{}
	if err := crawler.Crawl(context.Background(), crawler.Config{Start: *start, MaxPages: 10, Timeout: time.Second}, store); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(requests, []string{"/", "/bad", "/next", "/image"}) {
		t.Fatalf("requests = %v", requests)
	}
	if len(store.docs) != 2 {
		t.Fatalf("saved %d documents", len(store.docs))
	}
	for _, doc := range store.docs {
		if len(doc.ContentHash) != 64 || doc.CrawledAt.IsZero() || doc.Text == "" || doc.Title == "" {
			t.Fatalf("incomplete document: %+v", doc)
		}
	}
}

func TestCrawlStopsOnStorageFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, `<a href="/next">Next</a>`)
	}))
	defer server.Close()
	start, _ := url.Parse(server.URL)
	failure := errors.New("database unavailable")
	err := crawler.Crawl(context.Background(), crawler.Config{Start: *start, MaxPages: 10, Timeout: time.Second}, &memoryStore{err: failure})
	if !errors.Is(err, failure) {
		t.Fatalf("error = %v", err)
	}
}

func TestCrawlRejectsExternalRedirect(t *testing.T) {
	var externalRequests int
	external := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { externalRequests++ }))
	defer external.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, external.URL, http.StatusFound) }))
	defer server.Close()
	start, _ := url.Parse(server.URL)
	if err := crawler.Crawl(context.Background(), crawler.Config{Start: *start, MaxPages: 1, Timeout: time.Second}, &memoryStore{}); err == nil {
		t.Fatal("expected failure when nothing was saved")
	}
	if externalRequests != 0 {
		t.Fatal("followed external redirect")
	}
}

func TestParseConfig(t *testing.T) {
	for _, args := range [][]string{nil, {"example.com"}, {"ftp://example.com"}, {"-max-pages", "0", "https://example.com"}, {"-timeout", "0s", "https://example.com"}, {"-delay", "-1s", "https://example.com"}} {
		if _, err := crawler.ParseConfig(args, io.Discard); err == nil {
			t.Errorf("accepted %v", args)
		}
	}
	cfg, err := crawler.ParseConfig([]string{"https://EXAMPLE.com:443#section"}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Start.String() != "https://example.com/" {
		t.Fatalf("start = %s", cfg.Start.String())
	}
}
