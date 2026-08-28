package frontier

import (
	"net/url"

	"web-crawler/internal/document"
)

type Frontier struct {
	queue   []*url.URL
	visited map[string]struct{}
}

func (f *Frontier) Crawl(start url.URL, fetch func(*url.URL) (*document.Document, error)) error {
	f.queue = []*url.URL{&start}
	f.visited = make(map[string]struct{})

	f.visited[start.String()] = struct{}{}

	for len(f.queue) > 0 {
		current := f.queue[0]
		f.queue = f.queue[1:]

		doc, err := fetch(current)
		if err != nil {
			return err
		}

		for _, link := range doc.Links {
			key := link.String()

			if _, exists := f.visited[key]; exists {
				continue
			}

			f.visited[key] = struct{}{}
			f.queue = append(f.queue, &link)
		}
	}

	return nil
}
