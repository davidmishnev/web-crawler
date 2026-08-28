package pdf

import (
	"bytes"
	"fmt"
	"strings"

	d "web-crawler/internal/document"
	r "web-crawler/internal/resource"

	"github.com/ledongthuc/pdf"
)

type PDFParser struct{}

func (p PDFParser) CanParse(resource r.Resource) bool {
	contentType := strings.ToLower(resource.ContentType)

	return strings.HasPrefix(contentType, "application/pdf")
}

func (p PDFParser) Parse(resource r.Resource) (*d.Document, error) {
	reader, err := pdf.NewReader(
		bytes.NewReader(resource.Data),
		int64(len(resource.Data)),
	)
	if err != nil {
		return nil, fmt.Errorf("create PDF reader: %w", err)
	}

	text, err := extractText(reader)
	if err != nil {
		return nil, fmt.Errorf("extract PDF text: %w", err)
	}

	return &d.Document{
		URL:  resource.URL,
		Text: text,
	}, nil
}

func extractText(reader *pdf.Reader) (string, error) {
	var builder strings.Builder

	for pageNum := 1; pageNum <= reader.NumPage(); pageNum++ {
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
