package tests

import (
	"context"
	"errors"
	"net/url"
	"reflect"
	"testing"

	"web-crawler/internal/document"
	"web-crawler/internal/frontier"
)

func TestCrawlLimitAndDeduplication(t *testing.T) {
	start, _ := url.Parse("https://example.com/")
	var got []string
	f := frontier.Frontier{MaxPages: 3}
	err := f.Crawl(context.Background(), *start, func(u *url.URL) (*document.Document, error) {
		got = append(got, u.Path)
		doc := &document.Document{URL: *u}
		for _, raw := range []string{"/a#one", "/a#two", "mailto:x@y.com", "https://other.com/", "/b", "/c"} {
			link, _ := url.Parse(raw)
			doc.Links = append(doc.Links, *link)
		}
		return doc, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []string{"/", "/a", "/b"}) {
		t.Fatalf("visited=%v", got)
	}
}

func TestCrawlCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start, _ := url.Parse("https://example.com/")
	err := (&frontier.Frontier{}).Crawl(ctx, *start, func(*url.URL) (*document.Document, error) { t.Fatal("fetch after cancellation"); return nil, nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v", err)
	}
}
