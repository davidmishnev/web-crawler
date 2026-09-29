package html

import (
	"bytes"
	"context"
	"mime"
	"net/url"
	"strings"

	d "web-crawler/internal/document"
	"web-crawler/internal/parser"
	r "web-crawler/internal/resource"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

type HTMLParser struct{}

var _ parser.Parser = HTMLParser{}

func (p HTMLParser) CanParse(resource r.Resource) bool {
	mediaType, _, err := mime.ParseMediaType(resource.ContentType)
	return err == nil && (mediaType == "text/html" || mediaType == "application/xhtml+xml")
}

func (p HTMLParser) Parse(ctx context.Context, resource r.Resource) (*d.Document, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	contentType := resource.Headers.Get("Content-Type")
	if contentType == "" {
		contentType = resource.ContentType
	}
	input, err := charset.NewReader(bytes.NewReader(resource.Data), contentType)
	if err != nil {
		return nil, err
	}
	root, err := html.Parse(input)
	if err != nil {
		return nil, err
	}
	doc := &d.Document{URL: resource.URL, ContentType: resource.ContentType, Metadata: d.Metadata{Extra: make(map[string]string)}}
	base := resource.URL
	var hrefs []string
	var text, title strings.Builder
	baseSet := false
	var walk func(*html.Node, bool, bool) error
	walk = func(node *html.Node, inHead, inTitle bool) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if node.Type == html.ElementNode {
			switch node.Data {
			case "script", "style", "noscript", "template":
				return nil
			case "head":
				inHead = true
			case "title":
				inTitle = true
			case "html":
				doc.Metadata.Language = attribute(node, "lang")
			case "meta":
				name, content := strings.ToLower(attribute(node, "name")), attribute(node, "content")
				switch name {
				case "author":
					doc.Metadata.Author = content
				case "description":
					doc.Metadata.Description = content
				case "":
					if property := attribute(node, "property"); property != "" {
						doc.Metadata.Extra[property] = content
					}
				default:
					doc.Metadata.Extra[name] = content
				}
			case "base":
				if !baseSet {
					if href := attribute(node, "href"); href != "" {
						if ref, err := url.Parse(href); err == nil {
							if normalized, err := r.NormalizeURL(*resource.URL.ResolveReference(ref)); err == nil {
								base = normalized
								baseSet = true
							}
						}
					}
				}
			case "a", "area":
				if href := strings.TrimSpace(attribute(node, "href")); href != "" {
					hrefs = append(hrefs, href)
				}
			}
			if isBlock(node.Data) {
				text.WriteByte(' ')
			}
		}
		if node.Type == html.TextNode {
			if inTitle {
				title.WriteString(node.Data)
			} else if !inHead {
				text.WriteString(node.Data)
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if err := walk(child, inHead, inTitle); err != nil {
				return err
			}
		}
		if node.Type == html.ElementNode && isBlock(node.Data) {
			text.WriteByte(' ')
		}
		return nil
	}
	if err := walk(root, false, false); err != nil {
		return nil, err
	}
	doc.Title = strings.Join(strings.Fields(title.String()), " ")
	doc.Text = strings.Join(strings.Fields(text.String()), " ")
	seen := make(map[string]struct{})
	for _, href := range hrefs {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		ref, err := url.Parse(href)
		if err != nil {
			continue
		}
		link, err := r.NormalizeURL(*base.ResolveReference(ref))
		if err != nil {
			continue
		}
		if _, exists := seen[link.String()]; exists {
			continue
		}
		seen[link.String()] = struct{}{}
		doc.Links = append(doc.Links, link)
	}
	return doc, nil
}

func attribute(node *html.Node, name string) string {
	for _, attr := range node.Attr {
		if attr.Key == name {
			return attr.Val
		}
	}
	return ""
}

func isBlock(tag string) bool {
	switch tag {
	case "address", "article", "aside", "blockquote", "br", "dd", "div", "dl", "dt", "figcaption", "figure", "footer", "form", "h1", "h2", "h3", "h4", "h5", "h6", "header", "hr", "li", "main", "nav", "ol", "p", "pre", "section", "table", "td", "th", "tr", "ul":
		return true
	}
	return false
}
