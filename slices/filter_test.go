package slices_test

import (
	"fmt"
	stdslices "slices"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pseudomuto/go/slices"
)

func isEven(v int) bool { return v%2 == 0 }

func TestFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []int
		want []int
	}{
		{name: "nil", in: nil, want: nil},
		{name: "empty", in: []int{}, want: nil},
		{name: "nothing matches", in: []int{1, 3, 5}, want: nil},
		{name: "everything matches", in: []int{2, 4, 6}, want: []int{2, 4, 6}},
		{name: "some match", in: []int{1, 2, 3, 4}, want: []int{2, 4}},
		{name: "match at both ends", in: []int{2, 3, 4}, want: []int{2, 4}},
		{name: "duplicates are kept", in: []int{2, 2, 3}, want: []int{2, 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, slices.Filter(tt.in, isEven))
		})
	}
}

func TestFilterLeavesInputAloneAndDoesNotAlias(t *testing.T) {
	t.Parallel()

	in := []int{1, 2, 3, 4}

	got := slices.Filter(in, isEven)
	require.Equal(t, []int{2, 4}, got)
	require.Equal(t, []int{1, 2, 3, 4}, in, "Filter modified the slice it was given")

	got[0] = 99
	require.Equal(t, []int{1, 2, 3, 4}, in, "the result shares a backing array with the input")
}

func TestReject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []int
		want []int
	}{
		{name: "nil", in: nil, want: nil},
		{name: "empty", in: []int{}, want: nil},
		{name: "nothing rejected", in: []int{1, 3, 5}, want: []int{1, 3, 5}},
		{name: "everything rejected", in: []int{2, 4, 6}, want: nil},
		{name: "some rejected", in: []int{1, 2, 3, 4}, want: []int{1, 3}},
		{name: "rejection at both ends", in: []int{2, 3, 4}, want: []int{3}},
		{name: "duplicates are kept", in: []int{3, 3, 2}, want: []int{3, 3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, slices.Reject(tt.in, isEven))
		})
	}
}

func TestRejectPartitionsWithFilter(t *testing.T) {
	t.Parallel()

	in := []int{1, 2, 3, 4, 5, 6, 7}

	kept := slices.Filter(in, isEven)
	dropped := slices.Reject(in, isEven)

	require.Equal(t, []int{2, 4, 6}, kept)
	require.Equal(t, []int{1, 3, 5, 7}, dropped)

	union := stdslices.Concat(kept, dropped)
	stdslices.Sort(union)
	require.Equal(t, in, union, "Filter and Reject do not cover the input exactly once")
}

func ExampleFilter() {
	fmt.Println(slices.Filter([]int{1, 2, 3, 4, 5, 6}, func(v int) bool { return v%2 == 0 }))
	// Output: [2 4 6]
}

func ExampleReject() {
	fmt.Println(slices.Reject([]int{1, 2, 3, 4, 5, 6}, func(v int) bool { return v%2 == 0 }))
	// Output: [1 3 5]
}
