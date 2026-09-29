package tests

import (
	"context"
	"encoding/json"
	"net/url"
	"os"
	"testing"
	"time"

	"web-crawler/internal/document"
	"web-crawler/internal/storage"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Use only a disposable database: the test creates the documents table.
func TestStorageIntegration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to a disposable PostgreSQL database")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	store := storage.NewStorage(pool)
	// Exercise upgrading the original schema as well as repeated initialization.
	if _, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS documents (url TEXT, title TEXT, text TEXT)`); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := store.Init(ctx); err != nil {
			t.Fatal(err)
		}
	}
	u, _ := url.Parse("https://storage-test.invalid/" + time.Now().Format("150405.000000000"))
	doc := &document.Document{URL: *u, Title: "First", Text: "Привет", ContentType: "text/html", Links: []url.URL{*u}, Metadata: document.Metadata{Author: "Alice"}, ContentHash: "abc", CrawledAt: time.Now().UTC().Truncate(time.Microsecond)}
	defer pool.Exec(ctx, "DELETE FROM documents WHERE url = $1", u.String())
	if err := store.Save(ctx, doc); err != nil {
		t.Fatal(err)
	}
	doc.Title = "Updated"
	if err := store.Save(ctx, doc); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM documents WHERE url = $1", u.String()).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("count=%d", count)
	}
	var title, text, contentType, hash string
	var links, metadata []byte
	var crawledAt time.Time
	if err := pool.QueryRow(ctx, "SELECT title, text, content_type, links, metadata, content_hash, crawled_at FROM documents WHERE url = $1", u.String()).Scan(&title, &text, &contentType, &links, &metadata, &hash, &crawledAt); err != nil {
		t.Fatal(err)
	}
	var storedLinks []string
	var storedMetadata document.Metadata
	if err := json.Unmarshal(links, &storedLinks); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(metadata, &storedMetadata); err != nil {
		t.Fatal(err)
	}
	if title != doc.Title || text != doc.Text || contentType != doc.ContentType || hash != doc.ContentHash || !crawledAt.Equal(doc.CrawledAt) || len(storedLinks) != 1 || storedLinks[0] != u.String() || storedMetadata.Author != "Alice" {
		t.Fatal("stored document does not match input")
	}
}
