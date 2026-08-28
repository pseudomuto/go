package maps_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pseudomuto/go/maps"
)

func TestReduce(t *testing.T) {
	t.Parallel()

	// Key and value are the same type on purpose: this is exactly the case where
	// a positional (acc, k, v) callback would let a swap through silently.
	kv := func(acc string, e maps.Entry[string, string]) string {
		return acc + e.Key + "=" + e.Value + ";"
	}

	tests := []struct {
		name string
		in   map[string]string
		want string
	}{
		{name: "nil map returns init untouched", in: nil, want: "seed:"},
		{name: "empty map returns init untouched", in: map[string]string{}, want: "seed:"},
		{name: "single entry", in: map[string]string{"a": "1"}, want: "seed:a=1;"},
		{
			name: "entries fold in ascending key order",
			in:   map[string]string{"c": "3", "a": "1", "b": "2"},
			want: "seed:a=1;b=2;c=3;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, maps.Reduce(tt.in, "seed:", kv))
		})
	}
}

func TestReduceIsDeterministic(t *testing.T) {
	t.Parallel()

	m := map[string]string{"e": "5", "a": "1", "d": "4", "b": "2", "c": "3"}

	kv := func(acc string, e maps.Entry[string, string]) string { return acc + e.Key + e.Value }

	first := maps.Reduce(m, "", kv)

	for range 500 {
		require.Equal(t, first, maps.Reduce(m, "", kv), "repeated calls on the same map disagree")
	}
}

func TestReduceChangesAccumulatorType(t *testing.T) {
	t.Parallel()

	m := map[string]int{"a": 1, "bb": 2, "ccc": 3}

	// String keys and int values folded into a single int.
	total := maps.Reduce(m, 0, func(acc int, e maps.Entry[string, int]) int {
		return acc + len(e.Key)*e.Value
	})

	require.Equal(t, 1*1+2*2+3*3, total)
}

func TestReduceKeepsKeyAndValueDistinct(t *testing.T) {
	t.Parallel()

	// Guards the whole reason Entry exists: a positional callback on a
	// map[string]string compiles with key and value swapped.
	m := map[string]string{"k": "v"}

	got := maps.Reduce(m, "", func(acc string, e maps.Entry[string, string]) string {
		return acc + "key=" + e.Key + ",value=" + e.Value
	})

	require.Equal(t, "key=k,value=v", got)
}

func TestReduceKeys(t *testing.T) {
	t.Parallel()

	concat := func(acc, k string) string { return acc + k }

	tests := []struct {
		name string
		in   map[string]int
		want string
	}{
		{name: "nil map returns init untouched", in: nil, want: "seed"},
		{name: "empty map returns init untouched", in: map[string]int{}, want: "seed"},
		{name: "single key", in: map[string]int{"a": 1}, want: "seeda"},
		{
			name: "keys fold in ascending order",
			in:   map[string]int{"c": 3, "a": 1, "b": 2},
			want: "seedabc",
		},
		{
			name: "uppercase folds before lowercase",
			in:   map[string]int{"a": 1, "A": 2},
			want: "seedAa",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, maps.ReduceKeys(tt.in, "seed", concat))
		})
	}
}

func TestReduceKeysIsDeterministic(t *testing.T) {
	t.Parallel()

	// Concatenation is order sensitive, so raw map order would show up here.
	m := map[string]int{"c": 3, "a": 1, "b": 2, "d": 4, "e": 5}

	concat := func(acc, k string) string { return acc + k }

	first := maps.ReduceKeys(m, "", concat)

	for range 500 {
		require.Equal(t, first, maps.ReduceKeys(m, "", concat), "repeated calls on the same map disagree")
	}
}

func TestReduceValues(t *testing.T) {
	t.Parallel()

	concat := func(acc, v string) string { return acc + v }

	tests := []struct {
		name string
		in   map[string]string
		want string
	}{
		{name: "nil map returns init untouched", in: nil, want: "seed"},
		{name: "empty map returns init untouched", in: map[string]string{}, want: "seed"},
		{name: "single value", in: map[string]string{"k1": "a"}, want: "seeda"},
		{
			name: "values fold in ascending key order, not value order",
			in:   map[string]string{"k1": "z", "k2": "y", "k3": "x"},
			want: "seedzyx",
		},
		{
			name: "repeated values are all folded in",
			in:   map[string]string{"k1": "a", "k2": "a"},
			want: "seedaa",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, maps.ReduceValues(tt.in, "seed", concat))
		})
	}
}

func TestReduceValuesIsDeterministic(t *testing.T) {
	t.Parallel()

	m := map[string]string{"k1": "a", "k2": "b", "k3": "c", "k4": "d", "k5": "e"}

	concat := func(acc, v string) string { return acc + v }

	first := maps.ReduceValues(m, "", concat)

	for range 500 {
		require.Equal(t, first, maps.ReduceValues(m, "", concat), "repeated calls on the same map disagree")
	}
}

func TestReduceValuesChangesAccumulatorType(t *testing.T) {
	t.Parallel()

	m := map[string]string{"k1": "go", "k2": "is", "k3": "fun"}

	total := maps.ReduceValues(m, 0, func(acc int, v string) int { return acc + len(v) })

	require.Equal(t, 7, total)
}

func ExampleReduce() {
	m := map[string]int{"grace": 2, "ada": 1, "alan": 3}

	// Entries fold in ascending key order, so this is stable across runs.
	joined := maps.Reduce(m, "", func(acc string, e maps.Entry[string, int]) string {
		if acc != "" {
			acc += " "
		}

		return acc + fmt.Sprintf("%s=%d", e.Key, e.Value)
	})

	fmt.Println(joined)
	// Output: ada=1 alan=3 grace=2
}

func ExampleReduceKeys() {
	m := map[string]int{"grace": 2, "ada": 1, "alan": 3}

	// Keys fold in ascending order, so this is stable across runs.
	joined := maps.ReduceKeys(m, "", func(acc, k string) string {
		if acc == "" {
			return k
		}
		return acc + "," + k
	})

	fmt.Println(joined)
	// Output: ada,alan,grace
}

func ExampleReduceValues() {
	m := map[string]int{"a": 1, "b": 2, "c": 3}

	sum := maps.ReduceValues(m, 0, func(acc, v int) int { return acc + v })

	fmt.Println(sum)
	// Output: 6
}
