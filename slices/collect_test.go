package slices_test

import (
	"errors"
	"fmt"
	"iter"
	stdslices "slices"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pseudomuto/go/seq"
	"github.com/pseudomuto/go/slices"
)

var errBoom = errors.New("boom")

type pair struct {
	val string
	err error
}

func TestCollectErr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      []pair
		want    []string
		wantErr error
	}{
		{name: "empty", in: nil, want: nil, wantErr: nil},
		{
			name: "every pair succeeds",
			in:   []pair{{val: "a"}, {val: "b"}},
			want: []string{"a", "b"},
		},
		{
			name:    "error on the first pair",
			in:      []pair{{err: errBoom}, {val: "b"}},
			want:    nil,
			wantErr: errBoom,
		},
		{
			name:    "error in the middle",
			in:      []pair{{val: "a"}, {err: errBoom}, {val: "c"}},
			want:    nil,
			wantErr: errBoom,
		},
		{
			name:    "error on the last pair discards earlier values",
			in:      []pair{{val: "a"}, {val: "b"}, {err: errBoom}},
			want:    nil,
			wantErr: errBoom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := slices.CollectErr(pairsOf(tt.in...))

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func TestCollectErrStopsAtFirstError(t *testing.T) {
	t.Parallel()

	produced := 0

	var src iter.Seq2[string, error] = func(yield func(string, error) bool) {
		for _, p := range []pair{{val: "a"}, {val: "b"}, {err: errBoom}, {val: "d"}} {
			produced++
			if !yield(p.val, p.err) {
				return
			}
		}
	}

	got, err := slices.CollectErr(src)

	require.ErrorIs(t, err, errBoom)
	require.Nil(t, got)
	require.Equal(t, 3, produced, "CollectErr kept consuming the source after the first error")
}

func TestCollectErrWithMapErr(t *testing.T) {
	t.Parallel()

	t.Run("every value parses", func(t *testing.T) {
		t.Parallel()

		got, err := slices.CollectErr(seq.MapErr(stdslices.Values([]string{"1", "2", "3"}), strconv.Atoi))

		require.NoError(t, err)
		require.Equal(t, []int{1, 2, 3}, got)
	})

	t.Run("the error chain survives the round trip", func(t *testing.T) {
		t.Parallel()

		got, err := slices.CollectErr(seq.MapErr(stdslices.Values([]string{"1", "nope"}), strconv.Atoi))

		require.Nil(t, got)
		require.ErrorIs(t, err, strconv.ErrSyntax)

		var numErr *strconv.NumError
		require.ErrorAs(t, err, &numErr)
		require.Equal(t, "nope", numErr.Num, "the error should still name the value that failed")
	})
}

func ExampleCollectErr() {
	nums, err := slices.CollectErr(seq.MapErr(stdslices.Values([]string{"1", "2", "3"}), strconv.Atoi))
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(nums)
	// Output: [1 2 3]
}

func ExampleCollectErr_error() {
	// A failure anywhere discards everything, so nums is nil even though "1"
	// parsed fine.
	nums, err := slices.CollectErr(seq.MapErr(stdslices.Values([]string{"1", "nope", "3"}), strconv.Atoi))

	fmt.Println(nums)
	fmt.Println(err)
	// Output:
	// []
	// strconv.Atoi: parsing "nope": invalid syntax
}

// pairsOf builds a Seq2 directly, so these tests do not depend on whatever
// produced the pairs.
func pairsOf(ps ...pair) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		for _, p := range ps {
			if !yield(p.val, p.err) {
				return
			}
		}
	}
}
