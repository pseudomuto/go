package seq_test

import (
	"fmt"
	"iter"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pseudomuto/go/seq"
)

// person holds a slice, so it is not comparable. Keeping it that way means this
// file stops compiling if UniqBy's T is ever narrowed back to comparable.
type person struct {
	Name string
	Tags []string
}

func TestUniq(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{name: "empty", in: nil, want: nil},
		{name: "all distinct", in: []string{"a", "b", "c"}, want: []string{"a", "b", "c"}},
		{name: "adjacent duplicates", in: []string{"a", "a", "b"}, want: []string{"a", "b"}},
		{name: "non adjacent duplicates", in: []string{"a", "b", "a", "c", "b"}, want: []string{"a", "b", "c"}},
		{name: "every value the same", in: []string{"a", "a", "a"}, want: []string{"a"}},
		{name: "zero value is not treated as absent", in: []string{"", "", "a"}, want: []string{"", "a"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := slices.Collect(seq.Uniq(slices.Values(tt.in)))
			require.Equal(t, tt.want, got)
		})
	}
}

func TestUniqIsRepeatable(t *testing.T) {
	t.Parallel()

	uniq := seq.Uniq(slices.Values([]int{1, 1, 2, 3, 2}))

	require.Equal(t, []int{1, 2, 3}, slices.Collect(uniq))
	require.Equal(t, []int{1, 2, 3}, slices.Collect(uniq), "second pass reused the seen set from the first")
}

func TestUniqStopsEarly(t *testing.T) {
	t.Parallel()

	pulled := 0

	var src iter.Seq[int] = func(yield func(int) bool) {
		for _, v := range []int{1, 1, 2, 3, 4} {
			pulled++
			if !yield(v) {
				return
			}
		}
	}

	var got []int
	for v := range seq.Uniq(src) {
		got = append(got, v)
		if len(got) == 2 {
			break
		}
	}

	require.Equal(t, []int{1, 2}, got)
	require.Equal(t, 3, pulled, "source kept producing after the consumer stopped")
}

func TestUniqBy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []person
		want []person
	}{
		{name: "empty", in: nil, want: nil},
		{
			name: "all keys distinct",
			in:   []person{{Name: "ada"}, {Name: "grace"}},
			want: []person{{Name: "ada"}, {Name: "grace"}},
		},
		{
			name: "same key keeps the first value",
			in:   []person{{Name: "ada", Tags: []string{"first"}}, {Name: "ada", Tags: []string{"second"}}},
			want: []person{{Name: "ada", Tags: []string{"first"}}},
		},
		{
			name: "non adjacent duplicate keys",
			in:   []person{{Name: "ada"}, {Name: "grace"}, {Name: "ada"}, {Name: "alan"}},
			want: []person{{Name: "ada"}, {Name: "grace"}, {Name: "alan"}},
		},
		{
			name: "every key the same",
			in:   []person{{Name: "ada", Tags: []string{"a"}}, {Name: "ada"}, {Name: "ada"}},
			want: []person{{Name: "ada", Tags: []string{"a"}}},
		},
		{
			name: "zero key is not treated as absent",
			in:   []person{{Name: ""}, {Name: ""}, {Name: "ada"}},
			want: []person{{Name: ""}, {Name: "ada"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := slices.Collect(seq.UniqBy(slices.Values(tt.in), personName))
			require.Equal(t, tt.want, got)
		})
	}
}

func TestUniqByCallsKeyOncePerValue(t *testing.T) {
	t.Parallel()

	calls := 0
	uniq := seq.UniqBy(
		slices.Values([]person{{Name: "ada"}, {Name: "ada"}, {Name: "grace"}}),
		func(p person) string {
			calls++
			return p.Name
		},
	)

	require.Zero(t, calls, "key was called before the sequence was consumed")
	require.Len(t, slices.Collect(uniq), 2)
	require.Equal(t, 3, calls, "key should run once per source value, duplicates included")
}

func TestUniqByIsRepeatable(t *testing.T) {
	t.Parallel()

	uniq := seq.UniqBy(
		slices.Values([]person{{Name: "ada"}, {Name: "ada"}, {Name: "grace"}}),
		personName,
	)

	want := []person{{Name: "ada"}, {Name: "grace"}}

	require.Equal(t, want, slices.Collect(uniq))
	require.Equal(t, want, slices.Collect(uniq), "second pass reused the key set from the first")
}

func TestUniqByStopsEarly(t *testing.T) {
	t.Parallel()

	pulled := 0

	var src iter.Seq[person] = func(yield func(person) bool) {
		for _, p := range []person{{Name: "ada"}, {Name: "ada"}, {Name: "grace"}, {Name: "alan"}} {
			pulled++
			if !yield(p) {
				return
			}
		}
	}

	var got []person
	for p := range seq.UniqBy(src, personName) {
		got = append(got, p)
		if len(got) == 2 {
			break
		}
	}

	require.Equal(t, []person{{Name: "ada"}, {Name: "grace"}}, got)
	require.Equal(t, 3, pulled, "source kept producing after the consumer stopped")
}

func ExampleUniq() {
	names := slices.Values([]string{"ada", "grace", "ada", "alan", "grace"})

	fmt.Println(slices.Collect(seq.Uniq(names)))
	// Output: [ada grace alan]
}

func ExampleUniqBy() {
	type account struct {
		Name  string
		Email string
	}

	accounts := slices.Values([]account{
		{Name: "ada", Email: "ada@example.com"},
		{Name: "grace", Email: "grace@example.com"},
		{Name: "ada", Email: "ada@work.example.com"},
	})

	for a := range seq.UniqBy(accounts, func(a account) string { return a.Name }) {
		fmt.Println(a.Name, a.Email)
	}
	// Output:
	// ada ada@example.com
	// grace grace@example.com
}

func personName(p person) string { return p.Name }
