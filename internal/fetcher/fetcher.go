package fetcher

import (
	"context"
	"io"
	"net/http"
	"net/url"
	r "web-crawler/internal/resource"
)

type Fetcher struct {
	client *http.Client
}

func (f Fetcher) FetchResource(ctx context.Context, url url.URL) (*r.Resource, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &r.Resource{
		URL:         url,
		StatusCode:  resp.StatusCode,
		ContentType: resp.Header.Get("Content-Type"),
		Headers:     resp.Header,
		Data:        data,
	}, nil
}
