package chain_test

import (
	"fmt"
	"slices"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pseudomuto/go/chain"
	pgoslices "github.com/pseudomuto/go/slices"
)

func isEven(v int) bool { return v%2 == 0 }

// user is not comparable, so it only wraps because Slice and Seq are unconstrained.
type user struct {
	Name string
	Tags []string
}

func TestSeq_Filter(t *testing.T) {
	t.Parallel()

	got := chain.OfSeq(slices.Values([]int{1, 2, 3, 4})).Filter(isEven).Collect()

	require.Equal(t, []int{2, 4}, []int(got))
}

func TestSeq_Reject(t *testing.T) {
	t.Parallel()

	got := chain.OfSeq(slices.Values([]int{1, 2, 3, 4})).Reject(isEven).Collect()

	require.Equal(t, []int{1, 3}, []int(got))
}

func TestSeq_UniqBy(t *testing.T) {
	t.Parallel()

	t.Run("with Identity, dedupes on the values", func(t *testing.T) {
		t.Parallel()

		got := chain.OfSeq(slices.Values([]int{1, 1, 2, 3, 2})).UniqBy(chain.Identity).Collect()

		require.Equal(t, []int{1, 2, 3}, []int(got))
	})

	t.Run("with a derived key, works on non-comparable values", func(t *testing.T) {
		t.Parallel()

		users := []user{{Name: "ada", Tags: []string{"x"}}, {Name: "ada"}, {Name: "bob"}}

		got := chain.OfSeq(slices.Values(users)).UniqBy(func(u user) string { return u.Name }).Collect()

		require.Len(t, got, 2)
		require.Equal(t, []string{"x"}, got[0].Tags, "the first occurrence should win")
	})
}

func TestSeq_Map(t *testing.T) {
	t.Parallel()

	t.Run("same type", func(t *testing.T) {
		t.Parallel()

		got := chain.OfSeq(slices.Values([]int{1, 2, 3})).Map(func(v int) int { return v * 10 }).Collect()

		require.Equal(t, []int{10, 20, 30}, []int(got))
	})

	t.Run("changes the element type", func(t *testing.T) {
		t.Parallel()

		got := chain.OfSeq(slices.Values([]int{1, 2, 3})).Map(strconv.Itoa).Collect()

		require.Equal(t, []string{"1", "2", "3"}, []string(got))
	})

	t.Run("changes to a non-comparable type", func(t *testing.T) {
		t.Parallel()

		got := chain.OfSeq(slices.Values([]string{"ada", "bob"})).
			Map(func(n string) user { return user{Name: n} }).
			Collect()

		require.Equal(t, []user{{Name: "ada"}, {Name: "bob"}}, []user(got))
	})
}

func TestSeq_MapErr(t *testing.T) {
	t.Parallel()

	t.Run("all succeed", func(t *testing.T) {
		t.Parallel()

		got, err := pgoslices.CollectErr(chain.OfSeq(slices.Values([]string{"1", "2"})).MapErr(strconv.Atoi))

		require.NoError(t, err)
		require.Equal(t, []int{1, 2}, got)
	})

	t.Run("first error wins and discards the rest", func(t *testing.T) {
		t.Parallel()

		got, err := pgoslices.CollectErr(chain.OfSeq(slices.Values([]string{"1", "nope"})).MapErr(strconv.Atoi))

		require.ErrorIs(t, err, strconv.ErrSyntax)
		require.Nil(t, got)
	})
}

func TestSeq_Reduce(t *testing.T) {
	t.Parallel()

	t.Run("accumulator matches the element type", func(t *testing.T) {
		t.Parallel()

		got := chain.OfSeq(slices.Values([]int{1, 2, 3})).Reduce(10, func(acc, v int) int { return acc + v })

		require.Equal(t, 16, got)
	})

	t.Run("accumulator differs from the element type", func(t *testing.T) {
		t.Parallel()

		got := chain.OfSeq(slices.Values([]int{1, 2, 3})).
			Reduce("", func(acc string, v int) string { return acc + strconv.Itoa(v) })

		require.Equal(t, "123", got)
	})

	t.Run("empty returns init", func(t *testing.T) {
		t.Parallel()

		got := chain.OfSeq(slices.Values([]int(nil))).Reduce(42, func(acc, v int) int { return acc + v })

		require.Equal(t, 42, got)
	})
}

func TestSeq_Collect(t *testing.T) {
	t.Parallel()

	t.Run("keeps chaining as a Slice", func(t *testing.T) {
		t.Parallel()

		got := chain.OfSeq(slices.Values([]int{3, 1, 3, 2})).UniqBy(chain.Identity).Collect().Reject(isEven)

		require.Equal(t, []int{3, 1}, []int(got))
	})

	t.Run("assignable to a plain slice with no conversion", func(t *testing.T) {
		t.Parallel()

		var plain []int = chain.OfSeq(slices.Values([]int{1, 2})).Collect()

		require.Equal(t, []int{1, 2}, plain)
	})

	t.Run("empty collects to nil", func(t *testing.T) {
		t.Parallel()

		require.Nil(t, []int(chain.OfSeq(slices.Values([]int(nil))).Collect()))
	})
}

func TestSeq_Unwrap(t *testing.T) {
	t.Parallel()

	// Needed because Seq and iter.Seq are both named types, so plain assignment
	// between them does not compile. Handing the result to slices.Collect, which
	// takes an iter.Seq, is the proof that it converts.
	unwrapped := chain.OfSeq(slices.Values([]int{1, 2, 3})).Filter(isEven).Unwrap()

	require.Equal(t, []int{2}, slices.Collect(unwrapped))
}

func TestSeq_RangesDirectly(t *testing.T) {
	t.Parallel()

	// The point of a defined type over iter.Seq: consume it without unwrapping.
	var got []int
	for v := range chain.OfSeq(slices.Values([]int{1, 2, 3, 4})).Filter(isEven) {
		got = append(got, v)
	}

	require.Equal(t, []int{2, 4}, got)
}

func TestSeq_StaysLazyUntilTerminal(t *testing.T) {
	t.Parallel()

	calls := 0

	pipeline := chain.OfSeq(slices.Values([]int{1, 2, 3})).
		Map(func(v int) int {
			calls++
			return v
		}).
		Filter(isEven)

	require.Zero(t, calls, "chaining should not consume the sequence")

	require.Equal(t, []int{2}, []int(pipeline.Collect()))
	require.Equal(t, 3, calls, "fn should run once per source value")
}

func TestIdentity(t *testing.T) {
	t.Parallel()

	require.Equal(t, 7, chain.Identity(7))
	require.Equal(t, "go", chain.Identity("go"))
}

func ExampleOfSeq() {
	// Reads left to right, and keeps going through a type change.
	joined := chain.OfSeq(slices.Values([]int{1, 2, 3, 4, 5, 6})).
		Reject(func(v int) bool { return v > 4 }).
		UniqBy(chain.Identity).
		Reduce("", func(acc string, v int) string { return acc + strconv.Itoa(v) })

	fmt.Println(joined)
	// Output: 1234
}

func ExampleSeq_Map() {
	// Map may change the element type, so the chain continues on the new one.
	got := chain.OfSeq(slices.Values([]int{1, 2, 3})).
		Map(strconv.Itoa).
		Map(func(s string) int { return len(s) }).
		Collect()

	fmt.Println(got)
	// Output: [1 1 1]
}
