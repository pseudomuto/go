package validate_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pseudomuto/go/validate"
)

func TestFail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		msg      string
		path     []string
		wantPath string
		wantMsg  string
	}{
		{name: "no path", msg: "is required", wantPath: "", wantMsg: "is required"},
		{
			name:     "one segment",
			msg:      "is required",
			path:     []string{"name"},
			wantPath: "name",
			wantMsg:  "name: is required",
		},
		{
			name:     "several segments",
			msg:      "is required",
			path:     []string{"user", "address", "street"},
			wantPath: "user.address.street",
			wantMsg:  "user.address.street: is required",
		},
		{
			name:     "empty segments are dropped",
			msg:      "is required",
			path:     []string{"", "user", "", "name", ""},
			wantPath: "user.name",
			wantMsg:  "user.name: is required",
		},
		{
			name:     "every segment empty is the same as no path",
			msg:      "is required",
			path:     []string{"", ""},
			wantPath: "",
			wantMsg:  "is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := validate.Fail(tt.msg, tt.path...)

			require.Equal(t, tt.wantPath, e.Path())
			require.EqualError(t, e, tt.wantMsg)
		})
	}
}

func TestFailDoesNotRetainThePathSlice(t *testing.T) {
	t.Parallel()

	path := []string{"user", "name"}

	e := validate.Fail("is required", path...)
	path[0] = "clobbered"

	require.Equal(t, "user.name", e.Path())
}

func TestErrorPathAccessors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    []string
		wantNS  string
		wantFld string
	}{
		{name: "empty path", path: nil, wantNS: "", wantFld: ""},
		{name: "one segment is all field", path: []string{"name"}, wantNS: "", wantFld: "name"},
		{name: "two segments split", path: []string{"user", "name"}, wantNS: "user", wantFld: "name"},
		{
			name:    "deep path keeps everything above the leaf",
			path:    []string{"user", "address", "street"},
			wantNS:  "user.address",
			wantFld: "street",
		},
		{
			name:    "repeated segments are not collapsed",
			path:    []string{"user", "address", "address", "street"},
			wantNS:  "user.address.address",
			wantFld: "street",
		},
		{
			// A segment holding a dot is just two segments; nothing quotes it.
			name:    "a dotted segment splits like any other",
			path:    []string{"user.address", "street"},
			wantNS:  "user.address",
			wantFld: "street",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := validate.Fail("boom", tt.path...)

			require.Equal(t, tt.wantNS, e.Namespace())
			require.Equal(t, tt.wantFld, e.Field())
		})
	}
}

func TestErrorNamespaceAndFieldRejoinIntoPath(t *testing.T) {
	t.Parallel()

	for _, path := range [][]string{nil, {"name"}, {"user", "name"}, {"user", "address", "street"}} {
		e := validate.Fail("boom", path...)

		rejoined := e.Field()
		if ns := e.Namespace(); ns != "" {
			rejoined = ns + "." + e.Field()
		}

		require.Equal(t, e.Path(), rejoined, "Namespace and Field do not reconstruct Path for %v", path)
	}
}

func TestErrorFormatting(t *testing.T) {
	t.Parallel()

	require.EqualError(t, validate.Fail("is required", "user", "name"), "user.name: is required")
	require.EqualError(t, validate.Fail("is required", "name"), "name: is required")
	require.EqualError(t, validate.Fail("is required"), "is required",
		"an unplaced failure should not be prefixed with a bare colon")
}

func TestErrorIsComparedByPathAndMessage(t *testing.T) {
	t.Parallel()

	e := validate.Fail("is required", "user", "name")

	require.ErrorIs(t, e, validate.Fail("is required", "user", "name"))
	require.ErrorIs(t, e, validate.Fail("is required", "user.name"),
		"the path is what matters, not how it was segmented")
	require.NotErrorIs(t, e, validate.Fail("is required", "acct", "name"))
	require.NotErrorIs(t, e, validate.Fail("is required", "user", "email"))
	require.NotErrorIs(t, e, validate.Fail("is invalid", "user", "name"))
}

