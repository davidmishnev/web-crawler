package frontier

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"web-crawler/internal/document"
	"web-crawler/internal/resource"
)

type Frontier struct {
	// MaxPages bounds both requests and queue size. Zero means 100.
	MaxPages int
	Delay    time.Duration
}

// Crawl visits URLs in breadth-first order within the starting URL's host.
// A nil document skips a URL; an error stops the crawl.
func (f *Frontier) Crawl(ctx context.Context, start url.URL, fetch func(*url.URL) (*document.Document, error)) error {
	start, err := resource.NormalizeURL(start)
	if err != nil {
		return err
	}
	limit := f.MaxPages
	if limit <= 0 {
		limit = 100
	}
	queue := []url.URL{start}
	visited := map[string]struct{}{start.String(): {}}
	fetched := make(map[string]struct{})
	for index := 0; index < len(queue); index++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		current := queue[index]
		if _, exists := fetched[current.String()]; exists {
			continue
		}
		if index > 0 && f.Delay > 0 {
			timer := time.NewTimer(f.Delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}
		fetched[current.String()] = struct{}{}
		doc, err := fetch(&current)
		if err != nil {
			return fmt.Errorf("crawl %s: %w", current.String(), err)
		}
		if doc == nil {
			continue
		}
		// A redirect destination may already be waiting in the queue.
		if final, err := resource.NormalizeURL(doc.URL); err == nil {
			visited[final.String()] = struct{}{}
			fetched[final.String()] = struct{}{}
		}
		for _, link := range doc.Links {
			if len(queue) >= limit {
				break
			}
			resolved := doc.URL.ResolveReference(&link)
			normalized, err := resource.NormalizeURL(*resolved)
			if err != nil || normalized.Host != start.Host {
				continue
			}
			key := normalized.String()
			if _, exists := visited[key]; exists {
				continue
			}
			visited[key] = struct{}{}
			queue = append(queue, normalized)
		}
	}
	return ctx.Err()
}
