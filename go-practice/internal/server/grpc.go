package server

import (
	"log/slog"

	pb "github.com/LeeJejune/go-school/go-practice/gen/school/v1"
	"github.com/LeeJejune/go-school/go-practice/internal/config"
	"github.com/LeeJejune/go-school/go-practice/internal/service"
	"github.com/go-kratos/kratos/v3/middleware/logging"
	"github.com/go-kratos/kratos/v3/middleware/recovery"
	kgrpc "github.com/go-kratos/kratos/v3/transport/grpc"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

func newGRPC(addr string, cfg config.Config, logger *slog.Logger) *kgrpc.Server {
	return kgrpc.NewServer(
		kgrpc.Address(addr), kgrpc.Timeout(cfg.RequestTimeout),
		kgrpc.Middleware(recovery.Recovery(recovery.WithLogger(logger)), logging.Server(logger)),
		kgrpc.Options(grpc.MaxRecvMsgSize(64<<10)),
	)
}

func NewCatalog(lc fx.Lifecycle, shutdown fx.Shutdowner, cfg config.Config, logger *slog.Logger, handler *service.Catalog) *kgrpc.Server {
	srv := newGRPC(cfg.CatalogAddr, cfg, logger)
	pb.RegisterCatalogServiceServer(srv, handler)
	manage(lc, shutdown, logger, srv)
	return srv
}

func NewReading(lc fx.Lifecycle, shutdown fx.Shutdowner, cfg config.Config, logger *slog.Logger, handler *service.Reading) *kgrpc.Server {
	srv := newGRPC(cfg.ReadingAddr, cfg, logger)
	pb.RegisterReadingServiceServer(srv, handler)
	manage(lc, shutdown, logger, srv)
	return srv
}
