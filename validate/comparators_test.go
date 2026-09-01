package validate_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pseudomuto/go/validate"
)

func TestGT(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mark    int
		in      int
		wantErr string
	}{
		{name: "above the mark passes", mark: 0, in: 1},
		{name: "far above the mark passes", mark: 0, in: 1000},
		{name: "the mark itself fails", mark: 0, in: 0, wantErr: "not greater than 0"},
		{name: "below the mark fails", mark: 0, in: -1, wantErr: "not greater than 0"},
		{name: "negative mark", mark: -10, in: -9},
		{name: "negative mark at boundary", mark: -10, in: -10, wantErr: "not greater than -10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requireCheck(t, validate.GT(tt.mark), tt.in, tt.wantErr)
		})
	}
}

func TestGTE(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mark    int
		in      int
		wantErr string
	}{
		{name: "above the mark passes", mark: 0, in: 1},
		{name: "the mark itself passes", mark: 0, in: 0},
		{name: "below the mark fails", mark: 0, in: -1, wantErr: "not greater than or equal to 0"},
		{name: "negative mark at boundary passes", mark: -10, in: -10},
		{name: "negative mark below fails", mark: -10, in: -11, wantErr: "not greater than or equal to -10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requireCheck(t, validate.GTE(tt.mark), tt.in, tt.wantErr)
		})
	}
}

func TestLT(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mark    int
		in      int
		wantErr string
	}{
		{name: "below the mark passes", mark: 10, in: 9},
		{name: "far below the mark passes", mark: 10, in: -1000},
		{name: "the mark itself fails", mark: 10, in: 10, wantErr: "not less than 10"},
		{name: "above the mark fails", mark: 10, in: 11, wantErr: "not less than 10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requireCheck(t, validate.LT(tt.mark), tt.in, tt.wantErr)
		})
	}
}

func TestLTE(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mark    int
		in      int
		wantErr string
	}{
		{name: "below the mark passes", mark: 10, in: 9},
		{name: "the mark itself passes", mark: 10, in: 10},
		{name: "above the mark fails", mark: 10, in: 11, wantErr: "not less than or equal to 10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requireCheck(t, validate.LTE(tt.mark), tt.in, tt.wantErr)
		})
	}
}

// The four comparators are only useful if they order every cmp.Ordered type the
// way that type is normally ordered, not just int.
func TestComparatorsSpanOrderedTypes(t *testing.T) {
	t.Parallel()

	t.Run("strings compare lexically", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, validate.GT("apple")("banana"))
		require.EqualError(t, validate.GT("banana")("apple"), "not greater than banana")
		require.NoError(t, validate.LTE("b")("b"))
		// "B" < "b", so case matters and is not normalised away.
		require.EqualError(t, validate.GT("b")("B"), "not greater than b")
	})

	t.Run("floats keep their fractional part", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, validate.GT(1.5)(1.51))
		require.EqualError(t, validate.GT(1.5)(1.5), "not greater than 1.5")
		require.NoError(t, validate.LT(1.5)(1.49))
	})

	t.Run("unsigned", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, validate.GTE(uint8(0))(0))
		require.EqualError(t, validate.LT(uint8(10))(255), "not less than 10")
	})
}

// GT/GTE and LT/LTE differ only at the boundary, so pin that down directly
// rather than trusting the tables above to have covered it.
func TestComparatorBoundariesDisagreeOnlyAtTheMark(t *testing.T) {
	t.Parallel()

	const mark = 5

	require.Error(t, validate.GT(mark)(mark), "GT should reject the mark")
	require.NoError(t, validate.GTE(mark)(mark), "GTE should accept the mark")
	require.Error(t, validate.LT(mark)(mark), "LT should reject the mark")
	require.NoError(t, validate.LTE(mark)(mark), "LTE should accept the mark")

	for _, v := range []int{mark - 1, mark + 1} {
		require.Equal(t, validate.GT(mark)(v) == nil, validate.GTE(mark)(v) == nil,
			"GT and GTE disagree away from the mark, at %d", v)
		require.Equal(t, validate.LT(mark)(v) == nil, validate.LTE(mark)(v) == nil,
			"LT and LTE disagree away from the mark, at %d", v)
	}
}

