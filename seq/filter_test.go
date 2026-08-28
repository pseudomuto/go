package seq_test

import (
	"fmt"
	"iter"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pseudomuto/go/seq"
)

func isEven(v int) bool { return v%2 == 0 }

func TestFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []int
		want []int
	}{
		{name: "empty", in: nil, want: nil},
		{name: "nothing matches", in: []int{1, 3, 5}, want: nil},
		{name: "everything matches", in: []int{2, 4, 6}, want: []int{2, 4, 6}},
		{name: "some match", in: []int{1, 2, 3, 4}, want: []int{2, 4}},
		{name: "match at both ends", in: []int{2, 3, 4}, want: []int{2, 4}},
		{name: "duplicates are kept", in: []int{2, 2, 3}, want: []int{2, 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := slices.Collect(seq.Filter(slices.Values(tt.in), isEven))
			require.Equal(t, tt.want, got)
		})
	}
}

func TestFilterIsLazy(t *testing.T) {
	t.Parallel()

	calls := 0
	evens := seq.Filter(slices.Values([]int{1, 2, 3, 4}), func(v int) bool {
		calls++
		return isEven(v)
	})

	require.Zero(t, calls, "pred ran before the sequence was consumed")
	require.Equal(t, []int{2, 4}, slices.Collect(evens))
	require.Equal(t, 4, calls, "pred should run exactly once per source value")
}

func TestFilterStopsEarly(t *testing.T) {
	t.Parallel()

	var pulled, tested int

	var src iter.Seq[int] = func(yield func(int) bool) {
		for v := range 10 {
			pulled++
			if !yield(v) {
				return
			}
		}
	}

	var got []int
	for v := range seq.Filter(src, func(v int) bool { tested++; return isEven(v) }) {
		got = append(got, v)
		if len(got) == 2 {
			break
		}
	}

	require.Equal(t, []int{0, 2}, got)
	require.Equal(t, 3, pulled, "source kept producing after the consumer stopped")
	require.Equal(t, 3, tested, "pred ran for values the source never had to produce")
}

func TestReject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []int
		want []int
	}{
		{name: "empty", in: nil, want: nil},
		{name: "nothing rejected", in: []int{1, 3, 5}, want: []int{1, 3, 5}},
		{name: "everything rejected", in: []int{2, 4, 6}, want: nil},
		{name: "some rejected", in: []int{1, 2, 3, 4}, want: []int{1, 3}},
		{name: "rejection at both ends", in: []int{2, 3, 4}, want: []int{3}},
		{name: "duplicates are kept", in: []int{3, 3, 2}, want: []int{3, 3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := slices.Collect(seq.Reject(slices.Values(tt.in), isEven))
			require.Equal(t, tt.want, got)
		})
	}
}

func TestRejectPartitionsWithFilter(t *testing.T) {
	t.Parallel()

	in := []int{1, 2, 3, 4, 5, 6, 7}

	kept := slices.Collect(seq.Filter(slices.Values(in), isEven))
	dropped := slices.Collect(seq.Reject(slices.Values(in), isEven))

	require.Equal(t, []int{2, 4, 6}, kept)
	require.Equal(t, []int{1, 3, 5, 7}, dropped)

	union := slices.Concat(kept, dropped)
	slices.Sort(union)
	require.Equal(t, in, union, "Filter and Reject do not cover the input exactly once")
}

func TestRejectIsLazy(t *testing.T) {
	t.Parallel()

	calls := 0
	odds := seq.Reject(slices.Values([]int{1, 2, 3, 4}), func(v int) bool {
		calls++
		return isEven(v)
	})

	require.Zero(t, calls, "pred ran before the sequence was consumed")
	require.Equal(t, []int{1, 3}, slices.Collect(odds))
	require.Equal(t, 4, calls, "pred should run exactly once per source value")
}

func TestRejectStopsEarly(t *testing.T) {
	t.Parallel()

	pulled := 0

	var src iter.Seq[int] = func(yield func(int) bool) {
		for v := range 10 {
			pulled++
			if !yield(v) {
				return
			}
		}
	}

	var got []int
	for v := range seq.Reject(src, isEven) {
		got = append(got, v)
		if len(got) == 2 {
			break
		}
	}

	require.Equal(t, []int{1, 3}, got)
	require.Equal(t, 4, pulled, "source kept producing after the consumer stopped")
}

func ExampleFilter() {
	nums := slices.Values([]int{1, 2, 3, 4, 5, 6})

	fmt.Println(slices.Collect(seq.Filter(nums, func(v int) bool { return v%2 == 0 })))
	// Output: [2 4 6]
}

func ExampleReject() {
	nums := slices.Values([]int{1, 2, 3, 4, 5, 6})

	fmt.Println(slices.Collect(seq.Reject(nums, func(v int) bool { return v%2 == 0 })))
	// Output: [1 3 5]
}
