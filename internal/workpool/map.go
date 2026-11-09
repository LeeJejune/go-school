package workpool

import (
	"context"
	"errors"
	"sync"
)

// Map bounds concurrency and preserves input order. fn must honor ctx when it
// blocks; cancellation cannot forcibly interrupt an arbitrary Go function.
func Map[T, R any](ctx context.Context, values []T, workers int, fn func(context.Context, T) (R, error)) ([]R, error) {
	if workers < 1 {
		return nil, errors.New("workers must be positive")
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	results := make([]R, len(values))
	if len(values) == 0 {
		return results, nil
	}
	workers = min(workers, len(values))
	jobs := make(chan int)
	var wg sync.WaitGroup
	var once sync.Once
	var firstErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(jobs) // The producer owns closing the channel.
		for i := range values {
			select {
			case jobs <- i:
			case <-ctx.Done():
				return
			}
		}
	}()
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case i, ok := <-jobs:
					if !ok {
						return
					}
					result, err := fn(ctx, values[i])
					if err != nil {
						once.Do(func() { firstErr = err; cancel() })
						return
					}
					results[i] = result // Each input index has exactly one writer.
				}
			}
		}()
	}
	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
