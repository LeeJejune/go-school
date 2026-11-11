package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	pb "github.com/LeeJejune/go-school/go-practice/gen/school/v1"
	"github.com/LeeJejune/go-school/go-practice/internal/config"
	"github.com/LeeJejune/go-school/go-practice/internal/gateway"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type catalogClientStub struct {
	pb.CatalogServiceClient
	list   func(context.Context) (*pb.ListResourcesResponse, error)
	upsert func(context.Context, *pb.UpsertResourceRequest) (*pb.UpsertResourceResponse, error)
}

func (s catalogClientStub) UpsertResource(ctx context.Context, req *pb.UpsertResourceRequest, _ ...grpc.CallOption) (*pb.UpsertResourceResponse, error) {
	return s.upsert(ctx, req)
}

func (s catalogClientStub) ListResources(ctx context.Context, _ *pb.ListResourcesRequest, _ ...grpc.CallOption) (*pb.ListResourcesResponse, error) {
	if s.list == nil {
		return &pb.ListResourcesResponse{}, nil
	}
	return s.list(ctx)
}

type readingClientStub struct {
	pb.ReadingServiceClient
	create func(context.Context, *pb.CreateListRequest) (*pb.CreateListResponse, error)
	get    func(context.Context, *pb.GetListRequest) (*pb.GetListResponse, error)
	set    func(context.Context, *pb.SetReadRequest) (*pb.SetReadResponse, error)
}

func (s readingClientStub) CreateList(ctx context.Context, req *pb.CreateListRequest, _ ...grpc.CallOption) (*pb.CreateListResponse, error) {
	if s.create == nil {
		return nil, status.Error(codes.Unimplemented, "unexpected create")
	}
	return s.create(ctx, req)
}
func (s readingClientStub) GetList(ctx context.Context, req *pb.GetListRequest, _ ...grpc.CallOption) (*pb.GetListResponse, error) {
	if s.get == nil {
		return nil, status.Error(codes.Unimplemented, "unexpected get")
	}
	return s.get(ctx, req)
}
func (s readingClientStub) SetRead(ctx context.Context, req *pb.SetReadRequest, _ ...grpc.CallOption) (*pb.SetReadResponse, error) {
	if s.set == nil {
		return nil, status.Error(codes.Unimplemented, "unexpected update")
	}
	return s.set(ctx, req)
}
func testRouter(clients *gateway.Clients, budget time.Duration) *gin.Engine {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewRouter(config.Config{RequestTimeout: budget}, logger, NewHandler(clients, gateway.NewDashboard(clients)))
}
func serve(router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, request)
	return response
}

func TestGinJSONContractAndRPCMapping(t *testing.T) {
	var resource *pb.UpsertResourceRequest
	var created *pb.CreateListRequest
	var updated *pb.SetReadRequest
	list := &pb.ReadingList{Id: "saved", Title: "Go", CreatedAt: "today", Items: []*pb.ReadingItem{{Resource: &pb.Resource{Id: "go-tour"}}}}
	clients := &gateway.Clients{
		Catalog: catalogClientStub{upsert: func(_ context.Context, req *pb.UpsertResourceRequest) (*pb.UpsertResourceResponse, error) {
			resource = req
			return &pb.UpsertResourceResponse{Resource: &pb.Resource{Id: req.Id, Title: req.Title, Url: req.Url}}, nil
		}},
		Reading: readingClientStub{
			create: func(_ context.Context, req *pb.CreateListRequest) (*pb.CreateListResponse, error) {
				created = req
				return &pb.CreateListResponse{ReadingList: list}, nil
			},
			get: func(_ context.Context, req *pb.GetListRequest) (*pb.GetListResponse, error) {
				if req.Id != "saved" {
					return nil, status.Error(codes.NotFound, "missing")
				}
				return &pb.GetListResponse{ReadingList: list}, nil
			},
			set: func(_ context.Context, req *pb.SetReadRequest) (*pb.SetReadResponse, error) {
				updated = req
				return &pb.SetReadResponse{}, nil
			},
		},
	}
	router := testRouter(clients, time.Second)
	upserted := serve(router, "PUT", "/v1/resources/my-resource", `{"title":"My resource","url":"https://example.com/"}`)
	if upserted.Code != 200 || resource == nil || resource.Id != "my-resource" || resource.Title != "My resource" || !strings.Contains(upserted.Body.String(), `"url":"https://example.com/"`) {
		t.Fatalf("upsert status=%d body=%s request=%v", upserted.Code, upserted.Body, resource)
	}
	result := serve(router, "POST", "/v1/lists", `{"requestKey":"study-001","title":"Go","resourceIds":["go-tour"],"ignored":1}`)
	if result.Code != 200 || created == nil || created.RequestKey != "study-001" || len(created.ResourceIds) != 1 {
		t.Fatalf("status=%d body=%s request=%v", result.Code, result.Body, created)
	}
	var response struct {
		ReadingList struct {
			CreatedAt string `json:"createdAt"`
			Items     []struct {
				IsRead *bool `json:"isRead"`
			} `json:"items"`
		} `json:"readingList"`
	}
	if err := json.Unmarshal(result.Body.Bytes(), &response); err != nil || response.ReadingList.CreatedAt != "today" || response.ReadingList.Items[0].IsRead == nil || *response.ReadingList.Items[0].IsRead {
		t.Fatalf("response=%s error=%v", result.Body, err)
	}
	result = serve(router, "PATCH", "/v1/lists/saved/items/go-tour", `{"isRead":false}`)
	if result.Code != 200 || updated == nil || updated.IsRead || updated.ListId != "saved" || updated.ResourceId != "go-tour" {
		t.Fatalf("update=%v status=%d", updated, result.Code)
	}
	updated = nil
	result = serve(router, "PATCH", "/v1/lists/saved/items/go-tour", `{}`)
	if result.Code != 200 || updated == nil || updated.IsRead {
		t.Fatal("omitted isRead must retain the documented false default")
	}
	result = serve(router, "GET", "/v1/dashboard/saved", "")
	if result.Code != 200 || !strings.Contains(result.Body.String(), `"catalog":{"resources":[]}`) || !strings.Contains(result.Body.String(), `"readingList"`) {
		t.Fatalf("dashboard=%s status=%d", result.Body, result.Code)
	}
	if result := serve(router, "GET", "/v1/lists/missing", ""); result.Code != 404 {
		t.Fatalf("missing=%d", result.Code)
	}
	if result := serve(router, "GET", "/missing", ""); result.Code != 404 {
		t.Fatalf("route=%d", result.Code)
	}
	if result := serve(router, "DELETE", "/v1/resources", ""); result.Code != 405 {
		t.Fatalf("method=%d", result.Code)
	}
}

