package tests

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"

	htmlparser "web-crawler/internal/parser/html"
	"web-crawler/internal/resource"
)

func TestParseDocument(t *testing.T) {
	base, _ := url.Parse("https://example.com/old/page")
	res := resource.Resource{URL: *base, ContentType: "text/html; charset=utf-8", Data: []byte(`<html lang="ru"><head><title> A &amp; B </title><base href="/docs/"><meta name="author" content="Alice"><meta name="description" content="Summary"></head><body><p>Hello <b>world</b>!</p><p>Next</p><script>secret()</script><a href="a#x">One</a><a href="a#y">Two</a><a href="mailto:a@b.com">Mail</a></body></html>`)}
	parser := htmlparser.HTMLParser{}
	if !parser.CanParse(res) {
		t.Fatal("did not accept HTML")
	}
	doc, err := parser.Parse(context.Background(), res)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Title != "A & B" || doc.Text != "Hello world! Next OneTwoMail" {
		t.Fatalf("title=%q text=%q", doc.Title, doc.Text)
	}
	if len(doc.Links) != 1 || doc.Links[0].String() != "https://example.com/docs/a" {
		t.Fatalf("links=%v", doc.Links)
	}
	if doc.Metadata.Author != "Alice" || doc.Metadata.Language != "ru" || doc.Metadata.Description != "Summary" {
		t.Fatalf("metadata=%+v", doc.Metadata)
	}
}

func TestParseCharsetAndCancellation(t *testing.T) {
	res := resource.Resource{ContentType: "text/html", Headers: http.Header{"Content-Type": {"text/html; charset=windows-1251"}}, Data: append([]byte("<p>"), []byte{0xcf, 0xf0, 0xe8, 0xe2, 0xe5, 0xf2}...)}
	doc, err := (htmlparser.HTMLParser{}).Parse(context.Background(), res)
	if err != nil || doc.Text != "Привет" {
		t.Fatalf("doc=%+v error=%v", doc, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := (htmlparser.HTMLParser{}).Parse(ctx, res); !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v", err)
	}
}
