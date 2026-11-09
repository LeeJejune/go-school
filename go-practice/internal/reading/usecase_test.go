package reading

import (
	"context"
	"errors"
	"testing"
	"time"
)

type stubCatalog struct{ err error }

func (c stubCatalog) GetResources(ctx context.Context, ids []string) ([]Resource, error) {
	if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) > time.Second {
		return nil, errors.New("missing lookup deadline")
	}
	return []Resource{{ID: ids[0], Title: "Go", URL: "https://go.dev"}}, c.err
}

type stubRepository struct {
	created *CreateParams
	err     error
}

func (s *stubRepository) Create(_ context.Context, p CreateParams) (List, error) {
	s.created = &p
	return List{ID: "saved", Title: p.Title, CreatedAt: time.Unix(0, 0)}, s.err
}
func (*stubRepository) Get(context.Context, string) (List, error)           { return List{}, ErrNotFound }
func (*stubRepository) SetRead(context.Context, string, string, bool) error { return nil }

func TestCreateValidationAndNormalization(t *testing.T) {
	repo := &stubRepository{}
	u := New(repo, stubCatalog{}, 300*time.Millisecond)
	input := CreateInput{RequestKey: "learn-001", Title: "  Go 공부  ", ResourceIDs: []string{"go-tour"}}
	list, err := u.Create(context.Background(), input)
	if err != nil || list.Title != "Go 공부" || repo.created == nil || len(repo.created.Fingerprint) != 64 {
		t.Fatalf("list=%v err=%v", list, err)
	}
	for _, ids := range [][]string{nil, {"go-tour", "go-tour"}, {""}} {
		input.ResourceIDs = ids
		repo.created = nil
		if _, err := u.Create(context.Background(), input); !errors.Is(err, ErrInvalidInput) || repo.created != nil {
			t.Fatalf("ids=%v err=%v wrote=%v", ids, err, repo.created)
		}
	}
}

func TestCatalogFailureDoesNotWrite(t *testing.T) {
	for _, tc := range []struct{ cause, want error }{{errors.New("connection down"), ErrCatalogUnavailable}, {ErrResourceNotFound, ErrInvalidInput}} {
		repo := &stubRepository{}
		u := New(repo, stubCatalog{err: tc.cause}, 300*time.Millisecond)
		_, err := u.Create(context.Background(), CreateInput{RequestKey: "learn-002", Title: "Go", ResourceIDs: []string{"go-tour"}})
		if !errors.Is(err, tc.want) || repo.created != nil {
			t.Fatalf("err=%v wrote=%v", err, repo.created)
		}
	}
}

func TestFingerprintIncludesResourceOrder(t *testing.T) {
	repo := &stubRepository{}
	u := New(repo, stubCatalog{}, 300*time.Millisecond)
	input := CreateInput{RequestKey: "learn-003", Title: "Go", ResourceIDs: []string{"a", "b"}}
	if _, err := u.Create(context.Background(), input); err != nil {
		t.Fatal(err)
	}
	first := repo.created.Fingerprint
	input.ResourceIDs = []string{"b", "a"}
	if _, err := u.Create(context.Background(), input); err != nil {
		t.Fatal(err)
	}
	if first == repo.created.Fingerprint {
		t.Fatal("resource order must change the request fingerprint")
	}
}
