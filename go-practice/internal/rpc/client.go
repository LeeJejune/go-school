// Package rpc contains shared grpc-go client and interceptor configuration.
package rpc

import (
	"log/slog"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/timeout"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// NewClient creates a reusable connection. Actual RPCs report connection failures.
func NewClient(target string, budget time.Duration, logger *slog.Logger) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			timeout.UnaryClientInterceptor(budget),
			logging.UnaryClientInterceptor(logAdapter(logger)),
		),
	)
	if err != nil {
		return nil, err
	}
	conn.Connect()
	return conn, nil
}
