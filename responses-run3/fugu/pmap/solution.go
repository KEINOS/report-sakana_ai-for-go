package pmap

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// ParallelMapOrdered applies fn to every element of input using at most workers
// concurrent goroutines, returning the results in the original input order.
//
// It cancels remaining work on the first error or parent cancellation, waits for
// all started goroutines to finish, and returns nil with an error that preserves
// errors.Is semantics. Panics from fn are converted into descriptive errors.
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
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := make([]int, len(input))

	n := workers
	if n > len(input) {
		n = len(input)
	}

	tasks := make(chan int)
	var (
		once     sync.Once
		firstErr error
		wg       sync.WaitGroup
	)

	setErr := func(err error) {
		once.Do(func() {
			firstErr = err
			cancel()
		})
	}

	wg.Add(n)
	for range n {
		go func() {
			defer wg.Done()
			for i := range tasks {
				if ctx.Err() != nil {
					return
				}
				v, err := safeCall(ctx, fn, input[i])
				if err != nil {
					setErr(err)
					return
				}
				results[i] = v // unique index per task: no race
			}
		}()
	}

feed:
	for i := range input {
		select {
		case <-ctx.Done():
			break feed
		case tasks <- i:
		}
	}
	close(tasks)
	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// safeCall invokes fn, converting any panic into a descriptive error.
func safeCall(
	ctx context.Context,
	fn func(context.Context, int) (int, error),
	x int,
) (v int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("pmap: panic in fn: %v", r)
		}
	}()
	return fn(ctx, x)
}
