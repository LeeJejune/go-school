// Package catalog owns resource validation and storage contracts.
package catalog

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"unicode/utf8"
)

var ErrInvalidInput = errors.New("invalid resource input")
var ErrNotFound = errors.New("resource not found")

type Resource struct{ ID, Title, URL string }

type Repository interface {
	List(context.Context) ([]Resource, error)
	Batch(context.Context, []string) ([]Resource, error)
	Upsert(context.Context, Resource) (Resource, error)
}

type Catalog struct{ repo Repository }

func New(repo Repository) *Catalog { return &Catalog{repo: repo} }

func (c *Catalog) List(ctx context.Context) ([]Resource, error) { return c.repo.List(ctx) }

func (c *Catalog) Batch(ctx context.Context, ids []string) ([]Resource, error) {
	if len(ids) == 0 || len(ids) > 20 {
		return nil, fmt.Errorf("%w: ids must contain 1 to 20 resources", ErrInvalidInput)
	}
	return c.repo.Batch(ctx, ids)
}

func (c *Catalog) Upsert(ctx context.Context, resource Resource) (Resource, error) {
	invalid := func(message string) (Resource, error) {
		return Resource{}, fmt.Errorf("%w: %s", ErrInvalidInput, message)
	}
	if len(resource.ID) == 0 || len(resource.ID) > 64 {
		return invalid("id must contain 1 to 64 ASCII characters")
	}
	for _, ch := range resource.ID {
		if !(ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || ch == '-' || ch == '_') {
			return invalid("id accepts letters, numbers, hyphen and underscore")
		}
	}
	resource.Title = strings.TrimSpace(resource.Title)
	if resource.Title == "" || utf8.RuneCountInString(resource.Title) > 255 {
		return invalid("title must contain 1 to 255 characters")
	}
	if utf8.RuneCountInString(resource.URL) > 2048 {
		return invalid("url is too long")
	}
	parsed, err := url.Parse(resource.URL)
	if err != nil || parsed.Hostname() == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return invalid("url must be an absolute HTTP or HTTPS URL")
	}
	return c.repo.Upsert(ctx, resource)
}
