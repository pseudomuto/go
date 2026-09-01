// Package validate provides declarative struct validation with no reflection.
//
// Validation is spelled out as ordinary Go: you name the field, hand over its
// value, and list the checks it has to pass. There are no struct tags to parse,
// no field lookups by name, and nothing that can only fail at runtime. A check
// that does not fit the field's type is a compile error, and renaming a field
// does not silently disconnect its rules.
//
//	func (a Address) Validate() error {
//		return validate.Validate("",
//			validate.Field("cidr", a.CIDR, validate.IsCIDRv4()),
//		)
//	}
//
//	func (u User) Validate() error {
//		return validate.Validate("user",
//			validate.Field("name", u.Name, validate.Required[string]()),
//			validate.Field("age", u.Age, validate.GTE(0), validate.LT(150)),
//			validate.Nested("address", u.Address),
//			validate.Each("tags", u.Tags, func(t string) validate.Rule {
//				return validate.Field("", t, validate.Required[string]())
//			}),
//		)
//	}
//
// A User with an empty name, an age of 200, a bad subnet and a blank second tag
// fails four times, and the returned error formats as:
//
//	user.name: is required
//	user.age: not less than 150
//	user.address.cidr: is not a valid CIDR(v4) range
//	user.tags[1]: is required
//
// # The pieces
//
// A [Check] is a function from one value to an error, nil when the value is
// fine. A handful of constructors turn checks and values into a [Rule]: [Field]
// binds a name and a value to the checks for it, [Group] nests a set of rules
// under a name, [Nested] hands a value to its own Validate method, and [Each]
// applies a rule to every element of a slice, with [EachNested] covering the
// common case of a slice that validates itself. [When], [Unless] and [WhenFunc]
// guard rules that only apply sometimes. [Validate] runs rules and returns nil,
// or an [Errors] holding one [Error] per failure.
//
// A Rule carries no type parameter, because the constructor has already closed
// over the value. That is what lets one [Validate] call cover fields of
// different types, and it is the whole reason this works without reflection.
//
// [Nested] is what lets an existing Validate method be reused rather than
// restated. Any type with one satisfies [Validator]; nothing has to be
// registered.
//
// # Paths
//
// Every failure carries a path: dot-separated segments naming its position, from
// the outside in. Paths are built by prepending, one segment per level, as the
// failure travels back out:
//
//	a check reports              is required
//	Field("street", ...)         street: is required
//	Group("address", ...)        address.street: is required
//	Validate("user", ...)        user.address.street: is required
//
// Nothing about that is special-cased, which is what makes it predictable.
// Segments accumulate, they never merge or replace one another, so two levels
// that pick the same name both appear. An empty name contributes no segment, so
// a [Rule] written without an opinion about where it sits composes at any depth,
// and [Validate] itself is just a [Group] at the root.
//
// [Each] and [EachNested] contribute an indexed segment, name[i], so an
// element's own failures sit under it: "user.contacts[2].street". A segment is
// just a string, so nothing special is needed to read or format one.
//
// A [Check] reports relative to the [Field] it was registered under, and cannot
// escape upwards. [Fail] is how it names something below itself, such as a
// key inside the map it was handed.
//
// # Everything runs
//
// Nothing short-circuits. Every rule runs, and every check within a rule runs,
// so one call reports every problem rather than the first. Results keep the
// order they were declared in, so the output is the same on every run.
//
// [When] and [Unless] are not an exception to that. A condition that does not
// hold drops the rules it guards and nothing else, so everything around it still
// runs and still reports.
//
// # Failures are structured
//
// [Errors] is a list, not a string. It unwraps to its elements, so errors.As
// reaches either the whole list or a single [Error], and each one exposes its
// [Error.Path] separately from its message. Callers that need to attach failures
// to form inputs can do that without parsing anything. [Error.Field] and
// [Error.Namespace] split the path at the last segment for the common case of
// keying by field name.
//
// # Writing checks
//
// The constructors here ([Required], [GT], [GTE], [LT], [LTE], [Unique], and the
// net checks [IsIP], [IsIPv4], [IsIPv6], [IsCIDR], [IsCIDRv4], [IsCIDRv6]) cover
// the common cases, but they are not privileged. A [Check] is any func(T) error,
// so a one-off rule is just a function literal, and a rule you reuse is a
// function that returns one. There is no registry to add it to.
//
// Message convention: a check's message completes the sentence "<field> ...",
// since [Field] supplies the name. "is required" and "not greater than 0", not
// "name is required".
package validate