func TestRequired(t *testing.T) {
	t.Parallel()

	t.Run("strings", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, validate.Required[string]()("a"))
		require.NoError(t, validate.Required[string]()(" "), "whitespace is not the zero value")
		require.EqualError(t, validate.Required[string]()(""), "is required")
	})

	t.Run("numbers", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, validate.Required[int]()(1))
		require.NoError(t, validate.Required[int]()(-1))
		require.EqualError(t, validate.Required[int]()(0), "is required")
		require.EqualError(t, validate.Required[float64]()(0.0), "is required")
	})

	t.Run("bools cannot express absence", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, validate.Required[bool]()(true))
		require.EqualError(t, validate.Required[bool]()(false), "is required",
			"false is the zero value, so Required cannot tell it from unset")
	})

	t.Run("pointers", func(t *testing.T) {
		t.Parallel()

		zero := 0

		require.NoError(t, validate.Required[*int]()(&zero), "a pointer to zero is still set")
		require.EqualError(t, validate.Required[*int]()(nil), "is required")
	})

	t.Run("structs", func(t *testing.T) {
		t.Parallel()

		type point struct{ X, Y int }

		require.NoError(t, validate.Required[point]()(point{X: 1}))
		require.EqualError(t, validate.Required[point]()(point{}), "is required")
	})
}

func TestUnique(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      []string
		wantErr string
	}{
		{name: "nil slice", in: nil},
		{name: "empty slice", in: []string{}},
		{name: "single element", in: []string{"a"}},
		{name: "all distinct", in: []string{"a", "b", "c"}},
		{name: "adjacent duplicates", in: []string{"a", "a"}, wantErr: "contains duplicate value: a"},
		{
			name:    "non adjacent duplicates",
			in:      []string{"a", "b", "c", "a"},
			wantErr: "contains duplicate value: a",
		},
		{
			name:    "the zero value counts as a value",
			in:      []string{"", ""},
			wantErr: "contains duplicate value: ",
		},
		{
			name:    "reports the first duplicate reached, not the first duplicated value",
			in:      []string{"a", "b", "b", "a"},
			wantErr: "contains duplicate value: b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requireCheck(t, validate.Unique[string](), tt.in, tt.wantErr)
		})
	}
}

func TestUniqueLeavesInputAlone(t *testing.T) {
	t.Parallel()

	in := []int{3, 1, 3}

	require.Error(t, validate.Unique[int]()(in))
	require.Equal(t, []int{3, 1, 3}, in, "Unique reordered or rewrote the slice it was given")
}

func TestUniqueOnStructs(t *testing.T) {
	t.Parallel()

	type port struct {
		Host string
		Num  int
	}

	distinct := []port{{Host: "a", Num: 1}, {Host: "a", Num: 2}}
	require.NoError(t, validate.Unique[port]()(distinct), "structs differing in one field are distinct")

	same := []port{{Host: "a", Num: 1}, {Host: "a", Num: 1}}
	require.EqualError(t, validate.Unique[port]()(same), "contains duplicate value: {a 1}")
}

// requireCheck runs c against in and asserts on the message, where an empty
// wantErr means the value should pass.
func requireCheck[T any](t *testing.T, c validate.Check[T], in T, wantErr string) {
	t.Helper()

	err := c(in)
	if wantErr == "" {
		require.NoError(t, err)
		return
	}

	require.EqualError(t, err, wantErr)
}

func ExampleGT() {
	fmt.Println(validate.GT(0)(1))
	fmt.Println(validate.GT(0)(0))
	// Output:
	// <nil>
	// not greater than 0
}

func ExampleGTE() {
	fmt.Println(validate.GTE(0)(0))
	fmt.Println(validate.GTE(0)(-1))
	// Output:
	// <nil>
	// not greater than or equal to 0
}

func ExampleLT() {
	fmt.Println(validate.LT(10)(9))
	fmt.Println(validate.LT(10)(10))
	// Output:
	// <nil>
	// not less than 10
}

func ExampleLTE() {
	fmt.Println(validate.LTE(10)(10))
	fmt.Println(validate.LTE(10)(11))
	// Output:
	// <nil>
	// not less than or equal to 10
}

func ExampleRequired() {
	fmt.Println(validate.Required[string]()("ada"))
	fmt.Println(validate.Required[string]()(""))
	// Output:
	// <nil>
	// is required
}

func ExampleUnique() {
	fmt.Println(validate.Unique[string]()([]string{"a", "b"}))
	fmt.Println(validate.Unique[string]()([]string{"a", "b", "a"}))
	// Output:
	// <nil>
	// contains duplicate value: a
}
