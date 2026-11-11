package http

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func requestScope(budget time.Duration, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		ctx, cancel := context.WithTimeout(c.Request.Context(), budget)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 64<<10)
		defer func() {
			level := slog.LevelInfo
			if c.Writer.Status() >= 500 {
				level = slog.LevelError
			}
			logger.Log(ctx, level, "HTTP request", "method", c.Request.Method, "path", c.Request.URL.Path, "status", c.Writer.Status(), "duration", time.Since(started))
		}()
		c.Next()
	}
}

func recoveryMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, recovered any) {
		logger.ErrorContext(c.Request.Context(), "HTTP panic recovered", "panic", recovered, "stack", string(debug.Stack()))
		writeError(c, status.Error(codes.Internal, "internal server error"))
	})
}
