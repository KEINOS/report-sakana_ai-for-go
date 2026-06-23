package pmap

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// ParallelMapOrdered applies fn to each element of input concurrently, using at
// most workers goroutines, and returns the results in the original input order.
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

	results := make([]int, len(input))
	sem := make(chan struct{}, workers)

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		firstErr error
	)

	setErr := func(err error) {
		mu.Lock()
		if firstErr == nil {
			firstErr = err
			cancel()
		}
		mu.Unlock()
	}

dispatch:
	for i, v := range input {
		select {
		case <-ctx.Done():
			setErr(context.Cause(ctx))
			break dispatch
		case sem <- struct{}{}:
		}

		wg.Add(1)
		go func(i, v int) {
			defer wg.Done()
			defer func() { <-sem }()
			defer func() {
				if r := recover(); r != nil {
					setErr(fmt.Errorf("pmap: panic in fn for input[%d]: %v", i, r))
				}
			}()

			select {
			case <-ctx.Done():
				return
			default:
			}

			res, err := fn(ctx, v)
			if err != nil {
				setErr(err)
				return
			}
			results[i] = res
		}(i, v)
	}

	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}
	return results, nil
}
