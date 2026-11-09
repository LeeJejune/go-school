package catalog

import "errors"

var ErrInvalidInput = errors.New("ids must contain 1 to 20 resources")
var ErrNotFound = errors.New("resource not found")

type Resource struct{ ID, Title, URL string }

type Catalog struct{ resources []Resource }

func New() *Catalog {
	return &Catalog{resources: []Resource{
		{ID: "go-tour", Title: "A Tour of Go", URL: "https://go.dev/tour/"},
		{ID: "go-context", Title: "Go Concurrency Patterns: Context", URL: "https://go.dev/blog/context"},
		{ID: "go-sql", Title: "Accessing relational databases", URL: "https://go.dev/doc/database/"},
		{ID: "grpc-go", Title: "gRPC Go basics", URL: "https://grpc.io/docs/languages/go/basics/"},
	}}
}

func (c *Catalog) List() []Resource { return append([]Resource(nil), c.resources...) }

func (c *Catalog) Batch(ids []string) ([]Resource, error) {
	if len(ids) == 0 || len(ids) > 20 {
		return nil, ErrInvalidInput
	}
	resources := make(map[string]Resource, len(c.resources))
	for _, r := range c.resources {
		resources[r.ID] = r
	}
	out := make([]Resource, 0, len(ids))
	for _, id := range ids {
		r, ok := resources[id]
		if !ok {
			return nil, ErrNotFound
		}
		out = append(out, r)
	}
	return out, nil
}
