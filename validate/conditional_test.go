package validate_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pseudomuto/go/validate"
)

func TestWhen(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cond bool
		want []string
	}{
		{name: "true runs the guarded rules", cond: true, want: []string{"user.email", "user.phone"}},
		{name: "false contributes nothing", cond: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validate.Validate("user", validate.When(tt.cond,
				validate.Field("email", "", validate.Required[string]()),
				validate.Field("phone", "", validate.Required[string]()),
			))

			if tt.want == nil {
				require.NoError(t, err)
				return
			}

			require.Equal(t, tt.want, pathsOf(t, err))
		})
	}
}

// When is built on Group("", rules...) precisely so a skipped condition yields a
// real Rule rather than a nil one that panics when Validate calls it.
func TestWhenWithNothingToDoIsStillARule(t *testing.T) {
	t.Parallel()

	require.NoError(t, validate.Validate("user", validate.When(true)))
	require.NoError(t, validate.Validate("user", validate.When(false)))

	require.NotPanics(t, func() { _ = validate.When(false)() })
	require.NotPanics(t, func() { _ = validate.When(true)() })
}

func TestWhenAddsNoPathSegment(t *testing.T) {
	t.Parallel()

	guarded := validate.Validate("user", validate.Group("address",
		validate.When(true, validate.Field("street", "", validate.Required[string]())),
	))

	plain := validate.Validate("user", validate.Group("address",
		validate.Field("street", "", validate.Required[string]()),
	))

	require.Equal(t, plain, guarded, "a guarded rule should report the path it would report unguarded")
	require.Equal(t, []string{"user.address.street"}, pathsOf(t, guarded))
}

func TestWhenFalseDoesNotRunTheGuardedChecks(t *testing.T) {
	t.Parallel()

	calls := 0
	count := func(string) error {
		calls++
		return nil
	}

	require.NoError(t, validate.Validate("user", validate.When(false, validate.Field("x", "", count))))
	require.Zero(t, calls, "a skipped When still ran its checks")

	require.NoError(t, validate.Validate("user", validate.When(true, validate.Field("x", "", count))))
	require.Equal(t, 1, calls)
}

func TestWhenLeavesItsSiblingsAlone(t *testing.T) {
	t.Parallel()

	req := validate.Required[string]()

	err := validate.Validate("user",
		validate.Field("a", "", req),
		validate.When(false, validate.Field("skipped", "", req)),
		validate.Field("b", "", req),
		validate.When(true, validate.Field("c", "", req)),
		validate.Field("d", "", req),
	)

	require.Equal(t, []string{"user.a", "user.b", "user.c", "user.d"}, pathsOf(t, err),
		"a false When should drop only its own rules, and declaration order should hold")
}

func TestWhenNests(t *testing.T) {
	t.Parallel()

	req := validate.Required[string]()

	t.Run("inside another When", func(t *testing.T) {
		t.Parallel()

		run := func(outer, inner bool) error {
			return validate.Validate("user", validate.When(outer,
				validate.When(inner, validate.Field("x", "", req)),
			))
		}

		require.EqualError(t, run(true, true), "user.x: is required")
		require.NoError(t, run(true, false))
		require.NoError(t, run(false, true))
		require.NoError(t, run(false, false))
	})

	t.Run("inside a Group, which still names it", func(t *testing.T) {
		t.Parallel()

		err := validate.Validate("user", validate.Group("address",
			validate.When(true, validate.Field("street", "", req)),
		))

		require.Equal(t, []string{"user.address.street"}, pathsOf(t, err))
	})

	t.Run("guarding a Group", func(t *testing.T) {
		t.Parallel()

		err := validate.Validate("user", validate.When(true,
			validate.Group("address", validate.Field("street", "", req)),
		))

		require.Equal(t, []string{"user.address.street"}, pathsOf(t, err))
	})
}

func TestWhenIsReusable(t *testing.T) {
	t.Parallel()

	rule := validate.When(true, validate.Field("street", "", validate.Required[string]()))

	require.EqualError(t, validate.Validate("a", rule), "a.street: is required")
	require.EqualError(t, validate.Validate("b", rule), "b.street: is required")
	require.EqualError(t, validate.Validate("a", validate.Group("g", rule)), "a.g.street: is required")
	require.EqualError(t, validate.Validate("a", rule), "a.street: is required",
		"a previous run left something behind")
}

