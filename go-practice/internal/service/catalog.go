package service

import (
	"context"
	"errors"

	pb "github.com/LeeJejune/go-school/go-practice/gen/school/v1"
	"github.com/LeeJejune/go-school/go-practice/internal/catalog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Catalog struct {
	pb.UnimplementedCatalogServiceServer
	catalog *catalog.Catalog
}

func NewCatalog(catalog *catalog.Catalog) *Catalog { return &Catalog{catalog: catalog} }

func (s *Catalog) ListResources(context.Context, *pb.ListResourcesRequest) (*pb.ListResourcesResponse, error) {
	return &pb.ListResourcesResponse{Resources: resourcesToProto(s.catalog.List())}, nil
}

func (s *Catalog) BatchGetResources(_ context.Context, req *pb.BatchGetResourcesRequest) (*pb.BatchGetResourcesResponse, error) {
	resources, err := s.catalog.Batch(req.GetIds())
	if errors.Is(err, catalog.ErrInvalidInput) {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &pb.BatchGetResourcesResponse{Resources: resourcesToProto(resources)}, nil
}

func resourcesToProto(resources []catalog.Resource) []*pb.Resource {
	out := make([]*pb.Resource, 0, len(resources))
	for _, r := range resources {
		out = append(out, &pb.Resource{Id: r.ID, Title: r.Title, Url: r.URL})
	}
	return out
}
