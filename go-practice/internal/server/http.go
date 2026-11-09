package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	pb "github.com/LeeJejune/go-school/go-practice/gen/school/v1"
	"github.com/LeeJejune/go-school/go-practice/internal/config"
	"github.com/LeeJejune/go-school/go-practice/internal/gateway"
	khttp "github.com/go-kratos/kratos/v3/transport/http"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/fx"
)

func NewHTTP(lc fx.Lifecycle, shutdown fx.Shutdowner, cfg config.Config, logger *slog.Logger, clients *gateway.Clients, dashboard *gateway.Dashboard) (*khttp.Server, error) {
	mux := runtime.NewServeMux()
	if err := pb.RegisterCatalogServiceHandlerClient(context.Background(), mux, clients.Catalog); err != nil {
		return nil, err
	}
	if err := pb.RegisterReadingServiceHandlerClient(context.Background(), mux, clients.Reading); err != nil {
		return nil, err
	}
	if err := dashboard.Register(mux); err != nil {
		return nil, err
	}
	srv := khttp.NewServer(khttp.Address(cfg.HTTPAddr), khttp.Timeout(cfg.RequestTimeout), khttp.Filter(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			defer func() {
				logger.InfoContext(r.Context(), "http request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(started))
			}()
			r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
			next.ServeHTTP(w, r)
		})
	}))
	srv.ReadHeaderTimeout = 3 * time.Second
	srv.ReadTimeout = cfg.RequestTimeout + 3*time.Second
	srv.WriteTimeout = cfg.RequestTimeout + 3*time.Second
	srv.IdleTimeout = 30 * time.Second
	srv.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok\n")) })
	srv.HandlePrefix("/", mux)
	manage(lc, shutdown, logger, srv)
	return srv, nil
}
