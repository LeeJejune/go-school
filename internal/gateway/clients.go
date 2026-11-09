package gateway

import (
	"context"
	"log/slog"

	pb "github.com/LeeJejune/go-school/gen/school/v1"
	"github.com/LeeJejune/go-school/internal/config"
	"github.com/go-kratos/kratos/v3/middleware/logging"
	kgrpc "github.com/go-kratos/kratos/v3/transport/grpc"
	"go.uber.org/fx"
)

type Clients struct {
	Catalog pb.CatalogServiceClient
	Reading pb.ReadingServiceClient
}

func NewClients(lc fx.Lifecycle, cfg config.Config, logger *slog.Logger) (*Clients, error) {
	catalogConn, err := kgrpc.NewClient(context.Background(), kgrpc.WithEndpoint(cfg.CatalogTarget), kgrpc.WithTimeout(cfg.RequestTimeout), kgrpc.WithMiddleware(logging.Client(logger)))
	if err != nil {
		return nil, err
	}
	readingConn, err := kgrpc.NewClient(context.Background(), kgrpc.WithEndpoint(cfg.ReadingTarget), kgrpc.WithTimeout(cfg.RequestTimeout), kgrpc.WithMiddleware(logging.Client(logger)))
	if err != nil {
		_ = catalogConn.Close()
		return nil, err
	}
	lc.Append(fx.Hook{OnStop: func(context.Context) error {
		catalogErr := catalogConn.Close()
		readingErr := readingConn.Close()
		if catalogErr != nil {
			return catalogErr
		}
		return readingErr
	}})
	return &Clients{Catalog: pb.NewCatalogServiceClient(catalogConn), Reading: pb.NewReadingServiceClient(readingConn)}, nil
}
