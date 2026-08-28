package maps_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pseudomuto/go/maps"
)

func TestEntries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   map[string]int
		want []maps.Entry[string, int]
	}{
		{name: "nil map", in: nil, want: nil},
		{name: "empty map", in: map[string]int{}, want: nil},
		{
			name: "single entry",
			in:   map[string]int{"a": 1},
			want: []maps.Entry[string, int]{{Key: "a", Value: 1}},
		},
		{
			name: "entries come back in ascending key order",
			in:   map[string]int{"c": 3, "a": 1, "b": 2},
			want: []maps.Entry[string, int]{{Key: "a", Value: 1}, {Key: "b", Value: 2}, {Key: "c", Value: 3}},
		},
		{
			// values deliberately run counter to key order, so a scrambled
			// pairing would show up here.
			name: "uppercase sorts before lowercase",
			in:   map[string]int{"a": 1, "A": 2},
			want: []maps.Entry[string, int]{{Key: "A", Value: 2}, {Key: "a", Value: 1}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, maps.Entries(tt.in))
		})
	}
}

func TestEntriesSortsNumericKeysNumerically(t *testing.T) {
	t.Parallel()

	// String sorting would give 1, 10, 2.
	m := map[int]string{10: "ten", 2: "two", 1: "one"}

	require.Equal(t, []maps.Entry[int, string]{
		{Key: 1, Value: "one"},
		{Key: 2, Value: "two"},
		{Key: 10, Value: "ten"},
	}, maps.Entries(m))
}

func TestEntriesIsDeterministic(t *testing.T) {
	t.Parallel()

	m := map[string]int{"e": 5, "a": 1, "d": 4, "b": 2, "c": 3}

	first := maps.Entries(m)

	for range 500 {
		require.Equal(t, first, maps.Entries(m), "repeated calls on the same map disagree")
	}
}

func TestFromEntries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []maps.Entry[string, int]
		want map[string]int
	}{
		{name: "nil slice yields an empty map, not nil", in: nil, want: map[string]int{}},
		{name: "empty slice yields an empty map, not nil", in: []maps.Entry[string, int]{}, want: map[string]int{}},
		{
			name: "single entry",
			in:   []maps.Entry[string, int]{{Key: "a", Value: 1}},
			want: map[string]int{"a": 1},
		},
		{
			name: "several entries",
			in:   []maps.Entry[string, int]{{Key: "a", Value: 1}, {Key: "b", Value: 2}},
			want: map[string]int{"a": 1, "b": 2},
		},
		{
			name: "duplicate keys keep the last value",
			in:   []maps.Entry[string, int]{{Key: "a", Value: 1}, {Key: "a", Value: 2}},
			want: map[string]int{"a": 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, maps.FromEntries(tt.in))
		})
	}
}

func TestFromEntriesReturnsWritableMap(t *testing.T) {
	t.Parallel()

	m := maps.FromEntries[string, int](nil)
	require.NotNil(t, m)

	m["added"] = 1 // would panic on a nil map

	require.Equal(t, map[string]int{"added": 1}, m)
}

func TestFromEntriesLeavesInputAlone(t *testing.T) {
	t.Parallel()

	in := []maps.Entry[string, int]{{Key: "a", Value: 1}, {Key: "b", Value: 2}}

	m := maps.FromEntries(in)
	m["c"] = 3

	require.Equal(t, []maps.Entry[string, int]{{Key: "a", Value: 1}, {Key: "b", Value: 2}}, in,
		"FromEntries modified the slice it was given")
	require.Len(t, m, 3)
}

func TestEntriesRoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("map to entries and back", func(t *testing.T) {
		t.Parallel()

		m := map[string]int{"c": 3, "a": 1, "b": 2}

		require.Equal(t, m, maps.FromEntries(maps.Entries(m)))
	})

	t.Run("sorted entries to map and back", func(t *testing.T) {
		t.Parallel()

		es := []maps.Entry[string, int]{{Key: "a", Value: 1}, {Key: "b", Value: 2}}

		require.Equal(t, es, maps.Entries(maps.FromEntries(es)))
	})
}

func ExampleEntries() {
	m := map[string]int{"grace": 2, "ada": 1, "alan": 3}

	// Ascending key order, so this listing is stable across runs.
	for _, e := range maps.Entries(m) {
		fmt.Printf("%s=%d\n", e.Key, e.Value)
	}
	// Output:
	// ada=1
	// alan=3
	// grace=2
}

func ExampleFromEntries() {
	es := []maps.Entry[string, int]{
		{Key: "ada", Value: 1},
		{Key: "grace", Value: 2},
	}

	m := maps.FromEntries(es)

	fmt.Println(len(m), m["ada"], m["grace"])
	// Output: 2 1 2
}
