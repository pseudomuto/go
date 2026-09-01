package validate_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pseudomuto/go/validate"
)

func TestGroupNestsARuleWithoutRewritingIt(t *testing.T) {
	t.Parallel()

	a := address{}

	t.Run("the same rule composes at any depth", func(t *testing.T) {
		t.Parallel()

		require.EqualError(t, validate.Validate("address", a.Rules()),
			"address.street: is required\naddress.zip: is required")

		require.EqualError(t, validate.Validate("user", validate.Group("address", a.Rules())),
			"user.address.street: is required\nuser.address.zip: is required")

		require.EqualError(t, validate.Validate("user", validate.Group("home", validate.Group("addr", a.Rules()))),
			"user.home.addr.street: is required\nuser.home.addr.zip: is required")
	})

	t.Run("a group takes several rules", func(t *testing.T) {
		t.Parallel()

		err := validate.Validate("user", validate.Group("address",
			validate.Field("street", "", validate.Required[string]()),
			validate.Field("zip", "", validate.Required[string]()),
		))

		require.Equal(t, []string{"user.address.street", "user.address.zip"}, pathsOf(t, err))
	})

	t.Run("an empty name flattens rather than nests", func(t *testing.T) {
		t.Parallel()

		err := validate.Validate("user", validate.Group("",
			validate.Field("a", "", validate.Required[string]()),
			validate.Field("b", "", validate.Required[string]()),
		))

		require.Equal(t, []string{"user.a", "user.b"}, pathsOf(t, err))
	})
}

func TestFieldRunsEveryCheck(t *testing.T) {
	t.Parallel()

	calls := 0
	count := func(string) error {
		calls++
		return errors.New("nope")
	}

	err := validate.Validate("ns", validate.Field("name", "", count, count, count))

	require.Equal(t, 3, calls, "Field stopped at the first failing check")
	require.Len(t, pathsOf(t, err), 3)
}

func TestFieldDefersUntilValidateRuns(t *testing.T) {
	t.Parallel()

	called := false
	rule := validate.Field("name", "", func(string) error {
		called = true
		return nil
	})

	require.False(t, called, "building a Rule should not run its checks")

	require.NoError(t, validate.Validate("ns", rule))
	require.True(t, called)
}

func TestFieldPassesTheBoundValue(t *testing.T) {
	t.Parallel()

	var got []int

	err := validate.Validate("ns", validate.Field("n", 42, func(v int) error {
		got = append(got, v)
		return nil
	}, func(v int) error {
		got = append(got, v)
		return nil
	}))

	require.NoError(t, err)
	require.Equal(t, []int{42, 42}, got, "every check should see the value Field was given")
}

func TestFieldPlacesWhateverTheCheckReturns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		check validate.Check[string]
		want  []string
	}{
		{
			name:  "a plain error lands on the field",
			check: func(string) error { return errors.New("boom") },
			want:  []string{"ns.field"},
		},
		{
			name:  "an Error with no path of its own lands on the field",
			check: func(string) error { return validate.Fail("boom") },
			want:  []string{"ns.field"},
		},
		{
			name:  "an Error with a path lands below the field",
			check: func(string) error { return validate.Fail("boom", "zip") },
			want:  []string{"ns.field.zip"},
		},
		{
			name:  "a multi-segment path lands below the field too",
			check: func(string) error { return validate.Fail("boom", "inner", "zip") },
			want:  []string{"ns.field.inner.zip"},
		},
		{
			name:  "a wrapped Error is still found and placed",
			check: func(string) error { return fmt.Errorf("context: %w", validate.Fail("boom", "zip")) },
			want:  []string{"ns.field.zip"},
		},
		{
			name:  "an Errors contributes every element",
			check: func(string) error { return validate.Errors{validate.Fail("one", "a"), validate.Fail("two")} },
			want:  []string{"ns.field.a", "ns.field"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validate.Validate("ns", validate.Field("field", "", tt.check))
			require.Equal(t, tt.want, pathsOf(t, err))
		})
	}
}

// Placement is relative, so nothing a check returns can climb above the Field it
// was registered under.
func TestChecksCannotEscapeUpwards(t *testing.T) {
	t.Parallel()

	err := validate.Validate("user", validate.Group("address",
		validate.Field("zip", "", func(string) error {
			return validate.Fail("boom", "somewhere", "else")
		}),
	))

	require.EqualError(t, err, "user.address.zip.somewhere.else: boom")
}

