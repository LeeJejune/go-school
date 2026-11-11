package rpc

import (
	"context"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func logAdapter(logger *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, level logging.Level, message string, fields ...any) {
		logger.Log(ctx, slog.Level(level), message, fields...)
	})
}

func ServerOptions(budget time.Duration, logger *slog.Logger) []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.MaxRecvMsgSize(64 << 10),
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(logAdapter(logger)),
			func(ctx context.Context, req any, info *grpc.UnaryServerInfo, next grpc.UnaryHandler) (any, error) {
				ctx, cancel := context.WithTimeout(ctx, budget)
				defer cancel()
				return next(ctx, req)
			},
			recovery.UnaryServerInterceptor(recovery.WithRecoveryHandlerContext(func(ctx context.Context, recovered any) error {
				logger.ErrorContext(ctx, "RPC panic recovered", "panic", recovered, "stack", string(debug.Stack()))
				return status.Error(codes.Internal, "internal server error")
			})),
		),
	}
}
