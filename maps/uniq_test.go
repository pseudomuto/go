package maps_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pseudomuto/go/maps"
)

func TestUniqKeysBy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   map[string]int
		want []string
	}{
		{name: "nil map", in: nil, want: nil},
		{name: "empty map", in: map[string]int{}, want: nil},
		{
			name: "nothing collapses, every key survives in sorted order",
			in:   map[string]int{"b": 2, "a": 1, "c": 3},
			want: []string{"a", "b", "c"},
		},
		{
			name: "case folding keeps the first key in sorted order",
			in:   map[string]int{"Ada": 1, "ada": 2, "Grace": 3},
			want: []string{"Ada", "Grace"},
		},
		{
			name: "every key collapses to one",
			in:   map[string]int{"a": 1, "A": 2},
			want: []string{"A"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, slices.Collect(maps.UniqKeysBy(tt.in, strings.ToLower)))
		})
	}
}

func TestUniqKeysByIsDeterministic(t *testing.T) {
	t.Parallel()

	// "Ada" and "ada" collapse to the same key, so something has to break the
	// tie. Map iteration order is randomized, so the tie has to be broken by
	// sorted key order or the answer changes between identical calls.
	m := map[string]int{"Ada": 1, "ada": 2, "Grace": 3}

	first := slices.Collect(maps.UniqKeysBy(m, strings.ToLower))

	for range 500 {
		require.Equal(t, first, slices.Collect(maps.UniqKeysBy(m, strings.ToLower)),
			"repeated calls on the same map disagree")
	}
}

func TestUniqKeysByIsRepeatable(t *testing.T) {
	t.Parallel()

	uniq := maps.UniqKeysBy(map[string]int{"Ada": 1, "ada": 2, "Grace": 3}, strings.ToLower)

	want := []string{"Ada", "Grace"}

	require.Equal(t, want, slices.Collect(uniq))
	require.Equal(t, want, slices.Collect(uniq), "second pass disagreed with the first")
}

func TestUniqKeysBySeesLaterMapChanges(t *testing.T) {
	t.Parallel()

	m := map[string]int{"a": 1}
	uniq := maps.UniqKeysBy(m, strings.ToLower)

	require.Equal(t, []string{"a"}, slices.Collect(uniq))

	m["b"] = 2

	require.Equal(t, []string{"a", "b"}, slices.Collect(uniq),
		"the sequence served keys from a stale call-time snapshot")
}

func TestUniqKeysByStopsEarly(t *testing.T) {
	t.Parallel()

	m := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}

	var got []string
	for k := range maps.UniqKeysBy(m, strings.ToLower) {
		got = append(got, k)
		if len(got) == 2 {
			break
		}
	}

	require.Equal(t, []string{"a", "b"}, got)
}

func TestUniqValuesBy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   map[string]string
		want []string
	}{
		{name: "nil map", in: nil, want: nil},
		{name: "empty map", in: map[string]string{}, want: nil},
		{
			name: "distinct values all survive in key order",
			in:   map[string]string{"k2": "b", "k1": "a", "k3": "c"},
			want: []string{"a", "b", "c"},
		},
		{
			name: "repeated values collapse to the first by key order",
			in:   map[string]string{"k1": "a", "k2": "a", "k3": "b"},
			want: []string{"a", "b"},
		},
		{
			name: "case folding keeps the first value by key order",
			in:   map[string]string{"k1": "Ada", "k2": "ada", "k3": "Grace"},
			want: []string{"Ada", "Grace"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, slices.Collect(maps.UniqValuesBy(tt.in, strings.ToLower)))
		})
	}
}

func TestUniqValuesByIsDeterministic(t *testing.T) {
	t.Parallel()

	// Two entries hold values that collapse under the key func, so something
	// has to break the tie.
	m := map[string]string{"k1": "Ada", "k2": "ada", "k3": "Grace"}

	first := slices.Collect(maps.UniqValuesBy(m, strings.ToLower))

	for range 500 {
		require.Equal(t, first, slices.Collect(maps.UniqValuesBy(m, strings.ToLower)),
			"repeated calls on the same map disagree")
	}
}

func TestUniqValuesByIsRepeatable(t *testing.T) {
	t.Parallel()

	uniq := maps.UniqValuesBy(map[string]string{"k1": "Ada", "k2": "ada", "k3": "Grace"}, strings.ToLower)

	want := []string{"Ada", "Grace"}

	require.Equal(t, want, slices.Collect(uniq))
	require.Equal(t, want, slices.Collect(uniq), "second pass disagreed with the first")
}

func TestUniqValuesBySeesLaterMapChanges(t *testing.T) {
	t.Parallel()

	m := map[string]string{"k1": "a"}
	uniq := maps.UniqValuesBy(m, strings.ToLower)

	require.Equal(t, []string{"a"}, slices.Collect(uniq))

	m["k2"] = "b"

	require.Equal(t, []string{"a", "b"}, slices.Collect(uniq),
		"the sequence served values from a stale call-time snapshot")
}

func TestUniqValuesByStopsEarly(t *testing.T) {
	t.Parallel()

	m := map[string]string{"k1": "a", "k2": "b", "k3": "c", "k4": "d"}

	var got []string
	for v := range maps.UniqValuesBy(m, strings.ToLower) {
		got = append(got, v)
		if len(got) == 2 {
			break
		}
	}

	require.Equal(t, []string{"a", "b"}, got)
}

func ExampleUniqKeysBy() {
	// "Ada" and "ada" collapse under ToLower. UniqKeysBy walks the map in
	// ascending key order, so "Ada" wins on every run.
	m := map[string]int{"Ada": 1, "ada": 2, "Grace": 3}

	for k := range maps.UniqKeysBy(m, strings.ToLower) {
		fmt.Println(k)
	}
	// Output:
	// Ada
	// Grace
}

func ExampleUniqValuesBy() {
	// "Ada" and "ada" collapse under ToLower. UniqValuesBy walks the map in
	// ascending key order, so k1's value wins on every run.
	m := map[string]string{"k1": "Ada", "k2": "ada", "k3": "Grace"}

	for v := range maps.UniqValuesBy(m, strings.ToLower) {
		fmt.Println(v)
	}
	// Output:
	// Ada
	// Grace
}
