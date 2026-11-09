package reading

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("reading list or item not found")
	ErrConflict = errors.New("request key already belongs to a different payload")
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

// Store lives with its consumer. It describes the three operations this service
// needs; it is not a generic repository for every table.
type Store interface {
	Create(context.Context, CreateParams) (List, error)
	Get(context.Context, string) (List, error)
	SetRead(context.Context, string, string, bool) error
}

type Catalog interface {
	GetResources(context.Context, []string) ([]Resource, error)
}
