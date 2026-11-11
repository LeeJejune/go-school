// Package app declares the dependency graph. Business code never calls the DI container.
package app

import (
	"log/slog"
	"os"
	"time"

	"github.com/LeeJejune/go-school/go-practice/internal/catalog"
	"github.com/LeeJejune/go-school/go-practice/internal/catalogclient"
	"github.com/LeeJejune/go-school/go-practice/internal/config"
	grpccontroller "github.com/LeeJejune/go-school/go-practice/internal/controller/grpc"
	httpcontroller "github.com/LeeJejune/go-school/go-practice/internal/controller/http"
	"github.com/LeeJejune/go-school/go-practice/internal/gateway"
	"github.com/LeeJejune/go-school/go-practice/internal/reading"
	mysqlrepo "github.com/LeeJejune/go-school/go-practice/internal/repository/mysql"
	"github.com/LeeJejune/go-school/go-practice/internal/server"
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
		fx.Provide(catalog.New, grpccontroller.NewCatalog, server.NewCatalog),
		fx.Invoke(func(*server.GRPCServer) {}),
	))
}

func Reading() fx.Option {
	return fx.Options(common("reading"), fx.Module("reading",
		fx.Provide(
			mysqlrepo.NewDB,
			fx.Annotate(mysqlrepo.NewReadingRepository, fx.As(new(reading.Repository))),
			fx.Annotate(catalogclient.New, fx.As(new(reading.Catalog))),
			func(repo reading.Repository, catalog reading.Catalog, cfg config.Config) *reading.UseCase {
				return reading.New(repo, catalog, cfg.CatalogTimeout)
			},
			grpccontroller.NewReading, server.NewReading,
		),
		fx.Invoke(func(*server.GRPCServer) {}),
	))
}

func Gateway() fx.Option {
	return fx.Options(common("gateway"), fx.Module("gateway",
		fx.Provide(gateway.NewClients, gateway.NewDashboard, httpcontroller.NewHandler, httpcontroller.NewRouter, server.NewHTTP),
		fx.Invoke(func(*server.HTTPServer) {}),
	))
}
