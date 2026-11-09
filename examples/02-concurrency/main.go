package main

import (
	"context"
	"fmt"
	"time"

	"github.com/LeeJejune/go-school/internal/workpool"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	results, err := workpool.Map(ctx, []int{1, 2, 3, 4, 5}, 2, func(ctx context.Context, value int) (int, error) {
		timer := time.NewTimer(10 * time.Millisecond)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-timer.C:
			return value * value, nil
		}
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(results) // [1 4 9 16 25], completion order does not change output order.
}