func TestUnless(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cond bool
		want []string
	}{
		{name: "false runs the guarded rules", cond: false, want: []string{"user.quota"}},
		{name: "true contributes nothing", cond: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validate.Validate("user", validate.Unless(tt.cond,
				validate.Field("quota", 0, validate.GT(0)),
			))

			if tt.want == nil {
				require.NoError(t, err)
				return
			}

			require.Equal(t, tt.want, pathsOf(t, err))
		})
	}
}

func TestUnlessIsWhenNegated(t *testing.T) {
	t.Parallel()

	for _, cond := range []bool{true, false} {
		unless := validate.Validate("user", validate.Unless(cond,
			validate.Field("q", "", validate.Required[string]()),
		))

		when := validate.Validate("user", validate.When(!cond,
			validate.Field("q", "", validate.Required[string]()),
		))

		require.Equal(t, when, unless, "Unless(%t) and When(%t) disagree", cond, !cond)
	}
}

func TestWhenFunc(t *testing.T) {
	t.Parallel()

	t.Run("true builds and runs", func(t *testing.T) {
		t.Parallel()

		err := validate.Validate("user", validate.WhenFunc(true, func() validate.Rule {
			return validate.Field("email", "", validate.Required[string]())
		}))

		require.Equal(t, []string{"user.email"}, pathsOf(t, err))
	})

	t.Run("false never calls build", func(t *testing.T) {
		t.Parallel()

		called := false

		err := validate.Validate("user", validate.WhenFunc(false, func() validate.Rule {
			called = true
			return validate.Field("email", "", validate.Required[string]())
		}))

		require.NoError(t, err)
		require.False(t, called, "build ran despite a false condition")
	})

	t.Run("several rules via a Group", func(t *testing.T) {
		t.Parallel()

		err := validate.Validate("user", validate.WhenFunc(true, func() validate.Rule {
			return validate.Group("",
				validate.Field("email", "", validate.Required[string]()),
				validate.Field("phone", "", validate.Required[string]()),
			)
		}))

		require.Equal(t, []string{"user.email", "user.phone"}, pathsOf(t, err))
	})
}

func TestWhenFuncDefersUntilTheRuleRuns(t *testing.T) {
	t.Parallel()

	calls := 0
	rule := validate.WhenFunc(true, func() validate.Rule {
		calls++
		return validate.Field("email", "", validate.Required[string]())
	})

	require.Zero(t, calls, "building the Rule already called build")

	require.Error(t, validate.Validate("user", rule))
	require.Equal(t, 1, calls)

	require.Error(t, validate.Validate("user", rule))
	require.Equal(t, 2, calls, "build should run once per run, the same way Each's callback does")
}

// The reason WhenFunc exists: plain When would evaluate the argument at the call
// site, before it ever saw the condition.
func TestWhenFuncGuardsADereference(t *testing.T) {
	t.Parallel()

	type profile struct{ Bio string }

	var missing *profile

	rule := validate.WhenFunc(missing != nil, func() validate.Rule {
		return validate.Field("bio", missing.Bio, validate.Required[string]())
	})

	require.NotPanics(t, func() {
		require.NoError(t, validate.Validate("user", rule))
	})
}

func TestWhenFuncAddsNoPathSegment(t *testing.T) {
	t.Parallel()

	err := validate.Validate("user", validate.Group("address",
		validate.WhenFunc(true, func() validate.Rule {
			return validate.Field("street", "", validate.Required[string]())
		}),
	))

	require.Equal(t, []string{"user.address.street"}, pathsOf(t, err))
}

func ExampleWhen() {
	notify, email, phone := true, "", "555-0100"

	err := validate.Validate("user",
		validate.Field("name", "ada", validate.Required[string]()),
		validate.When(notify,
			validate.Field("email", email, validate.Required[string]()),
			validate.Field("phone", phone, validate.Required[string]()),
		),
	)

	fmt.Println(err)
	// Output: user.email: is required
}

func ExampleUnless() {
	isAdmin := false

	err := validate.Validate("user",
		validate.Field("name", "ada", validate.Required[string]()),
		validate.Unless(isAdmin, validate.Field("quota", 0, validate.GT(0))),
	)

	fmt.Println(err)
	// Output: user.quota: not greater than 0
}

func ExampleWhenFunc() {
	type profile struct{ Bio string }

	var missing *profile

	// Plain When would evaluate missing.Bio at the call site and panic. The
	// closure only runs once the condition holds.
	err := validate.Validate("user",
		validate.WhenFunc(missing != nil, func() validate.Rule {
			return validate.Field("bio", missing.Bio, validate.Required[string]())
		}),
	)

	fmt.Println(err)
	// Output: <nil>
}
