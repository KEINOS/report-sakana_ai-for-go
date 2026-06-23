package topk

import (
	"fmt"
	"slices"
	"testing"
)

func TestTopKFrequent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		words []string
		k     int
		want  []string
	}{
		{"basic", []string{"go", "rust", "go", "zig", "rust", "go"}, 2, []string{"go", "rust"}},
		{"lexical tie", []string{"b", "a", "c", "b", "a", "c"}, 2, []string{"a", "b"}},
		{"unicode", []string{"猫", "犬", "猫", "鳥", "犬"}, 3, []string{"犬", "猫", "鳥"}},
		{"k exceeds unique", []string{"b", "a", "b"}, 10, []string{"b", "a"}},
		{"zero", []string{"a"}, 0, []string{}},
		{"negative", []string{"a"}, -1, []string{}},
		{"empty", nil, 3, []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			before := slices.Clone(tt.words)
			got := TopKFrequent(tt.words, tt.k)
			if !slices.Equal(got, tt.want) {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
			if !slices.Equal(tt.words, before) {
				t.Fatalf("input mutated: got %q, want %q", tt.words, before)
			}
			if got == nil {
				t.Fatal("result must be non-nil")
			}
		})
	}
}

func BenchmarkTopKFrequent(b *testing.B) {
	words := make([]string, 100_000)
	for i := range words {
		words[i] = fmt.Sprintf("word-%05d", (i*7919)%10_000)
	}
	b.ReportAllocs()
	for b.Loop() {
		_ = TopKFrequent(words, 20)
	}
}

