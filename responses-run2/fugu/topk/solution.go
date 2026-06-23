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

	h := make(wordHeap, 0, k)
	for word, count := range counts {
		item := wordCount{word: word, count: count}
		if h.Len() < k {
			heap.Push(&h, item)
		} else if better(item, h[0]) {
			h[0] = item
			heap.Fix(&h, 0)
		}
	}

	sort.Slice(h, func(i, j int) bool {
		return better(h[i], h[j])
	})

	res := make([]string, len(h))
	for i, item := range h {
		res[i] = item.word
	}
	return res
}

type wordCount struct {
	word  string
	count int
}

func better(a, b wordCount) bool {
	if a.count != b.count {
		return a.count > b.count
	}
	return a.word < b.word
}

type wordHeap []wordCount

func (h wordHeap) Len() int { return len(h) }

func (h wordHeap) Less(i, j int) bool {
	if h[i].count != h[j].count {
		return h[i].count < h[j].count
	}
	return h[i].word > h[j].word
}

func (h wordHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *wordHeap) Push(x any) {
	*h = append(*h, x.(wordCount))
}

func (h *wordHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}