// Likewise for a Rule: running it twice, or at two depths, must not accumulate.
func TestRulesAreReusable(t *testing.T) {
	t.Parallel()

	rule := validate.Field("street", "", validate.Required[string]())

	require.EqualError(t, validate.Validate("a", rule), "a.street: is required")
	require.EqualError(t, validate.Validate("b", rule), "b.street: is required")
	require.EqualError(t, validate.Validate("a", validate.Group("c", rule)), "a.c.street: is required")
	require.EqualError(t, validate.Validate("a", rule), "a.street: is required",
		"a previous run left something behind")
}

// The contract from Nested's godoc: the child's own root stacks like any other
// segment, so which name to pass depends on whether the child names itself.
func TestNestedPathDependsOnWhetherTheChildRootsItself(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "a self-rooting child nests under an empty name",
			err:  validate.Validate("user", validate.Nested("", rootedAddress{})),
			want: "user.address.street",
		},
		{
			name: "naming a self-rooting child stacks both names",
			err:  validate.Validate("user", validate.Nested("home", rootedAddress{})),
			want: "user.home.address.street",
		},
		{
			name: "an unrooted child is named by its parent",
			err:  validate.Validate("user", validate.Nested("address", address{Zip: "z"})),
			want: "user.address.street",
		},
		{
			name: "an unrooted child can be named anything",
			err:  validate.Validate("user", validate.Nested("home", address{Zip: "z"})),
			want: "user.home.street",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, []string{tt.want}, pathsOf(t, tt.err))
		})
	}
}

func TestNestedPlacesWhateverValidateReturns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ret  error
		want []string
	}{
		{
			name: "a plain error lands on the nested value itself",
			ret:  errors.New("exploded"),
			want: []string{"user.address"},
		},
		{
			name: "a bare Fail lands there too",
			ret:  validate.Fail("is malformed"),
			want: []string{"user.address"},
		},
		{
			name: "a Fail with a path lands below it",
			ret:  validate.Fail("is malformed", "zip"),
			want: []string{"user.address.zip"},
		},
		{
			name: "an Errors contributes every element",
			ret:  validate.Errors{validate.Fail("one", "x"), validate.Fail("two")},
			want: []string{"user.address.x", "user.address"},
		},
		{
			name: "a wrapped Error is still found",
			ret:  fmt.Errorf("context: %w", validate.Fail("boom", "zip")),
			want: []string{"user.address.zip"},
		},
		{
			name: "a nested Validate brings its own root along",
			ret:  validate.Validate("inner", validate.Field("f", "", validate.Required[string]())),
			want: []string{"user.address.inner.f"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v := validatorFunc(func() error { return tt.ret })
			require.Equal(t, tt.want, pathsOf(t, validate.Validate("user", validate.Nested("address", v))))
		})
	}
}

func TestNestedContributesNothingWhenTheChildPasses(t *testing.T) {
	t.Parallel()

	v := validatorFunc(func() error { return nil })

	require.NoError(t, validate.Validate("user", validate.Nested("address", v)))
}

func TestNestedDefersUntilTheRuleRuns(t *testing.T) {
	t.Parallel()

	calls := 0
	rule := validate.Nested("address", validatorFunc(func() error {
		calls++
		return nil
	}))

	require.Zero(t, calls, "building the Rule called Validate")

	require.NoError(t, validate.Validate("user", rule))
	require.Equal(t, 1, calls)
}

func TestNestedKeepsSameTypedFieldsApart(t *testing.T) {
	t.Parallel()

	err := validate.Validate("user",
		validate.Nested("home", address{Zip: "M4B 1B3"}),
		validate.Nested("work", address{Street: "1 Main"}),
	)

	require.Equal(t, []string{"user.home.street", "user.work.zip"}, pathsOf(t, err))
}

func TestNestedComposesAndIsReusable(t *testing.T) {
	t.Parallel()

	rule := validate.Nested("address", address{Zip: "M4B 1B3"})

	require.EqualError(t, validate.Validate("a", rule), "a.address.street: is required")
	require.EqualError(t, validate.Validate("b", rule), "b.address.street: is required")
	require.EqualError(t, validate.Validate("a", validate.Group("g", rule)), "a.g.address.street: is required")
	require.EqualError(t, validate.Validate("a", rule), "a.address.street: is required",
		"a previous run left something behind")
}

func TestEach(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{name: "nil slice contributes nothing", in: nil},
		{name: "empty slice contributes nothing", in: []string{}},
		{name: "one element, passing", in: []string{"a"}},
		{name: "one element, failing", in: []string{""}, want: []string{"user.tags[0]"}},
		{
			name: "only the failing elements are reported, by index",
			in:   []string{"a", "", "c", ""},
			want: []string{"user.tags[1]", "user.tags[3]"},
		},
		{
			name: "all failing, in order",
			in:   []string{"", "", ""},
			want: []string{"user.tags[0]", "user.tags[1]", "user.tags[2]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validate.Validate("user", validate.Each("tags", tt.in, func(s string) validate.Rule {
				return validate.Field("", s, validate.Required[string]())
			}))

			if tt.want == nil {
				require.NoError(t, err)
				return
			}

			require.Equal(t, tt.want, pathsOf(t, err))
		})
	}
}

