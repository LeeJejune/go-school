package catalogclient

import (
	"context"

	pb "github.com/LeeJejune/go-school/gen/school/v1"
	"github.com/LeeJejune/go-school/internal/reading"
)

type Client struct{ client pb.CatalogServiceClient }

func New(client pb.CatalogServiceClient) *Client { return &Client{client: client} }

func (c *Client) GetResources(ctx context.Context, ids []string) ([]reading.Resource, error) {
	response, err := c.client.BatchGetResources(ctx, &pb.BatchGetResourcesRequest{Ids: ids})
	if err != nil {
		return nil, err
	}
	resources := make([]reading.Resource, 0, len(response.Resources))
	for _, resource := range response.Resources {
		resources = append(resources, reading.Resource{ID: resource.Id, Title: resource.Title, URL: resource.Url})
	}
	return resources, nil
}
