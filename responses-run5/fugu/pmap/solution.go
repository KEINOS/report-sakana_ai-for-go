package pmap

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

// ParallelMapOrdered applies fn to each element of input using at most workers
// concurrent goroutines, returning results in the original input order.
func ParallelMapOrdered(
	ctx context.Context,
	input []int,
	workers int,
	fn func(context.Context, int) (int, error),
) ([]int, error) {
	if workers <= 0 {
		return nil, fmt.Errorf("pmap: workers must be positive, got %d", workers)
	}
	if fn == nil {
		return nil, errors.New("pmap: fn must not be nil")
	}
	if len(input) == 0 {
		return []int{}, nil
	}

	if workers > len(input) {
		workers = len(input)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := make([]int, len(input))

	var (
		idx      atomic.Int64
		firstErr error
		once     sync.Once
	)
	setErr := func(err error) {
		once.Do(func() {
			firstErr = err
			cancel()
		})
	}

	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for {
				if ctx.Err() != nil {
					return
				}
				i := int(idx.Add(1)) - 1
				if i >= len(input) {
					return
				}
				v, err := safeCall(ctx, fn, input[i])
				if err != nil {
					setErr(err)
					return
				}
				results[i] = v
			}
		}()
	}
	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("pmap: context error: %w", err)
	}
	return results, nil
}

// safeCall invokes fn while converting any panic into a descriptive error.
func safeCall(
	ctx context.Context,
	fn func(context.Context, int) (int, error),
	v int,
) (out int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("pmap: panic in fn: %v", r)
		}
	}()
	return fn(ctx, v)
}
