package reading

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// UseCase is application/business logic: it imports neither gRPC nor GORM.
type UseCase struct {
	repo          Repository
	catalog       Catalog
	lookupTimeout time.Duration
}

func New(repo Repository, catalog Catalog, lookupTimeout time.Duration) *UseCase {
	return &UseCase{repo: repo, catalog: catalog, lookupTimeout: lookupTimeout}
}

func (u *UseCase) Create(ctx context.Context, input CreateInput) (List, error) {
	title := strings.TrimSpace(input.Title)
	invalid := func(message string) (List, error) { return List{}, fmt.Errorf("%w: %s", ErrInvalidInput, message) }
	if title == "" || utf8.RuneCountInString(title) > 100 {
		return invalid("title must contain 1 to 100 characters")
	}
	if len(input.RequestKey) < 8 || len(input.RequestKey) > 64 {
		return invalid("request_key must contain 8 to 64 ASCII characters")
	}
	for _, ch := range input.RequestKey {
		if !(ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || ch == '-' || ch == '_') {
			return invalid("request_key accepts letters, numbers, hyphen and underscore")
		}
	}
	if len(input.ResourceIDs) == 0 || len(input.ResourceIDs) > 20 {
		return invalid("select 1 to 20 resources")
	}
	seen := make(map[string]bool)
	for _, id := range input.ResourceIDs {
		if id == "" || seen[id] {
			return invalid("resource IDs must be nonempty and unique")
		}
		seen[id] = true
	}
	// Finish the RPC before beginning the repository's DB transaction.
	lookupCtx, cancel := context.WithTimeout(ctx, u.lookupTimeout)
	defer cancel()
	resources, err := u.catalog.GetResources(lookupCtx, input.ResourceIDs)
	if err != nil {
		if errors.Is(err, ErrResourceNotFound) {
			return invalid("one or more resources do not exist")
		}
		if ctx.Err() != nil {
			return List{}, ctx.Err()
		}
		return List{}, fmt.Errorf("%w: %v", ErrCatalogUnavailable, err)
	}
	payload, _ := json.Marshal(struct {
		Title string
		IDs   []string
	}{title, input.ResourceIDs})
	sum := sha256.Sum256(payload)
	return u.repo.Create(ctx, CreateParams{
		RequestKey: input.RequestKey, Fingerprint: hex.EncodeToString(sum[:]), Title: title, Resources: resources,
	})
}

func (u *UseCase) Get(ctx context.Context, id string) (List, error) {
	if id == "" {
		return List{}, fmt.Errorf("%w: id is required", ErrInvalidInput)
	}
	return u.repo.Get(ctx, id)
}

func (u *UseCase) SetRead(ctx context.Context, listID, resourceID string, isRead bool) error {
	if listID == "" || resourceID == "" {
		return fmt.Errorf("%w: list_id and resource_id are required", ErrInvalidInput)
	}
	return u.repo.SetRead(ctx, listID, resourceID, isRead)
}
