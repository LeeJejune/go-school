package server

import (
	"context"
	"log/slog"
	"net"
	"net/url"

	pb "github.com/LeeJejune/go-school/go-practice/gen/school/v1"
	"github.com/LeeJejune/go-school/go-practice/internal/config"
	grpccontroller "github.com/LeeJejune/go-school/go-practice/internal/controller/grpc"
	"github.com/LeeJejune/go-school/go-practice/internal/rpc"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type GRPCServer struct {
	*grpc.Server
	health   *health.Server
	address  string
	listener net.Listener
}

func newGRPC(address string, cfg config.Config, logger *slog.Logger) *GRPCServer {
	srv := &GRPCServer{Server: grpc.NewServer(rpc.ServerOptions(cfg.RequestTimeout, logger)...), address: address, health: health.NewServer()}
	healthpb.RegisterHealthServer(srv.Server, srv.health)
	reflection.Register(srv.Server) // Local learning setup.
	return srv
}

func NewCatalog(lc fx.Lifecycle, shutdown fx.Shutdowner, cfg config.Config, logger *slog.Logger, handler *grpccontroller.Catalog) *GRPCServer {
	srv := newGRPC(cfg.CatalogAddr, cfg, logger)
	pb.RegisterCatalogServiceServer(srv.Server, handler)
	manage(lc, shutdown, logger, srv)
	return srv
}

func NewReading(lc fx.Lifecycle, shutdown fx.Shutdowner, cfg config.Config, logger *slog.Logger, handler *grpccontroller.Reading) *GRPCServer {
	srv := newGRPC(cfg.ReadingAddr, cfg, logger)
	pb.RegisterReadingServiceServer(srv.Server, handler)
	manage(lc, shutdown, logger, srv)
	return srv
}

func (s *GRPCServer) Endpoint() (*url.URL, error) {
	if s.listener == nil {
		listener, err := net.Listen("tcp", s.address)
		if err != nil {
			return nil, err
		}
		s.listener = listener
	}
	return &url.URL{Scheme: "grpc", Host: s.listener.Addr().String()}, nil
}

func (s *GRPCServer) Start(context.Context) error {
	s.health.Resume()
	return s.Serve(s.listener)
}

func (s *GRPCServer) Stop(ctx context.Context) error {
	s.health.Shutdown()
	done := make(chan struct{})
	go func() { s.GracefulStop(); close(done) }()
	select {
	case <-done:
	case <-ctx.Done():
		s.Server.Stop()
	}
	return nil
}
