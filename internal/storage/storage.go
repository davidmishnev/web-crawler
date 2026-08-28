package storage

import (
	"context"
	d "web-crawler/internal/document"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func NewStorage(db *pgxpool.Pool) *Storage {
	return &Storage{
		db: db,
	}
}

func (s *Storage) Save(ctx context.Context, document *d.Document) error {
	_, err := s.db.Exec(
		ctx,
		`
		INSERT INTO documents (url, title, text)
		VALUES ($1, $2, $3)
		`,
		document.URL.String(),
		document.Title,
		document.Text,
	)

	return err
}
