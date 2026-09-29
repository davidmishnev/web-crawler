package pdf

import (
	"bytes"
	"context"
	"fmt"
	"mime"
	"strings"

	d "web-crawler/internal/document"
	"web-crawler/internal/parser"
	r "web-crawler/internal/resource"

	"github.com/ledongthuc/pdf"
)

type PDFParser struct{}

var _ parser.Parser = PDFParser{}

func (p PDFParser) CanParse(resource r.Resource) bool {
	contentType, _, err := mime.ParseMediaType(resource.ContentType)
	return err == nil && contentType == "application/pdf"
}

func (p PDFParser) Parse(ctx context.Context, resource r.Resource) (*d.Document, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	reader, err := pdf.NewReader(
		bytes.NewReader(resource.Data),
		int64(len(resource.Data)),
	)
	if err != nil {
		return nil, fmt.Errorf("create PDF reader: %w", err)
	}

	text, err := extractText(ctx, reader)
	if err != nil {
		return nil, fmt.Errorf("extract PDF text: %w", err)
	}

	return &d.Document{
		URL:         resource.URL,
		ContentType: resource.ContentType,
		Title:       reader.Trailer().Key("Info").Key("Title").Text(),
		Metadata:    d.Metadata{Author: reader.Trailer().Key("Info").Key("Author").Text()},
		Text:        text,
	}, nil
}

func extractText(ctx context.Context, reader *pdf.Reader) (string, error) {
	var builder strings.Builder

	for pageNum := 1; pageNum <= reader.NumPage(); pageNum++ {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		page := reader.Page(pageNum)

		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			return "", err
		}

		builder.WriteString(text)
		builder.WriteString("\n")
	}

	return builder.String(), nil
}
