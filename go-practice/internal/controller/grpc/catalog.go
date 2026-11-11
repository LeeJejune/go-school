package grpc

import (
	"context"
	"errors"
	"log/slog"

	pb "github.com/LeeJejune/go-school/go-practice/gen/school/v1"
	"github.com/LeeJejune/go-school/go-practice/internal/catalog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Catalog struct {
	pb.UnimplementedCatalogServiceServer
	catalog *catalog.Catalog
	logger  *slog.Logger
}

func NewCatalog(catalog *catalog.Catalog, logger *slog.Logger) *Catalog {
	return &Catalog{catalog: catalog, logger: logger}
}

func (s *Catalog) ListResources(ctx context.Context, _ *pb.ListResourcesRequest) (*pb.ListResourcesResponse, error) {
	resources, err := s.catalog.List(ctx)
	if err != nil {
		return nil, s.rpcError(ctx, err)
	}
	return &pb.ListResourcesResponse{Resources: resourcesToProto(resources)}, nil
}
func (s *Catalog) BatchGetResources(ctx context.Context, req *pb.BatchGetResourcesRequest) (*pb.BatchGetResourcesResponse, error) {
	resources, err := s.catalog.Batch(ctx, req.GetIds())
	if err != nil {
		return nil, s.rpcError(ctx, err)
	}
	return &pb.BatchGetResourcesResponse{Resources: resourcesToProto(resources)}, nil
}
func (s *Catalog) UpsertResource(ctx context.Context, req *pb.UpsertResourceRequest) (*pb.UpsertResourceResponse, error) {
	resource, err := s.catalog.Upsert(ctx, catalog.Resource{ID: req.GetId(), Title: req.GetTitle(), URL: req.GetUrl()})
	if err != nil {
		return nil, s.rpcError(ctx, err)
	}
	return &pb.UpsertResourceResponse{Resource: resourceToProto(resource)}, nil
}
func (s *Catalog) rpcError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, catalog.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, catalog.ErrNotFound):
		return status.Error(codes.NotFound, catalog.ErrNotFound.Error())
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return status.FromContextError(err).Err()
	case ctx.Err() != nil:
		return status.FromContextError(ctx.Err()).Err()
	default:
		s.logger.ErrorContext(ctx, "catalog operation failed", "error", err)
		return status.Error(codes.Internal, "could not access resource catalog")
	}
}
func resourcesToProto(resources []catalog.Resource) []*pb.Resource {
	out := make([]*pb.Resource, 0, len(resources))
	for _, resource := range resources {
		out = append(out, resourceToProto(resource))
	}
	return out
}

func resourceToProto(resource catalog.Resource) *pb.Resource {
	return &pb.Resource{Id: resource.ID, Title: resource.Title, Url: resource.URL}
}
