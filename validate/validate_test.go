package validate_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pseudomuto/go/validate"
)

type (
	user struct {
		Name     string
		Age      int
		Address  address
		Contacts []address
		Tags     []string
	}

	address struct {
		Street string
		Zip    string
	}

	// rootedAddress names itself, the way a type written for standalone use
	// would. address deliberately does not, so both styles are covered.
	rootedAddress struct {
		Street string
	}

	// ptrAddress declares Validate on a pointer receiver, so a slice of these
	// has to be a slice of pointers. Its nil guard is the one Nested's doc asks
	// a pointer type for.
	ptrAddress struct {
		Zip string
	}

	// validatorFunc adapts a function to [validate.Validator], so a test can
	// return whatever shape it likes from Validate.
	validatorFunc func() error
)

func TestValidateReturnsNilWhenNothingFails(t *testing.T) {
	t.Parallel()

	t.Run("no rules at all", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, validate.Validate("user"))
	})

	t.Run("rules that all pass", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, validate.Validate("user",
			validate.Field("name", "ada", validate.Required[string]()),
			validate.Group("address", validate.Field("zip", "M4B", validate.Required[string]())),
		))
	})

	t.Run("a rule with no checks", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, validate.Validate("user", validate.Field("name", "")))
	})

	t.Run("a group with no rules", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, validate.Validate("user", validate.Group("address")))
	})

	t.Run("the nil is a true nil, not a typed one", func(t *testing.T) {
		t.Parallel()

		// A non-nil interface holding a nil Errors would slip past this.
		err := validate.Validate("user", validate.Field("name", "ada", validate.Required[string]()))
		require.True(t, err == nil, "got a non-nil error interface: %#v", err)
	})
}

func TestValidateReportsEveryFailure(t *testing.T) {
	t.Parallel()

	err := user{Age: 200, Tags: []string{"a", "a"}}.Validate()
	require.EqualError(t, err, "user.name: is required\n"+
		"user.age: not less than 150\n"+
		"user.address.street: is required\n"+
		"user.address.zip: is required\n"+
		"user.tags: contains duplicate value: a")
}

func TestValidateKeepsDeclarationOrder(t *testing.T) {
	t.Parallel()

	err := validate.Validate("ns",
		validate.Field("c", "", validate.Required[string]()),
		validate.Group("g", validate.Field("b", "", validate.Required[string]())),
		validate.Field("a", 5, validate.LT(1), validate.LT(2)),
	)

	require.Equal(t, []string{"ns.c", "ns.g.b", "ns.a", "ns.a"}, pathsOf(t, err),
		"rules should run in the order given, and checks within a rule likewise")
}

// The core of the model: one segment per level, outermost first, nothing merged
// or special-cased.
func TestPathsAreBuiltByPrepending(t *testing.T) {
	t.Parallel()

	req := validate.Required[string]()

	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "field under a namespace",
			err:  validate.Validate("user", validate.Field("name", "", req)),
			want: "user.name",
		},
		{
			name: "one group",
			err:  validate.Validate("user", validate.Group("address", validate.Field("street", "", req))),
			want: "user.address.street",
		},
		{
			name: "groups nest arbitrarily deep",
			err: validate.Validate("a",
				validate.Group("b", validate.Group("c", validate.Group("d", validate.Field("e", "", req))))),
			want: "a.b.c.d.e",
		},
		{
			name: "an empty namespace contributes nothing",
			err:  validate.Validate("", validate.Group("address", validate.Field("street", "", req))),
			want: "address.street",
		},
		{
			name: "an empty group name contributes nothing",
			err:  validate.Validate("user", validate.Group("", validate.Field("name", "", req))),
			want: "user.name",
		},
		{
			name: "an empty field name contributes nothing",
			err:  validate.Validate("user", validate.Group("address", validate.Field("", "", req))),
			want: "user.address",
		},
		{
			name: "nothing named anywhere leaves an empty path",
			err:  validate.Validate("", validate.Field("", "", req)),
			want: "",
		},
		{
			name: "repeated names both appear rather than merging",
			err:  validate.Validate("user", validate.Group("user", validate.Field("user", "", req))),
			want: "user.user.user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, []string{tt.want}, pathsOf(t, tt.err))
		})
	}
}

// Validate is documented as a Group at the root, so the two had better agree.
func TestValidateIsGroupAtTheRoot(t *testing.T) {
	t.Parallel()

	rules := []validate.Rule{
		validate.Field("name", "", validate.Required[string]()),
		validate.Group("address", validate.Field("zip", "", validate.Required[string]())),
	}

	fromValidate := validate.Validate("user", rules...)
	fromGroup := validate.Group("user", rules...)()

	require.Equal(t, validate.Errors(fromGroup), fromValidate)
}

func TestValidateNesting(t *testing.T) {
	t.Parallel()

	street := validate.Field("street", "", validate.Required[string]())

	t.Run("a nested Validate stacks its namespace rather than replacing one", func(t *testing.T) {
		t.Parallel()

		err := validate.Validate("user", validate.Field("address", "", func(string) error {
			return validate.Validate("address", street)
		}))

		// Four levels named something, so four segments. The two "address"
		// segments are the outer Field and the inner Validate, not a bug.
		require.EqualError(t, err, "user.address.address.street: is required")
	})

	t.Run("an inner Validate with no namespace reads as expected", func(t *testing.T) {
		t.Parallel()

		err := validate.Validate("user", validate.Field("address", "", func(string) error {
			return validate.Validate("", street)
		}))

		require.EqualError(t, err, "user.address.street: is required")
	})

	t.Run("every nested failure comes through", func(t *testing.T) {
		t.Parallel()

		err := validate.Validate("user",
			validate.Field("address", "", func(string) error {
				return validate.Validate("", street,
					validate.Field("zip", "", validate.Required[string]()))
			}),
			validate.Field("name", "", validate.Required[string]()),
		)

		require.Equal(t, []string{"user.address.street", "user.address.zip", "user.name"}, pathsOf(t, err))
	})
}

