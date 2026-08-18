package resource

import (
	"net/http"
	"net/url"
	"web-crawler/internal/document"
)

type Resource struct {
	URL         url.URL
	StatusCode  int
	ContentType document.Content
	Headers     http.Header
	Data        []byte
}
