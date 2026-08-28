package slices_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pseudomuto/go/slices"
)

func TestMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []int
		want []string
	}{
		{name: "nil", in: nil, want: nil},
		{name: "empty", in: []int{}, want: nil},
		{name: "single value", in: []int{1}, want: []string{"1"}},
		{name: "several values", in: []int{1, 2, 3}, want: []string{"1", "2", "3"}},
		{name: "duplicates are kept", in: []int{1, 1, 2}, want: []string{"1", "1", "2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, slices.Map(tt.in, strconv.Itoa))
		})
	}
}

func TestMapAllocatesExactly(t *testing.T) {
	t.Parallel()

	// Output length is known up front, so the result should be allocated once
	// with exactly enough room. slices.Collect would grow geometrically and
	// leave cap 4 here.
	in := []int{1, 2, 3}

	got := slices.Map(in, strconv.Itoa)

	require.Len(t, got, 3)
	require.Equal(t, 3, cap(got), "result was grown instead of preallocated")
}

func TestMapLeavesInputAlone(t *testing.T) {
	t.Parallel()

	in := []int{1, 2, 3}

	got := slices.Map(in, func(v int) int { return v * 10 })

	require.Equal(t, []int{10, 20, 30}, got)
	require.Equal(t, []int{1, 2, 3}, in, "Map modified the slice it was given")
}

func TestMapCallsFnOncePerElementInOrder(t *testing.T) {
	t.Parallel()

	var seen []int

	got := slices.Map([]int{3, 1, 2}, func(v int) int {
		seen = append(seen, v)
		return v
	})

	require.Equal(t, []int{3, 1, 2}, got)
	require.Equal(t, []int{3, 1, 2}, seen, "fn should run once per element, in slice order")
}

// evenToStr fails on odd values so the error position is easy to pin.
func evenToStr(v int) (string, error) {
	if v%2 != 0 {
		return "", errBoom
	}

	return strconv.Itoa(v), nil
}

func TestMapErr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      []int
		want    []string
		wantErr error
	}{
		{name: "nil", in: nil, want: nil},
		{name: "empty", in: []int{}, want: nil},
		{name: "every element succeeds", in: []int{2, 4}, want: []string{"2", "4"}},
		{name: "error on the first element", in: []int{1, 2}, want: nil, wantErr: errBoom},
		{name: "error in the middle", in: []int{2, 3, 4}, want: nil, wantErr: errBoom},
		{
			name:    "error on the last element discards earlier values",
			in:      []int{2, 4, 1},
			want:    nil,
			wantErr: errBoom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := slices.MapErr(tt.in, evenToStr)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func TestMapErrStopsAtFirstError(t *testing.T) {
	t.Parallel()

	calls := 0

	got, err := slices.MapErr([]int{2, 4, 1, 6, 8}, func(v int) (string, error) {
		calls++
		return evenToStr(v)
	})

	require.ErrorIs(t, err, errBoom)
	require.Nil(t, got)
	require.Equal(t, 3, calls, "fn kept running after the first error")
}

func TestMapErrPreservesTheErrorChain(t *testing.T) {
	t.Parallel()

	got, err := slices.MapErr([]string{"1", "nope"}, strconv.Atoi)

	require.Nil(t, got)
	require.ErrorIs(t, err, strconv.ErrSyntax)

	var numErr *strconv.NumError
	require.ErrorAs(t, err, &numErr)
	require.Equal(t, "nope", numErr.Num)
}

func TestMapErrLeavesInputAlone(t *testing.T) {
	t.Parallel()

	in := []int{2, 3, 4}

	_, err := slices.MapErr(in, evenToStr)

	require.ErrorIs(t, err, errBoom)
	require.Equal(t, []int{2, 3, 4}, in, "MapErr modified the slice it was given")
}

func ExampleMap() {
	fmt.Println(slices.Map([]int{1, 2, 3}, strconv.Itoa))
	// Output: [1 2 3]
}

func ExampleMapErr() {
	nums, err := slices.MapErr([]string{"1", "2", "3"}, strconv.Atoi)
	fmt.Println(nums, err)

	// One bad value discards the whole result.
	nums, err = slices.MapErr([]string{"1", "nope"}, strconv.Atoi)
	fmt.Println(nums, err)
	// Output:
	// [1 2 3] <nil>
	// [] strconv.Atoi: parsing "nope": invalid syntax
}
