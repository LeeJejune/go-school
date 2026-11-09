package mysqlstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/LeeJejune/go-school/internal/reading"
)

// Run only against the dedicated local learning database after applying 001_init.
// No table truncation: each test uses its own request key and cleans its own rows.
func TestMySQLTransactionReplayAndReadState(t *testing.T) {
	dsn := os.Getenv("MYSQL_TEST_DSN")
	if dsn == "" {
		t.Skip("MYSQL_TEST_DSN is not set")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}
	store := New(db)
	key := fmt.Sprintf("test-%d", time.Now().UnixNano())
	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_, _ = db.ExecContext(cleanup, "DELETE FROM reading_lists WHERE request_key IN (?, ?)", key, key+"-bad")
	})
	params := reading.CreateParams{RequestKey: key, Fingerprint: "same", Title: "Go", Resources: []reading.Resource{{ID: "go-tour", Title: "Tour", URL: "https://go.dev/tour/"}}}
	var wg sync.WaitGroup
	ids := make(chan string, 6)
	errs := make(chan error, 6)
	for range 6 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			list, err := store.Create(ctx, params)
			if err != nil {
				errs <- err
				return
			}
			ids <- list.ID
		}()
	}
	wg.Wait()
	close(ids)
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
	var id string
	for got := range ids {
		if id != "" && got != id {
			t.Fatal("same request created multiple lists")
		}
		id = got
	}
	params.Fingerprint = "different"
	if _, err := store.Create(ctx, params); !errors.Is(err, reading.ErrConflict) {
		t.Fatalf("conflict=%v", err)
	}
	for range 2 {
		if err := store.SetRead(ctx, id, "go-tour", true); err != nil {
			t.Fatal(err)
		}
	}
	list, err := store.Get(ctx, id)
	if err != nil || !list.Items[0].IsRead {
		t.Fatalf("list=%+v err=%v", list, err)
	}
	bad := params
	bad.RequestKey = key + "-bad"
	bad.Resources = append(bad.Resources, bad.Resources[0])
	if _, err := store.Create(ctx, bad); err == nil {
		t.Fatal("expected item uniqueness failure")
	}
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM reading_lists WHERE request_key = ?", bad.RequestKey).Scan(&count); err != nil || count != 0 {
		t.Fatalf("rollback count=%d err=%v", count, err)
	}
}
