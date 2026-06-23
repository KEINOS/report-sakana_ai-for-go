package pmap

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// ParallelMapOrdered applies fn to each element of input concurrently using at
// most workers goroutines, returning the results in the original input order.
func ParallelMapOrdered(
	ctx context.Context,
	input []int,
	workers int,
	fn func(context.Context, int) (int, error),
) ([]int, error) {
	if workers <= 0 {
		return nil, fmt.Errorf("pmap: workers must be > 0, got %d", workers)
	}
	if fn == nil {
		return nil, errors.New("pmap: fn must not be nil")
	}
	if len(input) == 0 {
		return []int{}, nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if workers > len(input) {
		workers = len(input)
	}

	results := make([]int, len(input))
	jobs := make(chan int)

	var (
		wg       sync.WaitGroup
		errOnce  sync.Once
		firstErr error
	)

	setErr := func(err error) {
		errOnce.Do(func() {
			firstErr = err
			cancel()
		})
	}

	worker := func() {
		defer wg.Done()
		for i := range jobs {
			select {
			case <-ctx.Done():
				return
			default:
			}
			v, err := safeCall(ctx, fn, input[i])
			if err != nil {
				setErr(err)
				return
			}
			results[i] = v
		}
	}

	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go worker()
	}

feed:
	for i := range input {
		select {
		case <-ctx.Done():
			break feed
		case jobs <- i:
		}
	}
	close(jobs)

	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("pmap: canceled: %w", err)
	}
	return results, nil
}

// safeCall invokes fn, converting any panic into a descriptive error.
func safeCall(
	ctx context.Context,
	fn func(context.Context, int) (int, error),
	x int,
) (out int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("pmap: panic in fn: %v", r)
		}
	}()
	return fn(ctx, x)
}
