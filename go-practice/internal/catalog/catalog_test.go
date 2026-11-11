package catalog

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type repositoryStub struct {
	Repository
	saved Resource
	calls int
}

func (r *repositoryStub) Upsert(_ context.Context, resource Resource) (Resource, error) {
	r.saved = resource
	r.calls++
	return resource, nil
}

func TestUpsertValidatesBeforeStorage(t *testing.T) {
	repo := &repositoryStub{}
	service := New(repo)
	valid := Resource{ID: "my-resource_1", Title: "  내 자료  ", URL: "https://example.com/path"}
	for _, tc := range []struct{ field, value string }{
		{"id", ""}, {"id", "한글"}, {"id", "with/slash"}, {"id", strings.Repeat("a", 65)},
		{"title", "  "}, {"title", strings.Repeat("가", 256)},
		{"url", "relative"}, {"url", "file:///tmp/file"}, {"url", "https://"}, {"url", "https://example.com/" + strings.Repeat("a", 2048)},
	} {
		t.Run(tc.field+"-"+tc.value[:min(len(tc.value), 12)], func(t *testing.T) {
			resource := valid
			switch tc.field {
			case "id":
				resource.ID = tc.value
			case "title":
				resource.Title = tc.value
			case "url":
				resource.URL = tc.value
			}
			if _, err := service.Upsert(context.Background(), resource); !errors.Is(err, ErrInvalidInput) || repo.calls != 0 {
				t.Fatalf("error=%v storage calls=%d", err, repo.calls)
			}
		})
	}
	got, err := service.Upsert(context.Background(), valid)
	if err != nil || got.Title != "내 자료" || repo.saved != got || repo.calls != 1 {
		t.Fatalf("resource=%+v saved=%+v error=%v", got, repo.saved, err)
	}
}

func TestBatchRejectsUnboundedRequests(t *testing.T) {
	service := New(&repositoryStub{})
	for _, ids := range [][]string{nil, make([]string, 21)} {
		if _, err := service.Batch(context.Background(), ids); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("ids=%v error=%v", ids, err)
		}
	}
}
