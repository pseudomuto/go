package maps_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pseudomuto/go/maps"
)

func TestFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   map[string]int
		want map[string]int
	}{
		{name: "nil map yields an empty map, not nil", in: nil, want: map[string]int{}},
		{name: "empty map yields an empty map, not nil", in: map[string]int{}, want: map[string]int{}},
		{name: "nothing matches", in: map[string]int{"a": 1, "b": 3}, want: map[string]int{}},
		{
			name: "everything matches",
			in:   map[string]int{"a": 2, "b": 4},
			want: map[string]int{"a": 2, "b": 4},
		},
		{
			name: "some match",
			in:   map[string]int{"a": 1, "b": 2, "c": 3, "d": 4},
			want: map[string]int{"b": 2, "d": 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, maps.Filter(tt.in, evenValue))
		})
	}
}

func TestFilterKeepsKeyAndValueDistinct(t *testing.T) {
	t.Parallel()

	// Key and value share a type, so a transposed predicate would compile. If
	// pred saw them the wrong way round, e.Key would be "yes" or "no" and
	// nothing would match.
	m := map[string]string{"keep": "yes", "drop": "no"}

	got := maps.Filter(m, func(e maps.Entry[string, string]) bool { return e.Key == "keep" })

	require.Equal(t, map[string]string{"keep": "yes"}, got)
}

func TestFilterReturnsWritableMap(t *testing.T) {
	t.Parallel()

	got := maps.Filter(nil, func(maps.Entry[string, int]) bool { return true })
	require.NotNil(t, got)

	got["added"] = 1 // would panic on a nil map

	require.Equal(t, map[string]int{"added": 1}, got)
}

func TestFilterLeavesInputAlone(t *testing.T) {
	t.Parallel()

	in := map[string]int{"a": 1, "b": 2}

	got := maps.Filter(in, evenValue)
	got["c"] = 3

	require.Equal(t, map[string]int{"a": 1, "b": 2}, in, "Filter modified the map it was given")
	require.Len(t, got, 2)
}

func TestReject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   map[string]int
		want map[string]int
	}{
		{name: "nil map yields an empty map, not nil", in: nil, want: map[string]int{}},
		{name: "nothing rejected", in: map[string]int{"a": 1, "b": 3}, want: map[string]int{"a": 1, "b": 3}},
		{name: "everything rejected", in: map[string]int{"a": 2, "b": 4}, want: map[string]int{}},
		{
			name: "some rejected",
			in:   map[string]int{"a": 1, "b": 2, "c": 3, "d": 4},
			want: map[string]int{"a": 1, "c": 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, maps.Reject(tt.in, evenValue))
		})
	}
}

func TestRejectPartitionsWithFilter(t *testing.T) {
	t.Parallel()

	in := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}

	kept := maps.Filter(in, evenValue)
	dropped := maps.Reject(in, evenValue)

	require.Equal(t, map[string]int{"b": 2, "d": 4}, kept)
	require.Equal(t, map[string]int{"a": 1, "c": 3}, dropped)

	// Reassembling both halves must reproduce the input exactly: equal maps plus
	// matching total size means no entry was dropped or duplicated.
	union := maps.FromEntries(append(maps.Entries(kept), maps.Entries(dropped)...))

	require.Equal(t, in, union, "Filter and Reject do not cover the map exactly once")
	require.Equal(t, len(in), len(kept)+len(dropped))
}

func ExampleFilter() {
	m := map[string]int{"ada": 1, "grace": 2, "alan": 3}

	evens := maps.Filter(m, func(e maps.Entry[string, int]) bool { return e.Value%2 == 0 })

	fmt.Println(maps.Entries(evens))
	// Output: [{grace 2}]
}

func ExampleReject() {
	m := map[string]int{"ada": 1, "grace": 2, "alan": 3}

	odds := maps.Reject(m, func(e maps.Entry[string, int]) bool { return e.Value%2 == 0 })

	fmt.Println(maps.Entries(odds))
	// Output: [{ada 1} {alan 3}]
}

func evenValue(e maps.Entry[string, int]) bool { return e.Value%2 == 0 }
