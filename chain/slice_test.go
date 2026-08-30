package chain_test

import (
	"cmp"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pseudomuto/go/chain"
)

func TestSlice_Filter(t *testing.T) {
	t.Parallel()

	require.Equal(t, []int{2, 4}, []int(chain.OfSlice([]int{1, 2, 3, 4}).Filter(isEven)))
}

func TestSlice_Reject(t *testing.T) {
	t.Parallel()

	require.Equal(t, []int{1, 3}, []int(chain.OfSlice([]int{1, 2, 3, 4}).Reject(isEven)))
}

func TestSlice_UniqBy(t *testing.T) {
	t.Parallel()

	t.Run("with Identity, dedupes on the elements", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, []int{1, 2, 3}, []int(chain.OfSlice([]int{1, 1, 2, 3, 2}).UniqBy(chain.Identity)))
	})

	t.Run("with a derived key, works on non-comparable elements", func(t *testing.T) {
		t.Parallel()

		users := []user{{Name: "ada", Tags: []string{"x"}}, {Name: "ada"}, {Name: "bob"}}

		got := chain.OfSlice(users).UniqBy(func(u user) string { return u.Name })

		require.Len(t, got, 2)
		require.Equal(t, []string{"x"}, got[0].Tags, "the first occurrence should win")
	})
}

func TestSlice_Map(t *testing.T) {
	t.Parallel()

	t.Run("same type", func(t *testing.T) {
		t.Parallel()

		got := chain.OfSlice([]int{1, 2, 3}).Map(func(v int) int { return v * 10 })

		require.Equal(t, []int{10, 20, 30}, []int(got))
	})

	t.Run("changes the element type", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, []string{"1", "2", "3"}, []string(chain.OfSlice([]int{1, 2, 3}).Map(strconv.Itoa)))
	})

	t.Run("changes to a non-comparable type", func(t *testing.T) {
		t.Parallel()

		got := chain.OfSlice([]string{"ada", "bob"}).Map(func(n string) user { return user{Name: n} })

		require.Equal(t, []user{{Name: "ada"}, {Name: "bob"}}, []user(got))
	})
}

func TestSlice_MapErr(t *testing.T) {
	t.Parallel()

	t.Run("all succeed", func(t *testing.T) {
		t.Parallel()

		got, err := chain.OfSlice([]string{"1", "2"}).MapErr(strconv.Atoi)

		require.NoError(t, err)
		require.Equal(t, []int{1, 2}, []int(got))
	})

	t.Run("first error wins and discards the rest", func(t *testing.T) {
		t.Parallel()

		got, err := chain.OfSlice([]string{"1", "nope"}).MapErr(strconv.Atoi)

		require.ErrorIs(t, err, strconv.ErrSyntax)
		require.Nil(t, []int(got))
	})
}

func TestSlice_Reduce(t *testing.T) {
	t.Parallel()

	t.Run("accumulator matches the element type", func(t *testing.T) {
		t.Parallel()

		got := chain.OfSlice([]int{1, 2, 3}).Reduce(10, func(acc, v int) int { return acc + v })

		require.Equal(t, 16, got)
	})

	t.Run("accumulator differs from the element type", func(t *testing.T) {
		t.Parallel()

		got := chain.OfSlice([]int{1, 2, 3}).Reduce("", func(acc string, v int) string {
			return acc + strconv.Itoa(v)
		})

		require.Equal(t, "123", got)
	})

	t.Run("nil returns init", func(t *testing.T) {
		t.Parallel()

		got := chain.OfSlice([]int(nil)).Reduce(42, func(acc, v int) int { return acc + v })

		require.Equal(t, 42, got)
	})
}

func TestSlice_SortFunc(t *testing.T) {
	t.Parallel()

	t.Run("ascending", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, []int{1, 2, 3}, []int(chain.OfSlice([]int{3, 1, 2}).SortFunc(cmp.Compare[int])))
	})

	t.Run("descending", func(t *testing.T) {
		t.Parallel()

		descending := func(a, b int) int { return cmp.Compare(b, a) }

		require.Equal(t, []int{3, 2, 1}, []int(chain.OfSlice([]int{3, 1, 2}).SortFunc(descending)))
	})

	t.Run("leaves the receiver unsorted", func(t *testing.T) {
		t.Parallel()

		in := []int{3, 1, 2}

		require.Equal(t, []int{1, 2, 3}, []int(chain.OfSlice(in).SortFunc(cmp.Compare[int])))
		require.Equal(t, []int{3, 1, 2}, in, "SortFunc sorted the caller's slice in place")
	})
}

func TestSlice_Seq(t *testing.T) {
	t.Parallel()

	// Crosses into the lazy Seq wrapper, so the chain keeps going.
	got := chain.OfSlice([]int{1, 2, 3, 4}).Filter(isEven).Seq().Map(strconv.Itoa).Collect()

	require.Equal(t, []string{"2", "4"}, []string(got))
}

func TestSlice_IsASlice(t *testing.T) {
	t.Parallel()

	// The point of a defined type over []T: no Unwrap or Len needed.
	got := chain.OfSlice([]int{1, 2, 3, 4}).Filter(isEven)

	t.Run("len works", func(t *testing.T) {
		t.Parallel()

		require.Len(t, got, 2)
	})

	t.Run("indexing works", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, 2, got[0])
		require.Equal(t, 4, got[1])
	})

	t.Run("ranging works", func(t *testing.T) {
		t.Parallel()

		total := 0
		for _, v := range got {
			total += v
		}

		require.Equal(t, 6, total)
	})

	t.Run("assignable to a plain slice with no conversion", func(t *testing.T) {
		t.Parallel()

		var plain []int = got

		require.Equal(t, []int{2, 4}, plain)
	})

	t.Run("composite literal works", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, []int{1, 3}, []int(chain.Slice[int]{1, 2, 3}.Reject(isEven)))
	})
}

func TestSlice_LeavesInputAlone(t *testing.T) {
	t.Parallel()

	in := []int{1, 2, 3, 4}

	got := chain.OfSlice(in).Filter(isEven).Map(func(v int) int { return v * 10 })

	require.Equal(t, []int{20, 40}, []int(got))
	require.Equal(t, []int{1, 2, 3, 4}, in, "the chain modified the caller's slice")
}

func ExampleOfSlice() {
	names := []string{"ada", "grace", "ada", "alan", "grace"}

	// Dedupe, drop the long ones, then change type, all in one chain.
	lengths := chain.OfSlice(names).
		UniqBy(chain.Identity).
		Reject(func(n string) bool { return len(n) > 4 }).
		Map(func(n string) int { return len(n) })

	fmt.Println(lengths)
	// Output: [3 4]
}
