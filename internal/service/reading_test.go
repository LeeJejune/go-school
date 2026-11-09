package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/LeeJejune/go-school/internal/reading"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestErrorMappingAndSanitization(t *testing.T) {
	s := &Reading{logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	for _, tc := range []struct {
		err  error
		code codes.Code
	}{
		{reading.ErrInvalidInput, codes.InvalidArgument}, {reading.ErrConflict, codes.AlreadyExists},
		{reading.ErrNotFound, codes.NotFound}, {reading.ErrCatalogUnavailable, codes.Unavailable},
		{context.Canceled, codes.Canceled}, {context.DeadlineExceeded, codes.DeadlineExceeded},
		{errors.New("secret database credentials"), codes.Internal},
	} {
		err := s.rpcError(context.Background(), tc.err)
		if status.Code(err) != tc.code || strings.Contains(err.Error(), "secret") {
			t.Fatalf("got %v want %v", err, tc.code)
		}
	}
}
