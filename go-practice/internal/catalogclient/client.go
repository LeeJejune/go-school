package catalogclient

import (
	"context"
	"log/slog"

	pb "github.com/LeeJejune/go-school/go-practice/gen/school/v1"
	"github.com/LeeJejune/go-school/go-practice/internal/config"
	"github.com/LeeJejune/go-school/go-practice/internal/reading"
	"github.com/LeeJejune/go-school/go-practice/internal/rpc"
	"go.uber.org/fx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Client struct{ client pb.CatalogServiceClient }

func New(lc fx.Lifecycle, cfg config.Config, logger *slog.Logger) (*Client, error) {
	conn, err := rpc.NewClient(cfg.CatalogTarget, cfg.CatalogTimeout, logger)
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
