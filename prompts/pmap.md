Target Go 1.26. Return only one complete Go source file using package pmap.

Implement:
    func ParallelMapOrdered(
        ctx context.Context,
        input []int,
        workers int,
        fn func(context.Context, int) (int, error),
    ) ([]int, error)

Requirements:
- Preserve input order in successful output.
- Run no more than workers calls concurrently.
- workers <= 0 and nil fn return descriptive errors.
- Empty input returns a non-nil empty slice.
- On the first worker error or parent cancellation, cancel remaining work, wait for all started goroutines, and return nil plus an error preserving errors.Is.
- Convert a panic from fn into a descriptive error containing the panic value; never let it crash the process.
- Do not leak goroutines or race.
- Optimize for low overhead.
- Use only the standard library.
- Produce concise, idiomatic, maintainable Go 1.26 code.
