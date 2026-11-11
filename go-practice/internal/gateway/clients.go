package gateway

import (
	"context"
	"errors"
	"log/slog"

	pb "github.com/LeeJejune/go-school/go-practice/gen/school/v1"
	"github.com/LeeJejune/go-school/go-practice/internal/config"
	"github.com/LeeJejune/go-school/go-practice/internal/rpc"
	"go.uber.org/fx"
)

type Clients struct {
	Catalog pb.CatalogServiceClient
	Reading pb.ReadingServiceClient
}

func NewClients(lc fx.Lifecycle, cfg config.Config, logger *slog.Logger) (*Clients, error) {
	catalogConn, err := rpc.NewClient(cfg.CatalogTarget, cfg.RequestTimeout, logger)
	if err != nil {
		return nil, err
	}
	readingConn, err := rpc.NewClient(cfg.ReadingTarget, cfg.RequestTimeout, logger)
	if err != nil {
		_ = catalogConn.Close()
		return nil, err
	}
	lc.Append(fx.Hook{OnStop: func(context.Context) error { return errors.Join(catalogConn.Close(), readingConn.Close()) }})
	return &Clients{Catalog: pb.NewCatalogServiceClient(catalogConn), Reading: pb.NewReadingServiceClient(readingConn)}, nil
}
