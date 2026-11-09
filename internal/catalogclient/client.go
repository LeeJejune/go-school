package catalogclient

import (
	"context"
	"log/slog"

	pb "github.com/LeeJejune/go-school/gen/school/v1"
	"github.com/LeeJejune/go-school/internal/config"
	"github.com/LeeJejune/go-school/internal/reading"
	"github.com/go-kratos/kratos/v3/middleware/logging"
	kgrpc "github.com/go-kratos/kratos/v3/transport/grpc"
	"go.uber.org/fx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Client struct{ client pb.CatalogServiceClient }

func New(lc fx.Lifecycle, cfg config.Config, logger *slog.Logger) (*Client, error) {
	conn, err := kgrpc.NewClient(context.Background(), kgrpc.WithEndpoint(cfg.CatalogTarget), kgrpc.WithTimeout(cfg.CatalogTimeout), kgrpc.WithMiddleware(logging.Client(logger)))
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{OnStop: func(context.Context) error { return conn.Close() }})
	return &Client{client: pb.NewCatalogServiceClient(conn)}, nil
}

func (c *Client) GetResources(ctx context.Context, ids []string) ([]reading.Resource, error) {
	response, err := c.client.BatchGetResources(ctx, &pb.BatchGetResourcesRequest{Ids: ids})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, reading.ErrResourceNotFound
		}
		return nil, err
	}
	resources := make([]reading.Resource, 0, len(response.Resources))
	for _, resource := range response.Resources {
		resources = append(resources, reading.Resource{ID: resource.Id, Title: resource.Title, URL: resource.Url})
	}
	return resources, nil
}