func TestEachEmptyNameStillIndexes(t *testing.T) {
	t.Parallel()

	// Dropping the index would collapse every element onto one path, so the
	// segment is "[0]" rather than nothing.
	err := validate.Validate("user", validate.Each("", []string{"", ""}, func(s string) validate.Rule {
		return validate.Field("x", s, validate.Required[string]())
	}))

	require.Equal(t, []string{"user.[0].x", "user.[1].x"}, pathsOf(t, err))
}

func TestEachNests(t *testing.T) {
	t.Parallel()

	t.Run("over sub-objects", func(t *testing.T) {
		t.Parallel()

		contacts := []address{{Street: "1 Main", Zip: "M4B 1B3"}, {}, {Street: "2 Oak"}}

		err := validate.Validate("user", validate.Each("contacts", contacts, func(a address) validate.Rule {
			return validate.Nested("", a)
		}))

		require.Equal(t, []string{
			"user.contacts[1].street",
			"user.contacts[1].zip",
			"user.contacts[2].zip",
		}, pathsOf(t, err))
	})

	t.Run("inside another Each", func(t *testing.T) {
		t.Parallel()

		rows := [][]string{{"", "a"}, {"b", ""}}

		err := validate.Validate("m", validate.Each("rows", rows, func(row []string) validate.Rule {
			return validate.Each("col", row, func(s string) validate.Rule {
				return validate.Field("", s, validate.Required[string]())
			})
		}))

		require.Equal(t, []string{"m.rows[0].col[0]", "m.rows[1].col[1]"}, pathsOf(t, err))
	})

	t.Run("under a Group", func(t *testing.T) {
		t.Parallel()

		err := validate.Validate("user", validate.Group("profile",
			validate.Each("tags", []string{""}, func(s string) validate.Rule {
				return validate.Field("", s, validate.Required[string]())
			}),
		))

		require.Equal(t, []string{"user.profile.tags[0]"}, pathsOf(t, err))
	})
}

func TestEachRunsTheCallbackOncePerElementAndDefers(t *testing.T) {
	t.Parallel()

	calls := 0
	rule := validate.Each("tags", []string{"a", "b", "c"}, func(s string) validate.Rule {
		calls++
		return validate.Field("", s, validate.Required[string]())
	})

	require.Zero(t, calls, "building the Rule ran the callback")

	require.NoError(t, validate.Validate("user", rule))
	require.Equal(t, 3, calls, "the callback should run once per element")
}

func TestEachNested(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []address
		want []string
	}{
		{name: "nil slice contributes nothing", in: nil},
		{name: "empty slice contributes nothing", in: []address{}},
		{name: "every element valid", in: []address{{Street: "1 Main", Zip: "M4B 1B3"}}},
		{
			name: "one element, every failure reported",
			in:   []address{{}},
			want: []string{"user.contacts[0].street", "user.contacts[0].zip"},
		},
		{
			name: "only the failing elements, by index",
			in:   []address{{Street: "1 Main", Zip: "M4B 1B3"}, {Street: "2 Oak"}, {}},
			want: []string{"user.contacts[1].zip", "user.contacts[2].street", "user.contacts[2].zip"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validate.Validate("user", validate.EachNested("contacts", tt.in))

			if tt.want == nil {
				require.NoError(t, err)
				return
			}

			require.Equal(t, tt.want, pathsOf(t, err))
		})
	}
}

func TestEachNestedKeepsTheElementsOwnName(t *testing.T) {
	t.Parallel()

	// rootedAddress names itself, so it adds a segment beneath the index, just
	// as it does under Nested. The index alone names an unrooted element.
	rooted := validate.Validate("user", validate.EachNested("contacts", []rootedAddress{{}}))
	require.Equal(t, []string{"user.contacts[0].address.street"}, pathsOf(t, rooted))

	unrooted := validate.Validate("user", validate.EachNested("contacts", []address{{Zip: "z"}}))
	require.Equal(t, []string{"user.contacts[0].street"}, pathsOf(t, unrooted))
}

