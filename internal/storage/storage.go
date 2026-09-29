package storage

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	d "web-crawler/internal/document"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schema string

type Storage struct{ db *pgxpool.Pool }

func NewStorage(db *pgxpool.Pool) *Storage { return &Storage{db: db} }

func (s *Storage) Init(ctx context.Context) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin schema initialization: %w", err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, schema); err != nil {
		return fmt.Errorf("initialize documents schema: %w", err)
	}
	return tx.Commit(ctx)
}

func (s *Storage) Save(ctx context.Context, document *d.Document) error {
	if document == nil {
		return fmt.Errorf("cannot save a nil document")
	}
	links := make([]string, 0, len(document.Links))
	for _, link := range document.Links {
		links = append(links, link.String())
	}
	linksJSON, err := json.Marshal(links)
	if err != nil {
		return err
	}
	metadataJSON, err := json.Marshal(document.Metadata)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(ctx, `
		INSERT INTO documents (url, title, text, content_type, links, metadata, content_hash, crawled_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (url) DO UPDATE SET
		    title = EXCLUDED.title, text = EXCLUDED.text,
		    content_type = EXCLUDED.content_type, links = EXCLUDED.links,
		    metadata = EXCLUDED.metadata, content_hash = EXCLUDED.content_hash,
		    crawled_at = EXCLUDED.crawled_at`,
		document.URL.String(), document.Title, document.Text, document.ContentType,
		linksJSON, metadataJSON, document.ContentHash, document.CrawledAt)
	if err != nil {
		return fmt.Errorf("save document %s: %w", document.URL.String(), err)
	}
	return nil
}
