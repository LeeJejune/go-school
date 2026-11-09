// Package service contains transport adapters: the equivalent of Spring controllers.
package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	pb "github.com/LeeJejune/go-school/gen/school/v1"
	"github.com/LeeJejune/go-school/internal/reading"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Reading struct {
	pb.UnimplementedReadingServiceServer
	usecase *reading.UseCase
	logger  *slog.Logger
}

func NewReading(usecase *reading.UseCase, logger *slog.Logger) *Reading {
	return &Reading{usecase: usecase, logger: logger}
}

func (s *Reading) CreateList(ctx context.Context, req *pb.CreateListRequest) (*pb.CreateListResponse, error) {
	list, err := s.usecase.Create(ctx, reading.CreateInput{RequestKey: req.GetRequestKey(), Title: req.GetTitle(), ResourceIDs: req.GetResourceIds()})
	if err != nil {
		return nil, s.rpcError(ctx, err)
	}
	return &pb.CreateListResponse{ReadingList: toProto(list)}, nil
}

func (s *Reading) GetList(ctx context.Context, req *pb.GetListRequest) (*pb.GetListResponse, error) {
	list, err := s.usecase.Get(ctx, req.GetId())
	if err != nil {
		return nil, s.rpcError(ctx, err)
	}
	return &pb.GetListResponse{ReadingList: toProto(list)}, nil
}

func (s *Reading) SetRead(ctx context.Context, req *pb.SetReadRequest) (*pb.SetReadResponse, error) {
	if err := s.usecase.SetRead(ctx, req.GetListId(), req.GetResourceId(), req.GetIsRead()); err != nil {
		return nil, s.rpcError(ctx, err)
	}
	return &pb.SetReadResponse{}, nil
}

func (s *Reading) rpcError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, reading.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, reading.ErrNotFound):
		return status.Error(codes.NotFound, reading.ErrNotFound.Error())
	case errors.Is(err, reading.ErrConflict):
		return status.Error(codes.AlreadyExists, reading.ErrConflict.Error())
	case errors.Is(err, reading.ErrCatalogUnavailable):
		return status.Error(codes.Unavailable, reading.ErrCatalogUnavailable.Error())
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return status.FromContextError(err).Err()
	case ctx.Err() != nil:
		return status.FromContextError(ctx.Err()).Err()
	default:
		s.logger.ErrorContext(ctx, "reading operation failed", "error", err)
		return status.Error(codes.Internal, "could not persist reading list")
	}
}

func toProto(list reading.List) *pb.ReadingList {
	out := &pb.ReadingList{Id: list.ID, Title: list.Title, CreatedAt: list.CreatedAt.UTC().Format(time.RFC3339Nano)}
	for _, item := range list.Items {
		out.Items = append(out.Items, &pb.ReadingItem{Resource: &pb.Resource{Id: item.Resource.ID, Title: item.Resource.Title, Url: item.Resource.URL}, IsRead: item.IsRead})
	}
	return out
}
