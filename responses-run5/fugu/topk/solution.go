package topk

import (
	"container/heap"
	"sort"
)

func TopKFrequent(words []string, k int) []string {
	if k <= 0 || len(words) == 0 {
		return []string{}
	}

	counts := make(map[string]int, len(words))
	for _, word := range words {
		counts[word]++
	}

	if k > len(counts) {
		k = len(counts)
	}

	h := make(entryHeap, 0, k)
	for word, count := range counts {
		e := entry{word: word, count: count}
		if len(h) < k {
			heap.Push(&h, e)
			continue
		}
		if better(e, h[0]) {
			h[0] = e
			heap.Fix(&h, 0)
		}
	}

	sort.Slice(h, func(i, j int) bool {
		return better(h[i], h[j])
	})

	result := make([]string, len(h))
	for i, e := range h {
		result[i] = e.word
	}
	return result
}

type entry struct {
	word  string
	count int
}

func better(a, b entry) bool {
	if a.count != b.count {
		return a.count > b.count
	}
	return a.word < b.word
}

func worse(a, b entry) bool {
	if a.count != b.count {
		return a.count < b.count
	}
	return a.word > b.word
}

type entryHeap []entry

func (h entryHeap) Len() int { return len(h) }

func (h entryHeap) Less(i, j int) bool {
	return worse(h[i], h[j])
}

func (h entryHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *entryHeap) Push(x any) {
	*h = append(*h, x.(entry))
}

func (h *entryHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}
