package chain_test

import (
	"cmp"
	"fmt"
	stdslices "slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pseudomuto/go/chain"
	"github.com/pseudomuto/go/maps"
)

func evenValue(e maps.Entry[string, int]) bool { return e.Value%2 == 0 }

func TestMap_Filter(t *testing.T) {
	t.Parallel()

	got := chain.OfMap(map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}).Filter(evenValue)

	require.Equal(t, map[string]int{"b": 2, "d": 4}, map[string]int(got))
}

func TestMap_Reject(t *testing.T) {
	t.Parallel()

	got := chain.OfMap(map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}).Reject(evenValue)

	require.Equal(t, map[string]int{"a": 1, "c": 3}, map[string]int(got))
}

func TestMap_Map(t *testing.T) {
	t.Parallel()

	// Yields in map order, so compare sorted.
	got := chain.OfMap(map[string]int{"a": 1, "bb": 2}).
		Map(func(e maps.Entry[string, int]) string { return e.Key + strconv.Itoa(e.Value) }).
		Collect().
		SortFunc(cmp.Compare[string])

	require.Equal(t, []string{"a1", "bb2"}, []string(got))
}

func TestMap_MapErr(t *testing.T) {
	t.Parallel()

	fails := func(e maps.Entry[string, int]) (int, error) {
		if e.Value%2 != 0 {
			return 0, strconv.ErrRange
		}

		return e.Value, nil
	}

	ok, failed := 0, 0

	for _, err := range chain.OfMap(map[string]int{"a": 1, "b": 2, "c": 4}).MapErr(fails) {
		if err != nil {
			failed++
			continue
		}

		ok++
	}

	require.Equal(t, 2, ok)
	require.Equal(t, 1, failed)
}

func TestMap_MapKeys(t *testing.T) {
	t.Parallel()

	t.Run("same type", func(t *testing.T) {
		t.Parallel()

		got := stdslices.Sorted(chain.OfMap(map[string]int{"A": 1, "B": 2}).MapKeys(strings.ToLower).Unwrap())

		require.Equal(t, []string{"a", "b"}, got)
	})

	t.Run("changes the key type", func(t *testing.T) {
		t.Parallel()

		lengths := chain.OfMap(map[string]int{"a": 1, "bbb": 2}).MapKeys(func(k string) int { return len(k) })

		require.Equal(t, []int{1, 3}, stdslices.Sorted(lengths.Unwrap()))
	})
}

func TestMap_MapValues(t *testing.T) {
	t.Parallel()

	// Now keeps chaining as a Seq, because Seq no longer needs a comparable type.
	got := chain.OfMap(map[string]int{"a": 1, "b": 2}).
		MapValues(strconv.Itoa).
		Collect().
		SortFunc(cmp.Compare[string])

	require.Equal(t, []string{"1", "2"}, []string(got))
}

func TestMap_UniqKeysBy(t *testing.T) {
	t.Parallel()

	// "Ada" and "ada" collapse; ascending key order decides which survives.
	got := chain.OfMap(map[string]int{"Ada": 1, "ada": 2, "Grace": 3}).
		UniqKeysBy(strings.ToLower).
		Collect()

	require.Equal(t, []string{"Ada", "Grace"}, []string(got))
}

func TestMap_UniqValuesBy(t *testing.T) {
	t.Parallel()

	got := chain.OfMap(map[string]string{"k1": "Ada", "k2": "ada", "k3": "Grace"}).
		UniqValuesBy(strings.ToLower).
		Collect()

	require.Equal(t, []string{"Ada", "Grace"}, []string(got))
}

func TestMap_SortedKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   map[string]int
		want []string
	}{
		{name: "nil map", in: nil, want: nil},
		{name: "empty map", in: map[string]int{}, want: nil},
		{name: "ascending order", in: map[string]int{"c": 3, "a": 1, "b": 2}, want: []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, []string(chain.OfMap(tt.in).SortedKeys()))
		})
	}
}

func TestMap_SortedKeysFunc(t *testing.T) {
	t.Parallel()

	descending := func(a, b string) int { return cmp.Compare(b, a) }

	got := chain.OfMap(map[string]int{"a": 1, "b": 2, "c": 3}).SortedKeysFunc(descending)

	require.Equal(t, []string{"c", "b", "a"}, []string(got))
}

func TestMap_SortedValuesFunc(t *testing.T) {
	t.Parallel()

	// Now keeps chaining as a Slice, because Slice no longer needs a comparable type.
	got := chain.OfMap(map[string]int{"a": 3, "b": 1, "c": 2}).
		SortedValuesFunc(cmp.Compare[int]).
		Reject(isEven)

	require.Equal(t, []int{1, 3}, []int(got))
}

func TestMap_Entries(t *testing.T) {
	t.Parallel()

	// Also chains now: a Slice of Entry no longer needs V to be comparable.
	got := chain.OfMap(map[string]int{"b": 2, "a": 1}).
		Entries().
		Map(func(e maps.Entry[string, int]) string { return e.Key })

	require.Equal(t, []string{"a", "b"}, []string(got))
}

