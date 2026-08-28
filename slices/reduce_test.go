package slices_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pseudomuto/go/slices"
)

func TestReduce(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []int
		init int
		want int
	}{
		{name: "nil returns init untouched", in: nil, init: 42, want: 42},
		{name: "empty returns init untouched", in: []int{}, init: 42, want: 42},
		{name: "single value", in: []int{5}, init: 0, want: 5},
		{name: "sums several values", in: []int{1, 2, 3}, init: 0, want: 6},
		{name: "init takes part in the result", in: []int{1, 2, 3}, init: 10, want: 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := slices.Reduce(tt.in, tt.init, func(acc, v int) int { return acc + v })
			require.Equal(t, tt.want, got)
		})
	}
}

func TestReduceFoldsLeftToRight(t *testing.T) {
	t.Parallel()

	// ((100-1)-2)-3 == 94. Right to left would give 100-(1-(2-3)) == 98.
	got := slices.Reduce([]int{1, 2, 3}, 100, func(acc, v int) int { return acc - v })

	require.Equal(t, 94, got)
}

func TestReduceChangesAccumulatorType(t *testing.T) {
	t.Parallel()

	got := slices.Reduce([]string{"go", "is", "fun"}, 0, func(acc int, w string) int {
		return acc + len(w)
	})

	require.Equal(t, 7, got)
}

func TestReduceLeavesInputAlone(t *testing.T) {
	t.Parallel()

	in := []int{1, 2, 3}

	got := slices.Reduce(in, 0, func(acc, v int) int { return acc + v })

	require.Equal(t, 6, got)
	require.Equal(t, []int{1, 2, 3}, in, "Reduce modified the slice it was given")
}

func ExampleReduce() {
	sum := slices.Reduce([]int{1, 2, 3, 4}, 0, func(acc, v int) int { return acc + v })

	fmt.Println(sum)
	// Output: 10
}
