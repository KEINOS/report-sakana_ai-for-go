package pmap

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestParallelMapOrdered(t *testing.T) {
	t.Parallel()
	input := []int{5, 4, 3, 2, 1}
	got, err := ParallelMapOrdered(t.Context(), input, 3, func(_ context.Context, n int) (int, error) {
		time.Sleep(time.Duration(n) * time.Millisecond)
		return n * n, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []int{25, 16, 9, 4, 1}; !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestParallelMapValidation(t *testing.T) {
	t.Parallel()
	fn := func(_ context.Context, n int) (int, error) { return n, nil }
	if _, err := ParallelMapOrdered(t.Context(), []int{1}, 0, fn); err == nil {
		t.Fatal("workers=0 must return an error")
	}
	if _, err := ParallelMapOrdered(t.Context(), []int{1}, 1, nil); err == nil {
		t.Fatal("nil function must return an error")
	}
	got, err := ParallelMapOrdered(t.Context(), nil, 2, fn)
	if err != nil || got == nil || len(got) != 0 {
		t.Fatalf("empty input: got=%v err=%v", got, err)
	}
}

func TestParallelMapCancelsAndWaits(t *testing.T) {
	t.Parallel()
	sentinel := errors.New("stop")
	var active atomic.Int64
	var completed atomic.Int64
	input := make([]int, 100)
	for i := range input {
		input[i] = i
	}
	_, err := ParallelMapOrdered(t.Context(), input, 8, func(ctx context.Context, n int) (int, error) {
		active.Add(1)
		defer active.Add(-1)
		if n == 3 {
			return 0, sentinel
		}
		select {
		case <-ctx.Done():
			completed.Add(1)
			return 0, ctx.Err()
		case <-time.After(250 * time.Millisecond):
			completed.Add(1)
			return n, nil
		}
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("got error %v, want wrapping %v", err, sentinel)
	}
	if active.Load() != 0 {
		t.Fatalf("returned while %d workers were active", active.Load())
	}
	if completed.Load() >= 95 {
		t.Fatalf("cancellation ineffective; completed=%d", completed.Load())
	}
}

func TestParallelMapPanicBecomesError(t *testing.T) {
	t.Parallel()
	_, err := ParallelMapOrdered(t.Context(), []int{1, 2, 3}, 2, func(_ context.Context, n int) (int, error) {
		if n == 2 {
			panic("boom")
		}
		return n, nil
	})
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("panic must become descriptive error, got %v", err)
	}
}

func TestParallelMapParentCancellation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := ParallelMapOrdered(ctx, []int{1, 2, 3}, 2, func(ctx context.Context, n int) (int, error) {
		return n, ctx.Err()
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want context.Canceled", err)
	}
}

func TestParallelMapDoesNotLeakObviousGoroutines(t *testing.T) {
	before := runtime.NumGoroutine()
	for range 20 {
		_, _ = ParallelMapOrdered(t.Context(), []int{1, 2, 3, 4}, 4, func(_ context.Context, n int) (int, error) {
			if n == 2 {
				return 0, fmt.Errorf("item %d: failure", n)
			}
			return n, nil
		})
	}
	runtime.GC()
	time.Sleep(50 * time.Millisecond)
	after := runtime.NumGoroutine()
	if delta := after - before; delta > 4 {
		t.Fatalf("possible goroutine leak: before=%d after=%d delta=%d", before, after, delta)
	}
}

func BenchmarkParallelMapOrdered(b *testing.B) {
	input := make([]int, 10_000)
	fn := func(_ context.Context, n int) (int, error) { return n * 2, nil }
	b.ReportAllocs()
	for b.Loop() {
		_, err := ParallelMapOrdered(context.Background(), input, 8, fn)
		if err != nil {
			b.Fatal(err)
		}
	}
}

