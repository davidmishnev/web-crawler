package tests

import (
	"context"
	"fmt"
	"strings"
	"testing"

	pdfparser "web-crawler/internal/parser/pdf"
	"web-crawler/internal/resource"
)

func TestParsePDF(t *testing.T) {
	stream := "BT /F1 12 Tf 72 720 Td (Hello PDF) Tj ET"
	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(stream), stream),
		"<< /Title (Sample) /Author (Alice) >>",
	}
	var data strings.Builder
	data.WriteString("%PDF-1.4\n")
	var offsets []int
	for i, object := range objects {
		offsets = append(offsets, data.Len())
		fmt.Fprintf(&data, "%d 0 obj\n%s\nendobj\n", i+1, object)
	}
	xref := data.Len()
	fmt.Fprintf(&data, "xref\n0 %d\n0000000000 65535 f \n", len(objects)+1)
	for _, offset := range offsets {
		fmt.Fprintf(&data, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&data, "trailer\n<< /Size %d /Root 1 0 R /Info 6 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xref)
	doc, err := (pdfparser.PDFParser{}).Parse(context.Background(), resource.Resource{ContentType: "application/pdf", Data: []byte(data.String())})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(doc.Text, "Hello PDF") || doc.Title != "Sample" || doc.Metadata.Author != "Alice" {
		t.Fatalf("document=%+v", doc)
	}
}

func TestRejectInvalidPDF(t *testing.T) {
	if _, err := (pdfparser.PDFParser{}).Parse(context.Background(), resource.Resource{Data: []byte("not a PDF")}); err == nil {
		t.Fatal("accepted invalid PDF")
	}
}
