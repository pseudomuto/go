package seq_test

import (
	"errors"
	"fmt"
	"iter"
	"slices"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pseudomuto/go/seq"
)

var errOdd = errors.New("odd value")

type mapped struct {
	val string
	err error
}

func TestMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []int
		want []string
	}{
		{name: "empty", in: nil, want: nil},
		{name: "single value", in: []int{1}, want: []string{"1"}},
		{name: "several values", in: []int{1, 2, 3}, want: []string{"1", "2", "3"}},
		{name: "duplicates are kept", in: []int{1, 1, 2}, want: []string{"1", "1", "2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := slices.Collect(seq.Map(slices.Values(tt.in), strconv.Itoa))
			require.Equal(t, tt.want, got)
		})
	}
}

func TestMapIsLazy(t *testing.T) {
	t.Parallel()

	calls := 0
	double := seq.Map(slices.Values([]int{1, 2, 3}), func(v int) int {
		calls++
		return v * 2
	})

	require.Zero(t, calls, "fn ran before the sequence was consumed")
	require.Equal(t, []int{2, 4, 6}, slices.Collect(double))
	require.Equal(t, 3, calls)
}

func TestMapStopsEarly(t *testing.T) {
	t.Parallel()

	var pulled, mapped int

	var src iter.Seq[int] = func(yield func(int) bool) {
		for v := range 10 {
			pulled++
			if !yield(v) {
				return
			}
		}
	}

	var got []int
	for v := range seq.Map(src, func(v int) int { mapped++; return v * 2 }) {
		got = append(got, v)
		if len(got) == 3 {
			break
		}
	}

	require.Equal(t, []int{0, 2, 4}, got)
	require.Equal(t, 3, pulled, "source kept producing after the consumer stopped")
	require.Equal(t, 3, mapped, "fn ran for values the consumer never saw")
}

func TestMapErr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []int
		want []mapped
	}{
		{name: "empty", in: nil, want: nil},
		{name: "all succeed", in: []int{2, 4}, want: []mapped{{val: "2"}, {val: "4"}}},
		{name: "error first", in: []int{1, 2}, want: []mapped{{err: errOdd}, {val: "2"}}},
		{name: "error in the middle", in: []int{2, 3, 4}, want: []mapped{{val: "2"}, {err: errOdd}, {val: "4"}}},
		{name: "error last", in: []int{2, 1}, want: []mapped{{val: "2"}, {err: errOdd}}},
		{name: "every value fails", in: []int{1, 3}, want: []mapped{{err: errOdd}, {err: errOdd}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := collectMapped(seq.MapErr(slices.Values(tt.in), evenToString))
			require.Equal(t, tt.want, got)
		})
	}
}

func TestMapErrContinuesPastError(t *testing.T) {
	t.Parallel()

	pulled := 0

	var src iter.Seq[int] = func(yield func(int) bool) {
		for _, v := range []int{1, 2, 3, 4} {
			pulled++
			if !yield(v) {
				return
			}
		}
	}

	var got []string
	failures := 0

	for v, err := range seq.MapErr(src, evenToString) {
		if err != nil {
			failures++
			continue
		}

		got = append(got, v)
	}

	require.Equal(t, []string{"2", "4"}, got)
	require.Equal(t, 2, failures)
	require.Equal(t, 4, pulled, "MapErr stopped on an error instead of letting the consumer decide")
}

func TestMapErrStopsEarly(t *testing.T) {
	t.Parallel()

	var pulled, calls int

	var src iter.Seq[int] = func(yield func(int) bool) {
		for _, v := range []int{2, 4, 5, 6, 8} {
			pulled++
			if !yield(v) {
				return
			}
		}
	}

	var got []string
	var failure error

	for v, err := range seq.MapErr(src, func(v int) (string, error) { calls++; return evenToString(v) }) {
		if err != nil {
			failure = err
			break
		}

		got = append(got, v)
	}

	require.Equal(t, []string{"2", "4"}, got)
	require.ErrorIs(t, failure, errOdd)
	require.Equal(t, 3, pulled, "source kept producing after the consumer stopped")
	require.Equal(t, 3, calls, "fn ran for values the consumer never saw")
}

func TestMapErrIsLazy(t *testing.T) {
	t.Parallel()

	calls := 0
	strs := seq.MapErr(slices.Values([]int{1, 2, 3}), func(v int) (string, error) {
		calls++
		return evenToString(v)
	})

	require.Zero(t, calls, "fn ran before the sequence was consumed")
	require.Len(t, collectMapped(strs), 3)
	require.Equal(t, 3, calls, "fn should run exactly once per source value")
}

func ExampleMap() {
	nums := slices.Values([]int{1, 2, 3})

	for s := range seq.Map(nums, strconv.Itoa) {
		fmt.Println(s)
	}
	// Output:
	// 1
	// 2
	// 3
}

func ExampleMapErr() {
	vals := slices.Values([]string{"1", "2", "nope", "4"})

	for v, err := range seq.MapErr(vals, strconv.Atoi) {
		if err != nil {
			fmt.Println("stopping:", err)
			break
		}

		fmt.Println(v)
	}
	// Output:
	// 1
	// 2
	// stopping: strconv.Atoi: parsing "nope": invalid syntax
}

func ExampleMapErr_continueOnError() {
	vals := slices.Values([]string{"1", "nope", "3"})

	sum := 0

	for v, err := range seq.MapErr(vals, strconv.Atoi) {
		if err != nil {
			continue
		}

		sum += v
	}

	fmt.Println("sum:", sum)
	// Output: sum: 4
}

// evenToString fails on odd values, so tests can assert error position without
// depending on the shape of a stdlib error.
func evenToString(v int) (string, error) {
	if v%2 != 0 {
		return "", errOdd
	}

	return strconv.Itoa(v), nil
}

func collectMapped(s iter.Seq2[string, error]) []mapped {
	var out []mapped
	for v, err := range s {
		out = append(out, mapped{val: v, err: err})
	}

	return out
}
