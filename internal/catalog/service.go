package catalog

import (
	"context"

	pb "github.com/LeeJejune/go-school/gen/school/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type Service struct {
	pb.UnimplementedCatalogServiceServer
	resources []*pb.Resource
}

func New() *Service {
	return &Service{resources: []*pb.Resource{
		{Id: "go-tour", Title: "A Tour of Go", Url: "https://go.dev/tour/"},
		{Id: "go-context", Title: "Go Concurrency Patterns: Context", Url: "https://go.dev/blog/context"},
		{Id: "go-sql", Title: "Accessing relational databases", Url: "https://go.dev/doc/database/"},
		{Id: "grpc-go", Title: "gRPC Go basics", Url: "https://grpc.io/docs/languages/go/basics/"},
	}}
}

func (s *Service) ListResources(context.Context, *pb.ListResourcesRequest) (*pb.ListResourcesResponse, error) {
	out := &pb.ListResourcesResponse{}
	for _, resource := range s.resources {
		out.Resources = append(out.Resources, proto.Clone(resource).(*pb.Resource))
	}
	return out, nil
}

func (s *Service) BatchGetResources(_ context.Context, req *pb.BatchGetResourcesRequest) (*pb.BatchGetResourcesResponse, error) {
	if len(req.GetIds()) == 0 || len(req.GetIds()) > 20 {
		return nil, status.Error(codes.InvalidArgument, "ids must contain 1 to 20 resources")
	}
	out := &pb.BatchGetResourcesResponse{}
	for _, id := range req.GetIds() {
		var found *pb.Resource
		for _, resource := range s.resources {
			if resource.Id == id {
				found = proto.Clone(resource).(*pb.Resource)
				break
			}
		}
		if found == nil {
			return nil, status.Errorf(codes.NotFound, "resource %q does not exist", id)
		}
		out.Resources = append(out.Resources, found)
	}
	return out, nil
}
