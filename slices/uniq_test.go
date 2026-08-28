package slices_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pseudomuto/go/slices"
)

func TestUniq(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []int
		want []int
	}{
		{name: "nil", in: nil, want: nil},
		{name: "empty slice returns nil", in: []int{}, want: nil},
		{name: "single value", in: []int{1}, want: []int{1}},
		{name: "all distinct", in: []int{1, 2, 3}, want: []int{1, 2, 3}},
		{name: "adjacent duplicates", in: []int{1, 1, 2}, want: []int{1, 2}},
		{name: "non adjacent duplicates", in: []int{1, 2, 1, 3, 2}, want: []int{1, 2, 3}},
		{name: "every value the same", in: []int{7, 7, 7}, want: []int{7}},
		{name: "zero value is not treated as absent", in: []int{0, 0, 1}, want: []int{0, 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, slices.Uniq(tt.in))
		})
	}
}

func TestUniqLeavesInputAlone(t *testing.T) {
	t.Parallel()

	in := []int{3, 1, 3, 2, 1}

	got := slices.Uniq(in)

	require.Equal(t, []int{3, 1, 2}, got)
	require.Equal(t, []int{3, 1, 3, 2, 1}, in, "Uniq modified the slice it was given")
}

func TestUniqDoesNotAliasInput(t *testing.T) {
	t.Parallel()

	in := []int{1, 2, 3}

	got := slices.Uniq(in)
	require.Equal(t, in, got)

	got[0] = 99

	require.Equal(t, []int{1, 2, 3}, in, "the returned slice shares a backing array with the input")
}

func TestUniqComparableStructs(t *testing.T) {
	t.Parallel()

	type point struct{ x, y int }

	in := []point{{x: 1, y: 2}, {x: 3, y: 4}, {x: 1, y: 2}}

	require.Equal(t, []point{{x: 1, y: 2}, {x: 3, y: 4}}, slices.Uniq(in))
}

// member holds a slice, so it is not comparable. Keeping it that way means this
// file stops compiling if UniqBy's T is ever narrowed back to comparable.
type member struct {
	Name string
	Tags []string
}

func memberName(m member) string { return m.Name }

func TestUniqBy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []member
		want []member
	}{
		{name: "nil", in: nil, want: nil},
		{name: "empty slice returns nil", in: []member{}, want: nil},
		{
			name: "all keys distinct",
			in:   []member{{Name: "ada"}, {Name: "grace"}},
			want: []member{{Name: "ada"}, {Name: "grace"}},
		},
		{
			name: "same key keeps the first value",
			in:   []member{{Name: "ada", Tags: []string{"first"}}, {Name: "ada", Tags: []string{"second"}}},
			want: []member{{Name: "ada", Tags: []string{"first"}}},
		},
		{
			name: "non adjacent duplicate keys",
			in:   []member{{Name: "ada"}, {Name: "grace"}, {Name: "ada"}, {Name: "alan"}},
			want: []member{{Name: "ada"}, {Name: "grace"}, {Name: "alan"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, slices.UniqBy(tt.in, memberName))
		})
	}
}

func TestUniqByLeavesInputAlone(t *testing.T) {
	t.Parallel()

	in := []member{{Name: "ada"}, {Name: "grace"}, {Name: "ada"}}

	got := slices.UniqBy(in, memberName)

	require.Len(t, got, 2)
	require.Equal(t, []member{{Name: "ada"}, {Name: "grace"}, {Name: "ada"}}, in,
		"UniqBy modified the slice it was given")
}

func TestUniqByDoesNotAliasInput(t *testing.T) {
	t.Parallel()

	in := []member{{Name: "ada"}, {Name: "grace"}}

	got := slices.UniqBy(in, memberName)
	require.Equal(t, in, got)

	got[0].Name = "changed"

	require.Equal(t, "ada", in[0].Name, "the returned slice shares a backing array with the input")
}

func ExampleUniq() {
	fmt.Println(slices.Uniq([]string{"ada", "grace", "ada", "alan", "grace"}))
	// Output: [ada grace alan]
}

func ExampleUniqBy() {
	type account struct {
		Name  string
		Email string
	}

	accounts := []account{
		{Name: "ada", Email: "ada@example.com"},
		{Name: "grace", Email: "grace@example.com"},
		{Name: "ada", Email: "ada@work.example.com"},
	}

	for _, a := range slices.UniqBy(accounts, func(a account) string { return a.Name }) {
		fmt.Println(a.Name, a.Email)
	}
	// Output:
	// ada ada@example.com
	// grace grace@example.com
}
