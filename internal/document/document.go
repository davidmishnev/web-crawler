package document

import (
	"net/url"
	"time"
)

type Document struct {
	URL         url.URL
	ContentType Content
	Title       string
	Text        string
	Links       []url.URL
	Metadata    Metadata
	ContentHash string
	CrawledAt   time.Time
}
