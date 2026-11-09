package platform

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

func Env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func SignalContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}

// Every incoming RPC gets a budget, including calls made without a deadline.
func Budget(limit time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx, cancel := context.WithTimeout(ctx, limit)
		defer cancel()
		started := time.Now()
		response, err := handler(ctx, req)
		slog.Info("rpc", "method", info.FullMethod, "code", status.Code(err).String(), "duration", time.Since(started))
		return response, err
	}
}

func ServeGRPC(ctx context.Context, addr string, register func(*grpc.Server)) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	server := grpc.NewServer(grpc.UnaryInterceptor(Budget(2*time.Second)), grpc.MaxRecvMsgSize(64<<10))
	register(server)
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	reflection.Register(server) // Local learning server; disable or restrict in production.
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	slog.Info("grpc listening", "address", addr)
	select {
	case err := <-done:
		if errors.Is(err, grpc.ErrServerStopped) {
			return nil
		}
		return err
	case <-ctx.Done():
		healthServer.Shutdown()
		stopped := make(chan struct{})
		go func() { server.GracefulStop(); close(stopped) }()
		select {
		case <-stopped:
		case <-time.After(3 * time.Second):
			server.Stop()
		}
		return nil
	}
}

func LogFatal(err error) {
	if err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
