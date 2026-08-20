package pdf

import (
	r "web-crawler/internal/resource"
)

type PDFParser struct{}

func (p PDFParser) CanParse(r r.Resource) bool {
	return r.ContentType == "PDF"
}

//func (p PDFParser) Parse(ctx context.Context, r r.Resource) (*d.Document, error) {}
