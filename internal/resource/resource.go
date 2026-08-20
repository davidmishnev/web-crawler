package resource

import (
	"net/http"
	"net/url"
)

type Resource struct {
	URL         url.URL
	StatusCode  int
	ContentType string
	Headers     http.Header
	Data        []byte
}
