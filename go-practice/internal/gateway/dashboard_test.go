package gateway

import (
	"context"
	"testing"
	"time"

	pb "github.com/LeeJejune/go-school/go-practice/gen/school/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type catalogStub struct {
	pb.CatalogServiceClient
	call func(context.Context) (*pb.ListResourcesResponse, error)
}

func (s catalogStub) ListResources(ctx context.Context, _ *pb.ListResourcesRequest, _ ...grpc.CallOption) (*pb.ListResourcesResponse, error) {
	return s.call(ctx)
}

type readingStub struct {
	pb.ReadingServiceClient
	call func(context.Context) (*pb.GetListResponse, error)
}

func (s readingStub) GetList(ctx context.Context, _ *pb.GetListRequest, _ ...grpc.CallOption) (*pb.GetListResponse, error) {
	return s.call(ctx)
}

func TestDashboardRunsRPCsConcurrently(t *testing.T) {
	catalogStarted, readingStarted := make(chan struct{}), make(chan struct{})
	d := NewDashboard(&Clients{
		Catalog: catalogStub{call: func(ctx context.Context) (*pb.ListResourcesResponse, error) {
			close(catalogStarted)
			select {
			case <-readingStarted:
				return &pb.ListResourcesResponse{}, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}},
		Reading: readingStub{call: func(ctx context.Context) (*pb.GetListResponse, error) {
			close(readingStarted)
			select {
			case <-catalogStarted:
				return &pb.GetListResponse{ReadingList: &pb.ReadingList{Id: "saved"}}, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}},
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	got, err := d.Get(ctx, "saved")
	if err != nil || got.ReadingList.Id != "saved" {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestDashboardFailureCancelsAndJoinsOtherRPC(t *testing.T) {
	started, canceled := make(chan struct{}), make(chan struct{})
	d := NewDashboard(&Clients{
		Catalog: catalogStub{call: func(ctx context.Context) (*pb.ListResourcesResponse, error) {
			select {
			case <-started:
				return nil, status.Error(codes.Unavailable, "catalog down")
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}},
		Reading: readingStub{call: func(ctx context.Context) (*pb.GetListResponse, error) {
			close(started)
			<-ctx.Done()
			close(canceled)
			return nil, ctx.Err()
		}},
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := d.Get(ctx, "saved"); status.Code(err) != codes.Unavailable {
		t.Fatalf("err=%v", err)
	}
	select {
	case <-canceled:
	default:
		t.Fatal("returned before the other RPC ended")
	}
}