func TestMap_Reduce(t *testing.T) {
	t.Parallel()

	t.Run("accumulator matches the value type", func(t *testing.T) {
		t.Parallel()

		total := chain.OfMap(map[string]int{"a": 1, "b": 2, "c": 3}).
			Reduce(0, func(acc int, e maps.Entry[string, int]) int { return acc + e.Value })

		require.Equal(t, 6, total)
	})

	t.Run("accumulator differs from the value type", func(t *testing.T) {
		t.Parallel()

		// Folds in ascending key order, so this is stable.
		joined := chain.OfMap(map[string]int{"c": 3, "a": 1, "b": 2}).
			Reduce("", func(acc string, e maps.Entry[string, int]) string {
				return acc + e.Key + strconv.Itoa(e.Value)
			})

		require.Equal(t, "a1b2c3", joined)
	})
}

func TestMap_ReduceKeys(t *testing.T) {
	t.Parallel()

	joined := chain.OfMap(map[string]int{"c": 3, "a": 1, "b": 2}).
		ReduceKeys("", func(acc, k string) string { return acc + k })

	require.Equal(t, "abc", joined)
}

func TestMap_ReduceValues(t *testing.T) {
	t.Parallel()

	// Accumulator differs from the value type.
	joined := chain.OfMap(map[string]int{"c": 3, "a": 1, "b": 2}).
		ReduceValues("", func(acc string, v int) string { return acc + strconv.Itoa(v) })

	require.Equal(t, "123", joined)
}

func TestMap_IsAMap(t *testing.T) {
	t.Parallel()

	// The point of a defined type over map[K]V: no Unwrap or Len needed.
	got := chain.OfMap(map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}).Filter(evenValue)

	t.Run("len works", func(t *testing.T) {
		t.Parallel()

		require.Len(t, got, 2)
	})

	t.Run("indexing works", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, 2, got["b"])

		_, ok := got["a"]
		require.False(t, ok, "a should have been filtered out")
	})

	t.Run("ranging works", func(t *testing.T) {
		t.Parallel()

		total := 0
		for _, v := range got {
			total += v
		}

		require.Equal(t, 6, total)
	})

	t.Run("assignable to a plain map with no conversion", func(t *testing.T) {
		t.Parallel()

		var plain map[string]int = got

		require.Equal(t, map[string]int{"b": 2, "d": 4}, plain)
	})

	t.Run("composite literal works", func(t *testing.T) {
		t.Parallel()

		m := chain.Map[string, int]{"a": 1, "b": 2}

		require.Equal(t, []string{"a", "b"}, []string(m.SortedKeys()))
	})
}

func TestMap_LeavesInputAlone(t *testing.T) {
	t.Parallel()

	in := map[string]int{"a": 1, "b": 2}

	chain.OfMap(in).Filter(evenValue)["c"] = 3

	require.Equal(t, map[string]int{"a": 1, "b": 2}, in, "the chain modified the caller's map")
}

func TestMap_WrapsNonComparableValues(t *testing.T) {
	t.Parallel()

	// V is unconstrained, so this shape works.
	headers := map[string][]string{
		"accept": {"application/json"},
		"host":   {"example.com"},
	}

	got := chain.OfMap(headers).Map(func(e maps.Entry[string, []string]) int { return len(e.Value) })

	require.Equal(t, []int{1, 1}, stdslices.Sorted(got.Unwrap()))
}

func TestMap_CrossesContainers(t *testing.T) {
	t.Parallel()

	// map -> slice -> new element type -> dedupe on a derived key -> one value.
	report := chain.OfMap(map[string]int{"a": 1, "bb": 2, "ccc": 3, "dddd": 4}).
		Filter(evenValue).
		SortedKeys().
		Map(func(k string) user { return user{Name: k} }).
		UniqBy(func(u user) int { return len(u.Name) }).
		Reduce(0, func(acc int, u user) int { return acc + len(u.Name) })

	require.Equal(t, 6, report)
}

func ExampleOfMap() {
	m := map[string]int{"ada": 1, "grace": 2, "alan": 3}

	// Nested version needs two statements and a free-function call in the middle.
	joined := chain.OfMap(m).
		Filter(func(e maps.Entry[string, int]) bool { return e.Value%2 != 0 }).
		SortedKeys().
		Map(strings.ToUpper).
		Reduce("", func(acc, k string) string { return acc + k })

	fmt.Println(joined)
	// Output: ADAALAN
}

func ExampleMap_MapKeys() {
	m := map[string]int{"Accept": 1, "Host": 2}

	// MapKeys yields a sequence of transformed keys, not a re-keyed map, and it
	// arrives in map order, so sort before printing.
	keys := chain.OfMap(m).MapKeys(strings.ToLower).Collect().SortFunc(cmp.Compare[string])

	fmt.Println(keys)
	// Output: [accept host]
}
