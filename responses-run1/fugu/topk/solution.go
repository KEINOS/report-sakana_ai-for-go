package topk

import "container/heap"

func TopKFrequent(words []string, k int) []string {
	if k <= 0 || len(words) == 0 {
		return []string{}
	}

	counts := make(map[string]int, len(words))
	for _, word := range words {
		counts[word]++
	}

	h := make(wordHeap, 0, min(k, len(counts)))
	for word, count := range counts {
		item := wordCount{word: word, count: count}
		if h.Len() < k {
			heap.Push(&h, item)
		} else if better(item, h[0]) {
			h[0] = item
			heap.Fix(&h, 0)
		}
	}

	res := make([]string, h.Len())
	for i := len(res) - 1; i >= 0; i-- {
		res[i] = heap.Pop(&h).(wordCount).word
	}
	return res
}

type wordCount struct {
	word  string
	count int
}

type wordHeap []wordCount

func (h wordHeap) Len() int { return len(h) }

func (h wordHeap) Less(i, j int) bool {
	return worse(h[i], h[j])
}

func (h wordHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *wordHeap) Push(x any) {
	*h = append(*h, x.(wordCount))
}

func (h *wordHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

func better(a, b wordCount) bool {
	if a.count != b.count {
		return a.count > b.count
	}
	return a.word < b.word
}

func worse(a, b wordCount) bool {
	if a.count != b.count {
		return a.count < b.count
	}
	return a.word > b.word
}
