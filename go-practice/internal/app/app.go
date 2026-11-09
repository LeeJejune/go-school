// Package app declares the dependency graph. Business code never calls the DI container.
package app

import (
	"log/slog"
	"os"
	"time"

	"github.com/LeeJejune/go-school/go-practice/internal/catalog"
	"github.com/LeeJejune/go-school/go-practice/internal/catalogclient"
	"github.com/LeeJejune/go-school/go-practice/internal/config"
	"github.com/LeeJejune/go-school/go-practice/internal/data"
	"github.com/LeeJejune/go-school/go-practice/internal/gateway"
	"github.com/LeeJejune/go-school/go-practice/internal/reading"
	"github.com/LeeJejune/go-school/go-practice/internal/server"
	"github.com/LeeJejune/go-school/go-practice/internal/service"
	kgrpc "github.com/go-kratos/kratos/v3/transport/grpc"
	khttp "github.com/go-kratos/kratos/v3/transport/http"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func common(name string) fx.Option {
	return fx.Options(
		fx.Provide(config.Load, func() *slog.Logger { return slog.New(slog.NewJSONHandler(os.Stderr, nil)).With("service", name) }),
		fx.WithLogger(func(logger *slog.Logger) fxevent.Logger {
			out := &fxevent.SlogLogger{Logger: logger}
			out.UseLogLevel(slog.LevelDebug)
			return out
		}),
		fx.StartTimeout(10*time.Second), fx.StopTimeout(5*time.Second),
	)
}

func Catalog() fx.Option {
	return fx.Options(common("catalog"), fx.Module("catalog",
		fx.Provide(catalog.New, service.NewCatalog, server.NewCatalog),
		fx.Invoke(func(*kgrpc.Server) {}),
	))
}

func Reading() fx.Option {
	return fx.Options(common("reading"), fx.Module("reading",
		fx.Provide(
			data.NewDB,
			fx.Annotate(data.NewReadingRepository, fx.As(new(reading.Repository))),
			fx.Annotate(catalogclient.New, fx.As(new(reading.Catalog))),
			func(repo reading.Repository, catalog reading.Catalog, cfg config.Config) *reading.UseCase {
				return reading.New(repo, catalog, cfg.CatalogTimeout)
			},
			service.NewReading, server.NewReading,
		),
		fx.Invoke(func(*kgrpc.Server) {}),
	))
}

func Gateway() fx.Option {
	return fx.Options(common("gateway"), fx.Module("gateway",
		fx.Provide(gateway.NewClients, gateway.NewDashboard, server.NewHTTP),
		fx.Invoke(func(*khttp.Server) {}),
	))
}
