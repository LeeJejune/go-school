package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	pb "github.com/LeeJejune/go-school/gen/school/v1"
	"github.com/LeeJejune/go-school/internal/platform"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() { platform.LogFatal(run()) }

func run() error {
	ctx, stop := platform.SignalContext()
	defer stop()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	mux := runtime.NewServeMux()
	if err := pb.RegisterCatalogServiceHandlerFromEndpoint(ctx, mux, platform.Env("CATALOG_TARGET", "127.0.0.1:9001"), opts); err != nil {
		return err
	}
	if err := pb.RegisterReadingServiceHandlerFromEndpoint(ctx, mux, platform.Env("READING_TARGET", "127.0.0.1:9002"), opts); err != nil {
		return err
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			w.WriteHeader(http.StatusOK) // Process liveness, not dependency readiness.
			_, _ = w.Write([]byte("ok\n"))
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
		requestCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		mux.ServeHTTP(w, r.WithContext(requestCtx))
	})
	server := &http.Server{
		Addr: platform.Env("HTTP_ADDR", "127.0.0.1:8080"), Handler: handler,
		ReadHeaderTimeout: 3 * time.Second, ReadTimeout: 5 * time.Second,
		WriteTimeout: 5 * time.Second, IdleTimeout: 30 * time.Second,
	}
	done := make(chan error, 1)
	go func() { done <- server.ListenAndServe() }()
	slog.Info("http listening", "address", server.Addr)
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			_ = server.Close()
		}
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