func TestGinRejectsInvalidBodiesBeforeRPC(t *testing.T) {
	called := false
	clients := &gateway.Clients{Catalog: catalogClientStub{}, Reading: readingClientStub{set: func(context.Context, *pb.SetReadRequest) (*pb.SetReadResponse, error) {
		called = true
		return &pb.SetReadResponse{}, nil
	}}}
	router := testRouter(clients, time.Second)
	for _, body := range []string{"", "{", "null", "[]", `{} {}`, `{"isRead":"false"}`, `{"ignored":"` + strings.Repeat("x", 65<<10) + `"}`} {
		result := serve(router, "PATCH", "/v1/lists/saved/items/go-tour", body)
		if result.Code != 400 || called {
			t.Fatalf("status=%d RPC called=%v", result.Code, called)
		}
	}
	result := serve(router, "POST", "/v1/lists", `{"title":"Go"}`)
	if result.Code != 400 {
		t.Fatalf("missing required fields=%d", result.Code)
	}
	result = serve(router, "PUT", "/v1/resources/my-resource", `{"title":"Go"}`)
	if result.Code != 400 {
		t.Fatalf("missing URL=%d", result.Code)
	}
}

func TestGinTimeoutRecoveryAndErrorSanitization(t *testing.T) {
	for _, tc := range []struct {
		name string
		run  func(context.Context) (*pb.ListResourcesResponse, error)
		want int
		code codes.Code
	}{
		{"timeout", func(ctx context.Context) (*pb.ListResourcesResponse, error) { <-ctx.Done(); return nil, ctx.Err() }, 504, codes.DeadlineExceeded},
		{"panic", func(context.Context) (*pb.ListResourcesResponse, error) { panic("secret panic") }, 500, codes.Internal},
		{"internal", func(context.Context) (*pb.ListResourcesResponse, error) {
			return nil, status.Error(codes.Internal, "secret SQL")
		}, 500, codes.Internal},
		{"unknown", func(context.Context) (*pb.ListResourcesResponse, error) { return nil, errors.New("secret credentials") }, 500, codes.Internal},
		{"unavailable", func(context.Context) (*pb.ListResourcesResponse, error) {
			return nil, status.Error(codes.Unavailable, "secret target")
		}, 503, codes.Unavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			router := testRouter(&gateway.Clients{Catalog: catalogClientStub{list: tc.run}, Reading: readingClientStub{}}, 10*time.Millisecond)
			response := serve(router, "GET", "/v1/resources", "")
			var body errorResponse
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || response.Code != tc.want || body.Code != int32(tc.code) || strings.Contains(response.Body.String(), "secret") {
				t.Fatalf("status=%d body=%s error=%v", response.Code, response.Body, err)
			}
		})
	}
}

func TestGinPassesRequestCancellationToRPC(t *testing.T) {
	clients := &gateway.Clients{Catalog: catalogClientStub{list: func(ctx context.Context) (*pb.ListResourcesResponse, error) { <-ctx.Done(); return nil, ctx.Err() }}, Reading: readingClientStub{}}
	router := testRouter(clients, time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	request := httptest.NewRequest("GET", "/v1/resources", nil).WithContext(ctx)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != 499 {
		t.Fatalf("cancel status=%d body=%s", response.Code, response.Body)
	}
}
