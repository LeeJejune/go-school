package reading

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	pb "github.com/LeeJejune/go-school/gen/school/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	pb.UnimplementedReadingServiceServer
	store   Store
	catalog Catalog
}

func New(store Store, catalog Catalog) *Service { return &Service{store: store, catalog: catalog} }

func (s *Service) CreateList(ctx context.Context, req *pb.CreateListRequest) (*pb.CreateListResponse, error) {
	title := strings.TrimSpace(req.GetTitle())
	if title == "" || utf8.RuneCountInString(title) > 100 {
		return nil, status.Error(codes.InvalidArgument, "title must contain 1 to 100 characters")
	}
	if len(req.GetRequestKey()) < 8 || len(req.GetRequestKey()) > 64 {
		return nil, status.Error(codes.InvalidArgument, "request_key must contain 8 to 64 ASCII characters")
	}
	for _, ch := range req.RequestKey {
		if !(ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || ch == '-' || ch == '_') {
			return nil, status.Error(codes.InvalidArgument, "request_key accepts letters, numbers, hyphen and underscore")
		}
	}
	if len(req.GetResourceIds()) == 0 || len(req.GetResourceIds()) > 20 {
		return nil, status.Error(codes.InvalidArgument, "select 1 to 20 resources")
	}
	seen := make(map[string]bool)
	for _, id := range req.ResourceIds {
		if id == "" || seen[id] {
			return nil, status.Error(codes.InvalidArgument, "resource IDs must be nonempty and unique")
		}
		seen[id] = true
	}

	// Network calls happen before the SQL transaction: no DB connection is held
	// while waiting for another server. Catalog entries are snapshotted on save.
	lookupCtx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()
	resources, err := s.catalog.GetResources(lookupCtx, req.ResourceIds)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, status.Error(codes.InvalidArgument, "one or more resources do not exist")
		}
		if ctx.Err() != nil {
			return nil, status.FromContextError(ctx.Err()).Err()
		}
		return nil, status.Error(codes.Unavailable, "catalog is unavailable; retry with the same request_key")
	}
	// Fingerprint the normalized request, not changing catalog titles.
	payload, _ := json.Marshal(struct {
		Title string
		IDs   []string
	}{title, req.ResourceIds})
	sum := sha256.Sum256(payload)
	list, err := s.store.Create(ctx, CreateParams{
		RequestKey: req.RequestKey, Fingerprint: hex.EncodeToString(sum[:]), Title: title, Resources: resources,
	})
	if err != nil {
		return nil, rpcError(ctx, err)
	}
	return &pb.CreateListResponse{ReadingList: toProto(list)}, nil
}

func (s *Service) GetList(ctx context.Context, req *pb.GetListRequest) (*pb.GetListResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	list, err := s.store.Get(ctx, req.Id)
	if err != nil {
		return nil, rpcError(ctx, err)
	}
	return &pb.GetListResponse{ReadingList: toProto(list)}, nil
}

func (s *Service) SetRead(ctx context.Context, req *pb.SetReadRequest) (*pb.SetReadResponse, error) {
	if req.GetListId() == "" || req.GetResourceId() == "" {
		return nil, status.Error(codes.InvalidArgument, "list_id and resource_id are required")
	}
	if err := s.store.SetRead(ctx, req.ListId, req.ResourceId, req.IsRead); err != nil {
		return nil, rpcError(ctx, err)
	}
	return &pb.SetReadResponse{}, nil
}

func rpcError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, ErrConflict):
		return status.Error(codes.AlreadyExists, err.Error())
	case ctx.Err() != nil:
		return status.FromContextError(ctx.Err()).Err()
	default:
		slog.Error("reading store failed", "error", err)
		return status.Error(codes.Internal, "could not persist reading list")
	}
}

func toProto(list List) *pb.ReadingList {
	out := &pb.ReadingList{Id: list.ID, Title: list.Title, CreatedAt: list.CreatedAt.UTC().Format(time.RFC3339Nano)}
	for _, item := range list.Items {
		out.Items = append(out.Items, &pb.ReadingItem{
			Resource: &pb.Resource{Id: item.Resource.ID, Title: item.Resource.Title, Url: item.Resource.URL}, IsRead: item.IsRead,
		})
	}
	return out
}
