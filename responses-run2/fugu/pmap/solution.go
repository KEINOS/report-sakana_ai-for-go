package pmap

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

// ParallelMapOrdered applies fn to each element of input using at most workers
// concurrent goroutines, returning the results in input order.
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

	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	results := make([]int, len(input))

	var (
		next     atomic.Int64
		firstErr error
		errOnce  sync.Once
	)

	n := workers
	if n > len(input) {
		n = len(input)
	}

	var wg sync.WaitGroup
	wg.Add(n)
	for range n {
		go func() {
			defer wg.Done()
			for {
				i := int(next.Add(1)) - 1
				if i >= len(input) {
					return
				}
				if ctx.Err() != nil {
					return
				}
				v, err := call(ctx, fn, input[i])
				if err != nil {
					errOnce.Do(func() {
						firstErr = err
						cancel(err)
					})
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
	if err := context.Cause(ctx); err != nil {
		return nil, err
	}
	return results, nil
}

// call invokes fn while recovering any panic and converting it to an error.
func call(
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
