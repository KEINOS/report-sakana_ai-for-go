Target Go 1.26. Return only one complete Go source file using package topk.

Implement:
    func TopKFrequent(words []string, k int) []string

Requirements:
- Return at most k unique words, ordered by descending frequency.
- Resolve equal frequencies by ascending lexical order.
- k <= 0 or empty input returns a non-nil empty slice.
- Do not mutate words.
- Optimize for large inputs and small k: avoid sorting all input occurrences.
- Use only the standard library.
- Produce concise, idiomatic, maintainable Go 1.26 code.
