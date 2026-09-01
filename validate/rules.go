package validate

import "strconv"

// Rule is one step of a validation, bound to the values it checks and ready to
// run.
//
// The constructors are [Field] for one value's checks, [Group] for a set of
// rules written in place, [Nested] for a value that validates itself, [Each]
// or [EachNested] for the elements of a slice, and [When], [Unless] and
// [WhenFunc] for rules that only apply sometimes.
//
// A Rule carries no type parameter, because those constructors have already
// closed over the values, which is what lets one [Validate] call cover fields
// of different types. It is also what keeps this package free of reflection:
// the field name and value arrive at the call site, so nothing has to be
// recovered from struct tags at runtime, and a check that does not fit the
// field's type fails to compile.
//
// A Rule reports paths relative to itself. Whoever runs it prepends their own
// segment, so the same Rule composes at any depth.
type Rule func() Errors

// Field binds a name and a value to the checks that value has to pass.
//
// name is the path segment for this field, so it is yours to choose: the Go
// field name, the JSON name, whatever the caller of your API will recognise.
// Nothing is derived from the struct, which is the trade this package makes. You
// repeat the name once, and in exchange there are no tags to keep in sync and no
// reflection at runtime.
//
// All the checks run, including those after the first failure, so a field with
// several things wrong reports all of them.
//
// Every failure a check reports lands at or below name. A plain error lands on
// the field itself; an [Error] or [Errors] keeps whatever relative path it
// carries, beneath name. Nothing a check returns can escape upwards.
//
// An empty name adds no segment, so the checks report against whatever encloses
// the field. Field with no checks contributes nothing and is not an error.
func Field[V any](name string, val V, checks ...Check[V]) Rule {
	return func() Errors {
		var errs Errors
		for _, ch := range checks {
			err := ch(val)
			if err == nil {
				continue
			}

			errs = append(errs, toErrors(name, err)...)
		}

		return errs
	}
}

// Group nests rules one level down, under the path segment name.
//
// It is [Validate]'s recursive half: Validate is a Group at the root that turns
// an empty result into a nil error. Use it for a struct field validated in
// place, so the nesting is visible where you read it:
//
//	validate.Validate("user",
//		validate.Field("name", u.Name, validate.Required[string]()),
//		validate.Group("address",
//			validate.Field("street", u.Address.Street, validate.Required[string]()),
//			validate.Field("zip", u.Address.Zip, validate.Required[string]()),
//		),
//	)
//
//	// user.address.street: is required
//	// user.address.zip: is required
//
// When the nested type validates itself, reach for [Nested] instead and skip
// writing the rules out again. Group is for rules written in place, and for
// naming a [Rule] that a type hands back:
//
//	func (a Address) Rules() validate.Rule {
//		return validate.Group("",
//			validate.Field("street", a.Street, validate.Required[string]()),
//		)
//	}
//
//	validate.Validate("user", validate.Group("address", a.Rules())) // user.address.street
//	validate.Validate("address", a.Rules())                         // address.street
//
// An empty name adds no segment, so Group("", rules...) flattens several rules
// into one without moving them. Group with no rules contributes nothing.
func Group(name string, rules ...Rule) Rule {
	return func() Errors {
		var errs Errors
		for _, r := range rules {
			errs = append(errs, r().under(name)...)
		}

		return errs
	}
}

// Nested validates a value that already validates itself, under the path segment
// name.
//
// It is [Field] with one check that calls the value's Validate method, so
// everything Field does applies. Whatever Validate returns is placed at or below
// name, whether that is an [Errors] from another [Validate] call, a single
// [Error], or a plain error from somewhere else entirely.
//
//	func (u User) Validate() error {
//		return validate.Validate("user",
//			validate.Field("name", u.Name, validate.Required[string]()),
//			validate.Nested("address", u.Address),
//		)
//	}
//
// Paths compose as they do everywhere else, one segment per level, so which name
// to pass depends on whether the nested type roots itself:
//
//	// Address.Validate calls validate.Validate("address", ...)
//	validate.Nested("", u.Address)     // user.address.street
//	validate.Nested("home", u.Address) // user.home.address.street
//
//	// Address.Validate calls validate.Validate("", ...)
//	validate.Nested("address", u.Address) // user.address.street
//	validate.Nested("home", u.Home)       // user.home.street
//
// A type that names itself reads well standalone and nests under Nested(""). One
// that does not is named by whoever nests it, which is what you want when two
// fields share a type. For a slice of them, see [EachNested].
//
// A nil pointer panics here exactly as calling the method directly would. Finding
// the nil inside an interface needs reflection, which this package does not do,
// so give a pointer receiver the guard it wants anyway:
//
//	func (a *Address) Validate() error {
//		if a == nil {
//			return nil
//		}
//
//		return validate.Validate("", ...)
//	}
func Nested(name string, v Validator) Rule {
	return Field(name, v, Validator.Validate)
}

// Each applies rule to every element of vals, under an indexed path segment.
//
// The segment is name[i], so a failure in the third tag reads "user.tags[2]".
// Elements are visited in order, and all of them are visited: like everything
// else here, Each does not stop at the first failure.
//
// rule builds the [Rule] for one element. Give the inner [Field] an empty name
// to report on the element itself, since the index has already named it:
//
//	validate.Each("tags", u.Tags, func(t string) validate.Rule {
//		return validate.Field("", t, validate.Required[string]())
//	})
//
//	// user.tags[2]: is required
//
// Use it for elements that need checks of their own, or several rules at once.
// When every element simply validates itself, [EachNested] says the same thing
// without the callback.
//
// rule runs once per element, and not at all until the returned Rule is run. A
// nil or empty slice contributes nothing.
//
// An empty name still emits the index, since dropping it would collapse every
// element onto one path, so the segment is "[0]" rather than nothing.
func Each[T any](name string, vals []T, rule func(T) Rule) Rule {
	return func() Errors {
		var errs Errors
		for i, v := range vals {
			errs = append(errs, Group(indexed(name, i), rule(v))()...)
		}

		return errs
	}
}

// EachNested validates every element of a slice of self-validating values, under
// an indexed path segment.
//
// It is [Each] with [Nested] as the rule, which is what a slice of sub-objects
// almost always wants:
//
//	validate.EachNested("contacts", u.Contacts)
//
//	// user.contacts[1].street: is required
//	// user.contacts[1].zip: is required
//
// Each element keeps whatever name it gives itself, since the index has already
// said which one failed. So a type whose Validate roots itself contributes a
// segment here exactly as it does under [Nested]:
//
//	// Validate("", ...)        ->  user.contacts[1].street
//	// Validate("address", ...) ->  user.contacts[1].address.street
//
// vals is a []T rather than a []Validator so a slice of concrete types can be
// passed straight in, since Go will not convert []Address to []Validator. T still
// has to satisfy [Validator], so a Validate method on a pointer receiver means a
// slice of pointers.
//
// A nil or empty slice contributes nothing. Reach for [Each] when the elements
// need anything beyond their own Validate method.
func EachNested[T Validator](name string, vals []T) Rule {
	return Each(name, vals, func(v T) Rule { return Nested("", v) })
}

// indexed builds the path segment for the i'th element of a slice named name.
func indexed(name string, i int) string {
	return name + "[" + strconv.Itoa(i) + "]"
}
