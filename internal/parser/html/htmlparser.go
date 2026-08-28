package html

import (
	"net/url"
	"strings"

	d "web-crawler/internal/document"
	r "web-crawler/internal/resource"

	"golang.org/x/net/html"
)

type HTMLParser struct{}

func (p HTMLParser) CanParse(resource r.Resource) bool {
	contentType := strings.ToLower(resource.ContentType)

	return strings.HasPrefix(contentType, "text/html")
}

func (p HTMLParser) Parse(resource r.Resource) (*d.Document, error) {
	root, err := html.Parse(strings.NewReader(string(resource.Data)))
	if err != nil {
		return nil, err
	}

	doc := &d.Document{
		URL: resource.URL,
	}

	var walk func(*html.Node)

	walk = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "a" {
			for _, attr := range node.Attr {
				if attr.Key == "href" {
					link, err := url.Parse(attr.Val)
					if err == nil {
						doc.Links = append(doc.Links, *link)
					}
				}
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}

	walk(root)

	return doc, nil
}
