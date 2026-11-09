package reading

import (
	"context"
	"errors"
	"testing"
	"time"

	pb "github.com/LeeJejune/go-school/gen/school/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type stubCatalog struct{ err error }

func (c stubCatalog) GetResources(ctx context.Context, ids []string) ([]Resource, error) {
	if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) > time.Second {
		return nil, errors.New("missing lookup deadline")
	}
	return []Resource{{ID: ids[0], Title: "Go", URL: "https://go.dev"}}, c.err
}

type stubStore struct {
	created *CreateParams
	err     error
}

func (s *stubStore) Create(_ context.Context, p CreateParams) (List, error) {
	s.created = &p
	return List{ID: "saved", Title: p.Title, CreatedAt: time.Unix(0, 0)}, s.err
}
func (*stubStore) Get(context.Context, string) (List, error)           { return List{}, ErrNotFound }
func (*stubStore) SetRead(context.Context, string, string, bool) error { return nil }

func TestCreateValidationAndNormalization(t *testing.T) {
	store := &stubStore{}
	service := New(store, stubCatalog{})
	request := &pb.CreateListRequest{RequestKey: "learn-001", Title: "  Go 공부  ", ResourceIds: []string{"go-tour"}}
	response, err := service.CreateList(context.Background(), request)
	if err != nil || response.GetReadingList().GetTitle() != "Go 공부" || len(store.created.Fingerprint) != 64 {
		t.Fatalf("response=%v err=%v", response, err)
	}
	request.ResourceIds = []string{"go-tour", "go-tour"}
	store.created = nil
	_, err = service.CreateList(context.Background(), request)
	if status.Code(err) != codes.InvalidArgument || store.created != nil {
		t.Fatalf("duplicate validation: %v", err)
	}
}

func TestCatalogFailureDoesNotWrite(t *testing.T) {
	store := &stubStore{}
	service := New(store, stubCatalog{err: status.Error(codes.Unavailable, "down")})
	_, err := service.CreateList(context.Background(), &pb.CreateListRequest{RequestKey: "learn-002", Title: "Go", ResourceIds: []string{"go-tour"}})
	if status.Code(err) != codes.Unavailable || store.created != nil {
		t.Fatalf("err=%v wrote=%v", err, store.created)
	}
}

func TestErrorMapping(t *testing.T) {
	for _, tc := range []struct {
		err  error
		code codes.Code
	}{{ErrConflict, codes.AlreadyExists}, {ErrNotFound, codes.NotFound}} {
		if code := status.Code(rpcError(context.Background(), tc.err)); code != tc.code {
			t.Fatalf("got %v want %v", code, tc.code)
		}
	}
}
