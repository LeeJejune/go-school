package workpool

import (
	"context"
	"errors"
	"reflect"
	"sync/atomic"
	"testing"
)

func TestOrderAndWorkerBound(t *testing.T) {
	var active, peak atomic.Int32
	got, err := Map(context.Background(), []int{4, 2, 3, 1, 5}, 2, func(_ context.Context, n int) (int, error) {
		now := active.Add(1)
		defer active.Add(-1)
		for old := peak.Load(); now > old; old = peak.Load() {
			if peak.CompareAndSwap(old, now) {
				break
			}
		}
		return n * n, nil
	})
	if err != nil || !reflect.DeepEqual(got, []int{16, 4, 9, 1, 25}) {
		t.Fatalf("result=%v err=%v", got, err)
	}
	if peak.Load() > 2 {
		t.Fatal("worker limit exceeded")
	}
}

func TestFailureAndCancellation(t *testing.T) {
	want := errors.New("lookup failed")
	_, err := Map(context.Background(), []int{1, 2, 3}, 2, func(ctx context.Context, n int) (int, error) {
		if n == 1 {
			return 0, want
		}
		<-ctx.Done()
		return 0, ctx.Err()
	})
	if !errors.Is(err, want) {
		t.Fatalf("got %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = Map(ctx, []int{1}, 1, func(context.Context, int) (int, error) { t.Fatal("called after cancellation"); return 0, nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
}
