package mysql

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/LeeJejune/go-school/go-practice/internal/catalog"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCatalogUpsertAndBatchOrder(t *testing.T) {
	dsn := os.Getenv("MYSQL_TEST_DSN")
	if dsn == "" {
		t.Skip("MYSQL_TEST_DSN is not set")
	}
	db, err := gorm.Open(gormmysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Close() })
	repo := NewCatalogRepository(db)
	key := fmt.Sprintf("catalog-test-%d", time.Now().UnixNano())
	first, second := key+"-one", key+"-two"
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := db.WithContext(ctx).Where("id IN ?", []string{first, second}).Delete(&ResourceEntity{}).Error; err != nil {
			t.Error(err)
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, id := range []string{first, second} {
		if _, err := repo.Upsert(ctx, catalog.Resource{ID: id, Title: "Initial", URL: "https://example.com/"}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := repo.Upsert(ctx, catalog.Resource{ID: first, Title: "Changed", URL: "https://example.com/changed"}); err != nil {
		t.Fatal(err)
	}
	got, err := repo.Batch(ctx, []string{second, first, second})
	if err != nil || len(got) != 3 || got[0].ID != second || got[1].Title != "Changed" || got[1].URL != "https://example.com/changed" || got[2].ID != second {
		t.Fatalf("resources=%+v error=%v", got, err)
	}
	if _, err := repo.Batch(ctx, []string{first, key + "-missing"}); !errors.Is(err, catalog.ErrNotFound) {
		t.Fatalf("missing=%v", err)
	}
	resources, err := repo.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, resource := range resources {
		if resource.ID == first {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("upsert created %d copies", count)
	}
}
