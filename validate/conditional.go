package validate

// When runs rules only when cond holds, and contributes nothing when it does
// not.
//
// It is a [Group] with an empty name, so it adds no path segment of its own. A
// field guarded by When reports exactly the path it would report unguarded:
//
//	validate.Validate("user",
//		validate.Field("name", u.Name, validate.Required[string]()),
//		validate.When(u.Notify,
//			validate.Field("email", u.Email, validate.Required[string]()),
//			validate.Field("phone", u.Phone, validate.Required[string]()),
//		),
//	)
//
//	// user.name: is required
//	// user.email: is required   <- only when u.Notify
//
// A false condition drops the rules it guards and nothing else. The rules around
// it still run, and so do the checks inside the ones that are not guarded:
// nothing here short-circuits.
//
// cond is evaluated where you write it, like every other value this package
// binds. So are the rules, even when cond turns out false, which matters when
// building one would panic:
//
//	// panics on a nil Address: *u.Address is evaluated before When is called
//	validate.When(u.Address != nil, validate.Nested("address", *u.Address))
//
//	// fine: a pointer satisfies [Validator], so nothing is dereferenced here
//	validate.When(u.Address != nil, validate.Nested("address", u.Address))
//
// The second form is the usual one. Reach for [WhenFunc] when the argument
// really does have to be dereferenced or computed.
//
// When with no rules, and When with a false condition, are both still usable
// rules. They contribute nothing rather than returning something unsafe to run.
func When(cond bool, rules ...Rule) Rule {
	if !cond {
		rules = nil
	}

	return Group("", rules...)
}

// Unless runs rules only when cond does not hold.
//
// It is [When] with the condition negated, for when the natural way to say it is
// negative and When(!cond, ...) would read backwards:
//
//	validate.Unless(u.IsAdmin,
//		validate.Field("quota", u.Quota, validate.GT(0)),
//	)
//
// Everything [When] documents applies here, including that the rules are built
// whether or not they end up running. There is no UnlessFunc; negate the
// condition and use [WhenFunc], which is the rare case anyway.
func Unless(cond bool, rules ...Rule) Rule {
	return When(!cond, rules...)
}

// WhenFunc is [When] for a rule that cannot be built safely unless cond holds.
//
// build is called only when cond is true, and not until the returned [Rule]
// runs. That is what makes it safe to reach through a pointer inside:
//
//	validate.WhenFunc(u.Address != nil, func() validate.Rule {
//		return validate.Field("street", u.Address.Street, validate.Required[string]())
//	})
//
// Plain [When] would evaluate u.Address.Street before it ever saw the condition.
// Most of the time you do not need this: pass the pointer rather than the value
// and When does the job.
//
// build runs once each time the rule runs, the same way [Each]'s callback does,
// so a rule reused at two depths builds twice. It must return a usable [Rule];
// for several, return a [Group] with an empty name, which is what When uses
// internally.
//
// Like When, WhenFunc adds no path segment.
func WhenFunc(cond bool, build func() Rule) Rule {
	return func() Errors {
		if !cond {
			return nil
		}

		return build()()
	}
}
