package rpc

import (
	"context"
	"io"
	"log/slog"
	"net"
	"strings"
	"testing"
	"time"

	pb "github.com/LeeJejune/go-school/go-practice/gen/school/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type testCatalog struct {
	pb.UnimplementedCatalogServiceServer
	run func(context.Context) (*pb.ListResourcesResponse, error)
}

func (s testCatalog) ListResources(ctx context.Context, _ *pb.ListResourcesRequest) (*pb.ListResourcesResponse, error) {
	return s.run(ctx)
}

func TestServerInterceptorsHandlePanicAndDeadline(t *testing.T) {
	for _, tc := range []struct {
		name string
		run  func(context.Context) (*pb.ListResourcesResponse, error)
		want codes.Code
	}{
		{"panic", func(context.Context) (*pb.ListResourcesResponse, error) { panic("secret panic") }, codes.Internal},
		{"deadline", func(ctx context.Context) (*pb.ListResourcesResponse, error) {
			<-ctx.Done()
			return nil, status.FromContextError(ctx.Err()).Err()
		}, codes.DeadlineExceeded},
	} {
		t.Run(tc.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			listener := bufconn.Listen(1 << 20)
			server := grpc.NewServer(ServerOptions(20*time.Millisecond, logger)...)
			pb.RegisterCatalogServiceServer(server, testCatalog{run: tc.run})
			done := make(chan error, 1)
			go func() { done <- server.Serve(listener) }()
			t.Cleanup(func() { server.Stop(); _ = listener.Close(); <-done })
			conn, err := grpc.NewClient("passthrough:///memory", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = conn.Close() })
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			_, err = pb.NewCatalogServiceClient(conn).ListResources(ctx, &pb.ListResourcesRequest{})
			if status.Code(err) != tc.want || strings.Contains(err.Error(), "secret") {
				t.Fatalf("error=%v want=%v", err, tc.want)
			}
		})
	}
}
