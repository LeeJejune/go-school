package app

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	pb "github.com/LeeJejune/go-school/go-practice/gen/school/v1"
	"github.com/LeeJejune/go-school/go-practice/internal/config"
	"github.com/LeeJejune/go-school/go-practice/internal/server"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"google.golang.org/protobuf/encoding/protojson"
)

// Real TCP HTTP -> Gateway -> gRPC -> Reading/Catalog -> GORM -> MySQL.
// Requires the dedicated local database and make migrate; no mocks or sleeps.
func TestFullStack(t *testing.T) {
	dsn := os.Getenv("MYSQL_TEST_DSN")
	if dsn == "" {
		t.Skip("MYSQL_TEST_DSN is not set")
	}
	var catalogServer *server.GRPCServer
	catalogApp := fxtest.New(t, Catalog(), fx.NopLogger, fx.Decorate(func(cfg config.Config) config.Config {
		cfg.CatalogAddr = "127.0.0.1:0"
		cfg.MySQLDSN = dsn
		return cfg
	}), fx.Populate(&catalogServer))
	catalogApp.RequireStart()
	catalogStopped := false
	t.Cleanup(func() {
		if !catalogStopped {
			catalogApp.RequireStop()
		}
	})
	catalogEndpoint, err := catalogServer.Endpoint()
	if err != nil {
		t.Fatal(err)
	}
	var readingServer *server.GRPCServer
	readingApp := fxtest.New(t, Reading(), fx.NopLogger, fx.Decorate(func(cfg config.Config) config.Config {
		cfg.ReadingAddr = "127.0.0.1:0"
		cfg.CatalogTarget = catalogEndpoint.Host
		cfg.MySQLDSN = dsn
		return cfg
	}), fx.Populate(&readingServer))
	readingApp.RequireStart()
	t.Cleanup(func() { readingApp.RequireStop() })
	readingEndpoint, err := readingServer.Endpoint()
	if err != nil {
		t.Fatal(err)
	}
	var gatewayServer *server.HTTPServer
	gatewayApp := fxtest.New(t, Gateway(), fx.NopLogger, fx.Decorate(func(cfg config.Config) config.Config {
		cfg.HTTPAddr = "127.0.0.1:0"
		cfg.CatalogTarget = catalogEndpoint.Host
		cfg.ReadingTarget = readingEndpoint.Host
		return cfg
	}), fx.Populate(&gatewayServer))
	gatewayApp.RequireStart()
	t.Cleanup(func() { gatewayApp.RequireStop() })
	endpoint, err := gatewayServer.Endpoint()
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Timeout: 4 * time.Second}
	t.Cleanup(client.CloseIdleConnections)
	call := func(method, path string, payload any, want int) []byte {
		t.Helper()
		body, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest(method, endpoint.String()+path, bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		response, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		result, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != want {
			t.Fatalf("%s %s: status=%d want=%d body=%s", method, path, response.StatusCode, want, result)
		}
		return result
	}
	call("GET", "/healthz", nil, 200)
	key := fmt.Sprintf("e2e-%d", time.Now().UnixNano())
	first, second := key+"-one", key+"-two"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		defer db.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if _, err := db.ExecContext(ctx, "DELETE FROM reading_lists WHERE request_key IN (?, ?, ?)", key, key+"-bad", key+"-down"); err != nil {
			t.Error(err)
		}
		if _, err := db.ExecContext(ctx, "DELETE FROM catalog_resources WHERE id IN (?, ?)", first, second); err != nil {
			t.Error(err)
		}
	})
	for _, id := range []string{first, second} {
		call("PUT", "/v1/resources/"+id, map[string]string{"title": "  Test resource  ", "url": "https://example.com/"}, 200)
	}
	call("PUT", "/v1/resources/"+first, map[string]string{"title": "Updated", "url": "https://example.com/updated"}, 200)
	call("PUT", "/v1/resources/"+first, map[string]string{"title": "Invalid", "url": "file:///tmp/file"}, 400)
	var resources pb.ListResourcesResponse
	assertResources := func(body []byte) {
		t.Helper()
		if err := protojson.Unmarshal(body, &resources); err != nil {
			t.Fatal(err)
		}
		found := 0
		for _, resource := range resources.Resources {
			if resource.Id == first {
				found++
				if resource.Title != "Updated" {
					t.Fatalf("upsert not persisted: %v", resource)
				}
			}
			if resource.Id == second {
				found++
			}
		}
		if found != 2 {
			t.Fatalf("own resources missing or duplicated: %v", &resources)
		}
	}
	assertResources(call("GET", "/v1/resources", nil, 200))
	input := map[string]any{"requestKey": key, "title": "  Go 공부  ", "resourceIds": []string{second, first}}
	var created pb.CreateListResponse
	if err := protojson.Unmarshal(call("POST", "/v1/lists", input, 200), &created); err != nil {
		t.Fatal(err)
	}
	list := created.GetReadingList()
	if list.GetId() == "" || list.Title != "Go 공부" || len(list.Items) != 2 || list.Items[0].Resource.Id != second {
		t.Fatalf("created=%v", list)
	}
	// Catalog edits must not rewrite resource snapshots in an existing list.
	call("PUT", "/v1/resources/"+second, map[string]string{"title": "Changed later", "url": "https://example.com/later"}, 200)
	var snapshot pb.GetListResponse
	if err := protojson.Unmarshal(call("GET", "/v1/lists/"+list.Id, nil, 200), &snapshot); err != nil || snapshot.GetReadingList().GetItems()[0].GetResource().GetTitle() != "Test resource" {
		t.Fatalf("snapshot changed=%v error=%v", &snapshot, err)
	}
	var replay pb.CreateListResponse
	if err := protojson.Unmarshal(call("POST", "/v1/lists", input, 200), &replay); err != nil || replay.GetReadingList().GetId() != list.Id {
		t.Fatalf("replay=%v err=%v", &replay, err)
	}
	input["title"] = "different"
	call("POST", "/v1/lists", input, 409)
	input["requestKey"], input["resourceIds"] = key+"-bad", []string{"missing"}
	call("POST", "/v1/lists", input, 400)
	for _, isRead := range []bool{true, true, false, false, true} {
		call("PATCH", "/v1/lists/"+list.Id+"/items/"+second, map[string]bool{"isRead": isRead}, 200)
	}
	var dashboard struct {
		Catalog     json.RawMessage `json:"catalog"`
		ReadingList json.RawMessage `json:"readingList"`
	}
	if err := json.Unmarshal(call("GET", "/v1/dashboard/"+list.Id, nil, 200), &dashboard); err != nil {
		t.Fatal(err)
	}
	var saved pb.ReadingList
	if err := protojson.Unmarshal(dashboard.ReadingList, &saved); err != nil || saved.Id != list.Id || !saved.Items[0].IsRead {
		t.Fatalf("dashboard=%s err=%v", dashboard.ReadingList, err)
	}
	assertResources(dashboard.Catalog)
	call("GET", "/v1/lists/missing", nil, 404)
	call("GET", "/v1/dashboard/missing", nil, 404)
	// Dependency failure must not break reads of already-saved snapshots.
	catalogApp.RequireStop()
	catalogStopped = true
	call("GET", "/v1/lists/"+list.Id, nil, 200)
	input["requestKey"], input["resourceIds"] = key+"-down", []string{first}
	call("POST", "/v1/lists", input, 503)
	call("GET", "/v1/dashboard/"+list.Id, nil, 503)
}
