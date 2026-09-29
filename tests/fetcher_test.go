package tests

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"web-crawler/internal/fetcher"
)

func TestFetchRedirectStatusAndSize(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/redirect":
			http.Redirect(w, r, "/final", http.StatusFound)
		case "/final":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			io.WriteString(w, "<p>Hello</p>")
		case "/large":
			io.Copy(w, io.LimitReader(strings.NewReader(strings.Repeat("a", fetcher.MaxBodySize+1)), fetcher.MaxBodySize+1))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	loader := fetcher.NewFetcher(server.Client())
	u, _ := url.Parse(server.URL + "/redirect")
	res, err := loader.FetchResource(context.Background(), *u)
	if err != nil {
		t.Fatal(err)
	}
	if res.URL.Path != "/final" || res.ContentType != "text/html" || string(res.Data) != "<p>Hello</p>" {
		t.Fatalf("resource=%+v", res)
	}
	for _, path := range []string{"/missing", "/large"} {
		u, _ := url.Parse(server.URL + path)
		if _, err := loader.FetchResource(context.Background(), *u); err == nil {
			t.Errorf("accepted %s", path)
		}
	}
}
