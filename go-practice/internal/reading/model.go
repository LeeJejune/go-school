package reading

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound           = errors.New("reading list or item not found")
	ErrConflict           = errors.New("request key already belongs to a different payload")
	ErrInvalidInput       = errors.New("invalid input")
	ErrCatalogUnavailable = errors.New("catalog is unavailable; retry with the same request_key")
	ErrResourceNotFound   = errors.New("resource not found")
)

type Resource struct {
	ID, Title, URL string
}

type Item struct {
	Resource Resource
	IsRead   bool
}

type List struct {
	ID, Title string
	Items     []Item
	CreatedAt time.Time
}

type CreateParams struct {
	RequestKey, Fingerprint, Title string
	Resources                      []Resource
}

type CreateInput struct {
	RequestKey, Title string
	ResourceIDs       []string
}

// Repository lives with its consumer. It describes the three operations this service
// needs; it is not a generic repository for every table.
type Repository interface {
	Create(context.Context, CreateParams) (List, error)
	Get(context.Context, string) (List, error)
	SetRead(context.Context, string, string, bool) error
}

type Catalog interface {
	GetResources(context.Context, []string) ([]Resource, error)
}
