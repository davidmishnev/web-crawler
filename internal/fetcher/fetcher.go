package fetcher

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"time"
	r "web-crawler/internal/resource"
)

type Fetcher struct {
	client *http.Client
}

const MaxBodySize = 20 << 20 // 20 MiB, including decompressed responses.

func NewFetcher(client *http.Client) *Fetcher {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	return &Fetcher{client: client}
}

func (f Fetcher) FetchResource(ctx context.Context, target url.URL) (*r.Resource, error) {
	target, err := r.NormalizeURL(target)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "web-crawler/1.0")
	req.Header.Set("Accept", "text/html, application/xhtml+xml, application/pdf")
	client := f.client
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("GET %s: %s", target.String(), resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, MaxBodySize+1))
	if err != nil {
		return nil, err
	}

	if len(data) > MaxBodySize {
		return nil, fmt.Errorf("response exceeds %d bytes", MaxBodySize)
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}
	if mediaType, _, err := mime.ParseMediaType(contentType); err == nil {
		contentType = mediaType
	}
	return &r.Resource{
		URL:         *resp.Request.URL,
		StatusCode:  resp.StatusCode,
		ContentType: contentType,
		Headers:     resp.Header,
		Data:        data,
	}, nil
}
