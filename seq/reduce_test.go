package seq_test

import (
	"fmt"
	"iter"
	"slices"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pseudomuto/go/seq"
)

func TestReduce(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []int
		init int
		want int
	}{
		{name: "empty returns init untouched", in: nil, init: 42, want: 42},
		{name: "single value", in: []int{5}, init: 0, want: 5},
		{name: "sums several values", in: []int{1, 2, 3}, init: 0, want: 6},
		{name: "init takes part in the result", in: []int{1, 2, 3}, init: 10, want: 16},
		{name: "negative values", in: []int{-1, -2, 5}, init: 0, want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := seq.Reduce(slices.Values(tt.in), tt.init, func(acc, v int) int { return acc + v })
			require.Equal(t, tt.want, got)
		})
	}
}

func TestReduceFoldsLeftToRight(t *testing.T) {
	t.Parallel()

	t.Run("subtraction is left associative", func(t *testing.T) {
		t.Parallel()

		// ((100-1)-2)-3 == 94. Folding right to left would give
		// 100-(1-(2-3)) == 98, so this pins the direction.
		got := seq.Reduce(slices.Values([]int{1, 2, 3}), 100, func(acc, v int) int { return acc - v })

		require.Equal(t, 94, got)
	})

	t.Run("concatenation keeps source order", func(t *testing.T) {
		t.Parallel()

		got := seq.Reduce(slices.Values([]string{"a", "b", "c"}), "", func(acc, v string) string {
			return acc + v
		})

		require.Equal(t, "abc", got)
	})
}

func TestReduceChangesAccumulatorType(t *testing.T) {
	t.Parallel()

	t.Run("ints into a string", func(t *testing.T) {
		t.Parallel()

		got := seq.Reduce(slices.Values([]int{1, 2, 3}), "", func(acc string, v int) string {
			return acc + strconv.Itoa(v)
		})

		require.Equal(t, "123", got)
	})

	t.Run("strings into a count map", func(t *testing.T) {
		t.Parallel()

		got := seq.Reduce(slices.Values([]string{"a", "b", "a"}), map[string]int{},
			func(acc map[string]int, v string) map[string]int {
				acc[v]++
				return acc
			})

		require.Equal(t, map[string]int{"a": 2, "b": 1}, got)
	})
}

func TestReduceDrainsTheSequence(t *testing.T) {
	t.Parallel()

	pulled, calls := 0, 0

	var src iter.Seq[int] = func(yield func(int) bool) {
		for v := range 5 {
			pulled++
			if !yield(v) {
				return
			}
		}
	}

	got := seq.Reduce(src, 0, func(acc, v int) int { calls++; return acc + v })

	require.Equal(t, 10, got)
	require.Equal(t, 5, pulled, "Reduce stopped short of the end")
	require.Equal(t, 5, calls, "fn should run exactly once per value")
}

func ExampleReduce() {
	nums := slices.Values([]int{1, 2, 3, 4})

	sum := seq.Reduce(nums, 0, func(acc, v int) int { return acc + v })

	fmt.Println(sum)
	// Output: 10
}

func ExampleReduce_changingType() {
	words := slices.Values([]string{"go", "is", "fun"})

	// The accumulator is an int while the values are strings. Go infers both
	// from init and seq, so no explicit type arguments are needed.
	total := seq.Reduce(words, 0, func(acc int, w string) int { return acc + len(w) })

	fmt.Println(total)
	// Output: 7
}
