package parser

import (
	"context"
	d "web-crawler/internal/document"
	r "web-crawler/internal/resource"
)

type Parser interface {
	CanParse(resource r.Resource) bool

	Parse(ctx context.Context, resource r.Resource) (*d.Document, error)
}
