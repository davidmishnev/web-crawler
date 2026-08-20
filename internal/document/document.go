package document

import (
	"net/url"
	"time"
)

type Document struct {
	URL         url.URL
	ContentType string
	Title       string
	Text        string
	Links       []url.URL
	Metadata    Metadata
	ContentHash string
	CrawledAt   time.Time
}
