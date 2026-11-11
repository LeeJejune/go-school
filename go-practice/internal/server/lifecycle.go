// Package server connects HTTP/gRPC transports to the Fx application lifecycle.
package server

import (
	"context"
	"errors"
	"log/slog"
	"net/url"

	"go.uber.org/fx"
	"google.golang.org/grpc"
)

type managedServer interface {
	Endpoint() (*url.URL, error)
	Start(context.Context) error
	Stop(context.Context) error
}

// OnStart must return: the long-running server belongs in a goroutine. Binding
// first makes an occupied port a synchronous startup failure visible to Fx.
func manage(lc fx.Lifecycle, shutdown fx.Shutdowner, logger *slog.Logger, server managedServer) {
	done := make(chan struct{})
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			endpoint, err := server.Endpoint()
			if err != nil {
				return err
			}
			logger.InfoContext(ctx, "starting server", "endpoint", endpoint.String())
			go func() {
				defer close(done)
				// Fx's startup context expires after startup; it is not the server lifetime.
				if err := server.Start(context.Background()); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
					logger.Error("server failed", "error", err)
					_ = shutdown.Shutdown(fx.ExitCode(1))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if err := server.Stop(ctx); err != nil {
				return err
			}
			select {
			case <-done:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	})
}
