package maps_test

import (
	"cmp"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pseudomuto/go/maps"
)

func TestSortedKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   map[string]int
		want []string
	}{
		{name: "nil map", in: nil, want: nil},
		{name: "empty map", in: map[string]int{}, want: nil},
		{name: "single key", in: map[string]int{"a": 1}, want: []string{"a"}},
		{name: "unsorted keys", in: map[string]int{"c": 3, "a": 1, "b": 2}, want: []string{"a", "b", "c"}},
		{
			name: "uppercase sorts before lowercase",
			in:   map[string]int{"a": 1, "A": 2},
			want: []string{"A", "a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, maps.SortedKeys(tt.in))
		})
	}
}

func TestSortedKeysFunc(t *testing.T) {
	t.Parallel()

	m := map[string]int{"bb": 2, "a": 1, "cccc": 4}

	t.Run("by length", func(t *testing.T) {
		t.Parallel()

		byLength := func(a, b string) int { return cmp.Compare(len(a), len(b)) }

		require.Equal(t, []string{"a", "bb", "cccc"}, maps.SortedKeysFunc(m, byLength))
	})

	t.Run("descending", func(t *testing.T) {
		t.Parallel()

		descending := func(a, b string) int { return cmp.Compare(b, a) }

		require.Equal(t, []string{"cccc", "bb", "a"}, maps.SortedKeysFunc(m, descending))
	})

	t.Run("empty map", func(t *testing.T) {
		t.Parallel()

		require.Nil(t, maps.SortedKeysFunc(map[string]int{}, cmp.Compare[string]))
	})
}

func TestSortedKeysFuncWithStructKey(t *testing.T) {
	t.Parallel()

	// Sorting struct keys by a field is the case cmp.Ordered made impossible.
	m := map[pt]string{{X: 3}: "c", {X: 1}: "a", {X: 2}: "b"}

	byX := func(a, b pt) int { return cmp.Compare(a.X, b.X) }

	require.Equal(t, []pt{{X: 1}, {X: 2}, {X: 3}}, maps.SortedKeysFunc(m, byX))
}

func TestSortedValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   map[string]int
		want []int
	}{
		{name: "nil map", in: nil, want: nil},
		{name: "empty map", in: map[string]int{}, want: nil},
		{name: "single value", in: map[string]int{"a": 1}, want: []int{1}},
		{name: "unsorted values", in: map[string]int{"a": 3, "b": 1, "c": 2}, want: []int{1, 2, 3}},
		{
			name: "repeated values are all kept",
			in:   map[string]int{"a": 2, "b": 1, "c": 2},
			want: []int{1, 2, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, maps.SortedValues(tt.in))
		})
	}
}

func TestSortedValuesFunc(t *testing.T) {
	t.Parallel()

	m := map[string]string{"k1": "bb", "k2": "a", "k3": "cccc"}

	t.Run("by length", func(t *testing.T) {
		t.Parallel()

		byLength := func(a, b string) int { return cmp.Compare(len(a), len(b)) }

		require.Equal(t, []string{"a", "bb", "cccc"}, maps.SortedValuesFunc(m, byLength))
	})

	t.Run("descending", func(t *testing.T) {
		t.Parallel()

		descending := func(a, b string) int { return cmp.Compare(b, a) }

		require.Equal(t, []string{"cccc", "bb", "a"}, maps.SortedValuesFunc(m, descending))
	})

	t.Run("empty map", func(t *testing.T) {
		t.Parallel()

		require.Nil(t, maps.SortedValuesFunc(map[string]string{}, cmp.Compare[string]))
	})
}

func TestSortedValuesFuncWithStructValues(t *testing.T) {
	t.Parallel()

	// Struct values are not cmp.Ordered. Sorting them is what U any buys.
	type item struct {
		Name string
		Rank int
	}

	m := map[string]item{
		"c": {Name: "c", Rank: 3},
		"a": {Name: "a", Rank: 1},
		"b": {Name: "b", Rank: 2},
	}

	byRank := func(x, y item) int { return cmp.Compare(x.Rank, y.Rank) }

	require.Equal(t,
		[]item{{Name: "a", Rank: 1}, {Name: "b", Rank: 2}, {Name: "c", Rank: 3}},
		maps.SortedValuesFunc(m, byRank))
}

func ExampleSortedKeys() {
	m := map[string]int{"grace": 2, "ada": 1, "alan": 3}

	fmt.Println(maps.SortedKeys(m))
	// Output: [ada alan grace]
}

func ExampleSortedKeysFunc() {
	m := map[string]int{"bb": 2, "a": 1, "cccc": 4}

	byLength := func(a, b string) int { return cmp.Compare(len(a), len(b)) }

	fmt.Println(maps.SortedKeysFunc(m, byLength))
	// Output: [a bb cccc]
}

func ExampleSortedValues() {
	m := map[string]int{"a": 3, "b": 1, "c": 2}

	fmt.Println(maps.SortedValues(m))
	// Output: [1 2 3]
}

func ExampleSortedValuesFunc() {
	m := map[string]string{"k1": "bb", "k2": "a", "k3": "cccc"}

	byLength := func(a, b string) int { return cmp.Compare(len(a), len(b)) }

	fmt.Println(maps.SortedValuesFunc(m, byLength))
	// Output: [a bb cccc]
}
