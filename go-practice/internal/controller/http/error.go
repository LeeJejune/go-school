package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type errorResponse struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Details []any  `json:"details"`
}

func bindJSON(c *gin.Context, request any) bool {
	if err := c.ShouldBindBodyWithJSON(request); err != nil {
		writeError(c, status.Error(codes.InvalidArgument, "invalid JSON body"))
		return false
	}
	body, _ := c.Get(gin.BodyBytesKey)
	data, _ := body.([]byte)
	// Match the old protobuf endpoint's single-object JSON contract. In
	// particular, null must not accidentally change an item's bool to false.
	trimmed := bytes.TrimSpace(data)
	if !json.Valid(data) || len(trimmed) == 0 || trimmed[0] != '{' {
		writeError(c, status.Error(codes.InvalidArgument, "expected one JSON object"))
		return false
	}
	return true
}

func writeError(c *gin.Context, err error) {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		err = status.FromContextError(err).Err()
	}
	rpcStatus, ok := status.FromError(err)
	if !ok {
		rpcStatus = status.New(codes.Internal, "internal server error")
	}
	code, message := rpcStatus.Code(), rpcStatus.Message()
	httpStatus := http.StatusInternalServerError
	switch code {
	case codes.InvalidArgument, codes.FailedPrecondition, codes.OutOfRange:
		httpStatus = http.StatusBadRequest
	case codes.NotFound:
		httpStatus = http.StatusNotFound
	case codes.AlreadyExists, codes.Aborted:
		httpStatus = http.StatusConflict
	case codes.Unauthenticated:
		httpStatus = http.StatusUnauthorized
	case codes.PermissionDenied:
		httpStatus = http.StatusForbidden
	case codes.ResourceExhausted:
		httpStatus = http.StatusTooManyRequests
	case codes.Canceled:
		httpStatus = 499
	case codes.DeadlineExceeded:
		httpStatus = http.StatusGatewayTimeout
	case codes.Unavailable:
		httpStatus = http.StatusServiceUnavailable
		message = "upstream service unavailable"
	case codes.Unimplemented:
		httpStatus = http.StatusNotImplemented
	default:
		code, message = codes.Internal, "internal server error"
	}
	c.AbortWithStatusJSON(httpStatus, errorResponse{Code: int32(code), Message: message, Details: []any{}})
}