// Error stays a comparable struct, which is what lets errors.Is match by == and
// what lets callers key failures in a map.
func TestErrorIsComparable(t *testing.T) {
	t.Parallel()

	seen := map[validate.Error]int{}
	seen[validate.Fail("is required", "user", "name")]++
	seen[validate.Fail("is required", "user", "name")]++
	seen[validate.Fail("is required", "user", "email")]++

	require.Equal(t, 2, seen[validate.Fail("is required", "user", "name")])
	require.Len(t, seen, 2)
}

func TestErrorsError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   validate.Errors
		want string
	}{
		{name: "nil", in: nil, want: ""},
		{name: "empty", in: validate.Errors{}, want: ""},
		{
			name: "one failure",
			in:   validate.Errors{validate.Fail("is required", "user", "name")},
			want: "user.name: is required",
		},
		{
			name: "several are joined one per line, in order",
			in: validate.Errors{
				validate.Fail("is required", "user", "name"),
				validate.Fail("not greater than 0", "user", "age"),
			},
			want: "user.name: is required\nuser.age: not greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, tt.in.Error())
		})
	}
}

func TestErrorsUnwrap(t *testing.T) {
	t.Parallel()

	t.Run("yields every failure, in order", func(t *testing.T) {
		t.Parallel()

		ve := validate.Errors{
			validate.Fail("is required", "user", "name"),
			validate.Fail("not greater than 0", "user", "age"),
		}

		require.Equal(t, []error{ve[0], ve[1]}, ve.Unwrap())
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()

		require.Empty(t, validate.Errors(nil).Unwrap())
	})

	t.Run("the slice is fresh", func(t *testing.T) {
		t.Parallel()

		ve := validate.Errors{validate.Fail("is required", "user", "name")}

		got := ve.Unwrap()
		got[0] = errors.New("clobbered")

		require.Equal(t, validate.Fail("is required", "user", "name"), ve[0],
			"writing to Unwrap's result reached back into the Errors")
		require.NotSame(t, &ve.Unwrap()[0], &got[0], "two calls handed back the same backing array")
	})
}

func TestErrorsWorksWithErrorsIsAndAs(t *testing.T) {
	t.Parallel()

	ve := validate.Errors{
		validate.Fail("is required", "user", "name"),
		validate.Fail("not greater than 0", "user", "age"),
	}

	t.Run("Is finds any element", func(t *testing.T) {
		t.Parallel()

		require.ErrorIs(t, ve, ve[0])
		require.ErrorIs(t, ve, ve[1])
		require.NotErrorIs(t, ve, validate.Fail("is required", "user", "email"))
	})

	t.Run("As lands on the first element", func(t *testing.T) {
		t.Parallel()

		var e validate.Error
		require.True(t, errors.As(error(ve), &e))
		require.Equal(t, "name", e.Field())
	})

	t.Run("As reaches through a wrapper", func(t *testing.T) {
		t.Parallel()

		wrapped := fmt.Errorf("saving user: %w", ve)

		var got validate.Errors
		require.True(t, errors.As(wrapped, &got))
		require.Equal(t, ve, got)
	})
}

func ExampleFail() {
	// A check that looks inside the value it was handed can name what it found,
	// relative to its own field.
	hasZip := func(addr map[string]string) error {
		if addr["zip"] == "" {
			return validate.Fail("is required", "zip")
		}

		return nil
	}

	fmt.Println(validate.Validate("user", validate.Field("address", map[string]string{}, hasZip)))
	// Output: user.address.zip: is required
}

func ExampleError() {
	e := validate.Fail("is not a valid address", "user", "contact", "email")

	fmt.Println(e.Path())
	fmt.Println(e.Namespace(), "|", e.Field())
	fmt.Println(e)
	// Output:
	// user.contact.email
	// user.contact | email
	// user.contact.email: is not a valid address
}

func ExampleErrors() {
	err := validate.Validate("user",
		validate.Field("name", "", validate.Required[string]()),
		validate.Group("address", validate.Field("zip", "", validate.Required[string]())),
	)

	// The failures are structured, so a caller can key them by path rather than
	// parse the message.
	var verrs validate.Errors
	if errors.As(err, &verrs) {
		for _, e := range verrs {
			fmt.Printf("%s -> %s\n", e.Path(), e)
		}
	}
	// Output:
	// user.name -> user.name: is required
	// user.address.zip -> user.address.zip: is required
}
