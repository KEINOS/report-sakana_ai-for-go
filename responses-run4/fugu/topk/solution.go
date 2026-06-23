package topk

import (
	"container/heap"
	"sort"
)

type entry struct {
	word string
	freq int
}

type topKHeap []entry

func (h topKHeap) Len() int { return len(h) }

func (h topKHeap) Less(i, j int) bool {
	if h[i].freq != h[j].freq {
		return h[i].freq < h[j].freq
	}
	return h[i].word > h[j].word
}

func (h topKHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *topKHeap) Push(x any) {
	*h = append(*h, x.(entry))
}

func (h *topKHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

func better(a, b entry) bool {
	if a.freq != b.freq {
		return a.freq > b.freq
	}
	return a.word < b.word
}

func TopKFrequent(words []string, k int) []string {
	if k <= 0 || len(words) == 0 {
		return []string{}
	}

	counts := make(map[string]int)
	for _, word := range words {
		counts[word]++
	}

	if k > len(counts) {
		k = len(counts)
	}

	h := make(topKHeap, 0, k)
	for word, freq := range counts {
		e := entry{word: word, freq: freq}
		if len(h) < k {
			heap.Push(&h, e)
		} else if better(e, h[0]) {
			h[0] = e
			heap.Fix(&h, 0)
		}
	}

	top := make([]entry, len(h))
	copy(top, h)

	sort.Slice(top, func(i, j int) bool {
		return better(top[i], top[j])
	})

	result := make([]string, len(top))
	for i, e := range top {
		result[i] = e.word
	}
	return result
}