// The module-wide "no surprise mutation" rule applies to what a Check hands
// back, too: a check returning a shared Errors should not find it rewritten
// afterwards.
func TestValidateLeavesTheCheckErrorsAlone(t *testing.T) {
	t.Parallel()

	shared := validate.Errors{validate.Fail("boom")}

	err := validate.Validate("ns", validate.Field("field", "", func(string) error { return shared }))

	require.EqualError(t, err, "ns.field: boom")
	require.Equal(t, validate.Errors{validate.Fail("boom")}, shared,
		"Validate wrote a path back into the check's own Errors")
}

func TestValidateErrorSupportsErrorsAs(t *testing.T) {
	t.Parallel()

	err := user{Age: -1}.Validate()

	var verrs validate.Errors
	require.True(t, errors.As(err, &verrs), "the whole list should be reachable")
	require.Len(t, verrs, 4)
	require.Equal(t, "user.name", verrs[0].Path())
	require.Equal(t, "user", verrs[0].Namespace())
	require.Equal(t, "name", verrs[0].Field())

	var verr validate.Error
	require.True(t, errors.As(err, &verr), "a single failure should be reachable")
	require.Equal(t, "user.name", verr.Path(), "errors.As should land on the first failure")
}

func TestValidateErrorSupportsErrorsIs(t *testing.T) {
	t.Parallel()

	err := validate.Validate("user", validate.Group("address",
		validate.Field("zip", "", validate.Required[string]()),
	))

	require.ErrorIs(t, err, validate.Fail("is required", "user", "address", "zip"))
	require.NotErrorIs(t, err, validate.Fail("is required", "user", "zip"),
		"a shorter path is a different failure")
	require.NotErrorIs(t, err, validate.Fail("is invalid", "user", "address", "zip"))
}

func TestValidateHandlesHeterogeneousFields(t *testing.T) {
	t.Parallel()

	type record struct {
		Name    string
		Count   int
		Ratio   float64
		Enabled bool
	}

	r := record{}

	err := validate.Validate("record",
		validate.Field("name", r.Name, validate.Required[string]()),
		validate.Field("count", r.Count, validate.GT(0)),
		validate.Field("ratio", r.Ratio, validate.GT(0.0), validate.LTE(1.0)),
		validate.Field("enabled", r.Enabled, validate.Required[bool]()),
	)

	require.EqualError(t, err, "record.name: is required\n"+
		"record.count: not greater than 0\n"+
		"record.ratio: not greater than 0\n"+
		"record.enabled: is required")
}

func TestValidateReportsIndexedSubObjects(t *testing.T) {
	t.Parallel()

	u := user{
		Name:     "ada",
		Address:  address{Street: "1 Main", Zip: "M4B 1B3"},
		Contacts: []address{{Street: "2 Oak", Zip: "M4B 1B4"}, {Street: "3 Elm"}},
	}

	require.EqualError(t, u.Validate(), "user.contacts[1].zip: is required")
}

// pathsOf pulls the path out of every failure in err, in order.
func pathsOf(t *testing.T, err error) []string {
	t.Helper()

	var verrs validate.Errors
	require.True(t, errors.As(err, &verrs), "expected a validate.Errors, got %v", err)

	paths := make([]string, len(verrs))
	for i, e := range verrs {
		paths[i] = e.Path()
	}

	return paths
}

func ExampleValidate() {
	type account struct {
		Handle string
		Age    int
	}

	a := account{Age: -3}

	err := validate.Validate("account",
		validate.Field("handle", a.Handle, validate.Required[string]()),
		validate.Field("age", a.Age, validate.GTE(0), validate.LT(150)),
	)

	fmt.Println(err)
	// Output:
	// account.handle: is required
	// account.age: not greater than or equal to 0
}

func (a address) Rules() validate.Rule {
	return validate.Group("",
		validate.Field("street", a.Street, validate.Required[string]()),
		validate.Field("zip", a.Zip, validate.Required[string]()),
	)
}

// Validate makes address a [validate.Validator]. It stays unrooted, so whoever
// nests it picks the name.
func (a address) Validate() error {
	return validate.Validate("", a.Rules())
}

func (r rootedAddress) Validate() error {
	return validate.Validate("address",
		validate.Field("street", r.Street, validate.Required[string]()),
	)
}

func (p *ptrAddress) Validate() error {
	if p == nil {
		return nil
	}

	return validate.Validate("", validate.Field("zip", p.Zip, validate.Required[string]()))
}

func (f validatorFunc) Validate() error { return f() }

func (u user) Validate() error {
	return validate.Validate("user",
		validate.Field("name", u.Name, validate.Required[string]()),
		validate.Field("age", u.Age, validate.GTE(0), validate.LT(150)),
		validate.Nested("address", u.Address),
		validate.EachNested("contacts", u.Contacts),
		validate.Field("tags", u.Tags, validate.Unique[string]()),
	)
}
