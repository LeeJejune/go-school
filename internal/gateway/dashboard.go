package gateway

import (
	"context"
	"encoding/json"
	"net/http"

	pb "github.com/LeeJejune/go-school/gen/school/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
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

// This is a Gateway-only HTTP DTO. protojson retains the same protobuf JSON
// representation as the generated routes, including lowerCamelCase fields.
func (d *Dashboard) Register(mux *runtime.ServeMux) error {
	return mux.HandlePath(http.MethodGet, "/v1/dashboard/{id}", func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		fail := func(err error) { runtime.DefaultHTTPErrorHandler(r.Context(), mux, &runtime.JSONPb{}, w, r, err) }
		snapshot, err := d.Get(r.Context(), params["id"])
		if err != nil {
			fail(err)
			return
		}
		catalogJSON, err := protojson.Marshal(snapshot.Catalog)
		if err != nil {
			fail(status.Error(codes.Internal, "could not encode dashboard"))
			return
		}
		listJSON, err := protojson.Marshal(snapshot.ReadingList)
		if err != nil {
			fail(status.Error(codes.Internal, "could not encode dashboard"))
			return
		}
		payload, err := json.Marshal(struct {
			Catalog     json.RawMessage `json:"catalog"`
			ReadingList json.RawMessage `json:"readingList"`
		}{catalogJSON, listJSON})
		if err != nil {
			fail(status.Error(codes.Internal, "could not encode dashboard"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(payload)
	})
}