func TestEachNestedMatchesEachPlusNested(t *testing.T) {
	t.Parallel()

	in := []address{{Street: "1 Main"}, {}}

	spelledOut := validate.Validate("user", validate.Each("contacts", in, func(a address) validate.Rule {
		return validate.Nested("", a)
	}))

	require.Equal(t, spelledOut, validate.Validate("user", validate.EachNested("contacts", in)))
}

func TestEachNestedOnPointerElements(t *testing.T) {
	t.Parallel()

	// The nil element is safe only because ptrAddress.Validate guards its
	// receiver, which is exactly what Nested's doc asks of a pointer type.
	in := []*ptrAddress{{Zip: "M4B 1B3"}, {}, nil}

	err := validate.Validate("user", validate.EachNested("contacts", in))
	require.Equal(t, []string{"user.contacts[1].zip"}, pathsOf(t, err))
}

func TestEachNestedDefersUntilTheRuleRuns(t *testing.T) {
	t.Parallel()

	calls := 0
	v := validatorFunc(func() error {
		calls++
		return nil
	})

	rule := validate.EachNested("contacts", []validatorFunc{v, v, v})

	require.Zero(t, calls, "building the Rule called Validate")

	require.NoError(t, validate.Validate("user", rule))
	require.Equal(t, 3, calls, "Validate should run once per element")
}

func ExampleField() {
	// A Check is any func(T) error, so a one-off rule needs no ceremony.
	notReserved := func(name string) error {
		if name == "admin" {
			return fmt.Errorf("%q is reserved", name)
		}

		return nil
	}

	err := validate.Validate("account",
		validate.Field("handle", "admin", validate.Required[string](), notReserved),
	)

	fmt.Println(err)
	// Output: account.handle: "admin" is reserved
}

func ExampleGroup() {
	required := validate.Required[string]()

	err := validate.Validate("user",
		validate.Field("name", "", required),
		validate.Group("address",
			validate.Field("street", "", required),
			validate.Field("zip", "M4B 1B3", required),
		),
	)

	fmt.Println(err)
	// Output:
	// user.name: is required
	// user.address.street: is required
}

func ExampleGroup_reusable() {
	// A rule that does not name itself can be nested anywhere, or run alone.
	a := address{Zip: "M4B 1B3"}

	fmt.Println(validate.Validate("address", a.Rules()))
	fmt.Println(validate.Validate("user", validate.Group("home", a.Rules())))
	// Output:
	// address.street: is required
	// user.home.street: is required
}

func ExampleNested() {
	// address.Validate is unrooted, so whoever nests it picks the name.
	u := user{Address: address{Zip: "M4B 1B3"}}

	fmt.Println(validate.Validate("user", validate.Nested("address", u.Address)))
	// Output: user.address.street: is required
}

func ExampleNested_selfRooting() {
	// rootedAddress.Validate calls validate.Validate("address", ...), so it
	// names itself and nests cleanly under an empty name.
	fmt.Println(validate.Validate("user", validate.Nested("", rootedAddress{})))
	fmt.Println(validate.Validate("user", validate.Nested("home", rootedAddress{})))
	// Output:
	// user.address.street: is required
	// user.home.address.street: is required
}

func ExampleEach() {
	err := validate.Validate("user",
		validate.Each("tags", []string{"go", "", "fun", ""}, func(t string) validate.Rule {
			return validate.Field("", t, validate.Required[string]())
		}),
	)

	fmt.Println(err)
	// Output:
	// user.tags[1]: is required
	// user.tags[3]: is required
}

func ExampleEach_withExtraChecks() {
	// Each earns its callback when an element needs more than its own Validate.
	contacts := []address{{Street: "1 Main", Zip: "M4B 1B3"}, {Street: "2 Oak", Zip: "M4B 1B3"}}
	seen := map[string]bool{}

	err := validate.Validate("user",
		validate.Each("contacts", contacts, func(a address) validate.Rule {
			return validate.Group("",
				validate.Nested("", a),
				validate.Field("zip", a.Zip, func(zip string) error {
					if seen[zip] {
						return errors.New("is already used by another contact")
					}

					seen[zip] = true

					return nil
				}),
			)
		}),
	)

	fmt.Println(err)
	// Output: user.contacts[1].zip: is already used by another contact
}

func ExampleEachNested() {
	// address.Validate is unrooted, so the index is the only name an element
	// gets. Compare ExampleEach_withExtraChecks for the long way round.
	contacts := []address{{Street: "1 Main", Zip: "M4B 1B3"}, {Street: "2 Oak"}, {}}

	fmt.Println(validate.Validate("user", validate.EachNested("contacts", contacts)))
	// Output:
	// user.contacts[1].zip: is required
	// user.contacts[2].street: is required
	// user.contacts[2].zip: is required
}
