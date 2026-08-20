package html

import (
	r "web-crawler/internal/resource"
)

type HTMLParser struct{}

func (p HTMLParser) CanParse(r r.Resource) bool {
	return r.ContentType == "PDF"
}

//func (p HTMLParser) Parse(ctx context.Context, r r.Resource) (*d.Document, error) {}
