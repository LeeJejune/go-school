package gateway

import (
	"context"

	pb "github.com/LeeJejune/go-school/go-practice/gen/school/v1"
	"golang.org/x/sync/errgroup"
)

type Dashboard struct{ clients *Clients }

func NewDashboard(clients *Clients) *Dashboard { return &Dashboard{clients: clients} }

type Snapshot struct {
	Catalog     *pb.ListResourcesResponse
	ReadingList *pb.ReadingList
}

// Get is a BFF use case: independent RPCs run concurrently, but both must
// finish before the response. The request's cancellation reaches both clients.
func (d *Dashboard) Get(ctx context.Context, id string) (*Snapshot, error) {
	g, ctx := errgroup.WithContext(ctx)
	var catalog *pb.ListResourcesResponse
	var list *pb.GetListResponse
	g.Go(func() error {
		var err error
		catalog, err = d.clients.Catalog.ListResources(ctx, &pb.ListResourcesRequest{})
		return err
	})
	g.Go(func() error {
		var err error
		list, err = d.clients.Reading.GetList(ctx, &pb.GetListRequest{Id: id})
		return err
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return &Snapshot{Catalog: catalog, ReadingList: list.GetReadingList()}, nil
}
