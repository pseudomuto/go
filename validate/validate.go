package validate

type (
	// Check tests one value and describes what is wrong with it.
	//
	// A Check returns nil when the value is acceptable, and otherwise an error whose
	// message completes the sentence "<field> ...", so messages read as "is
	// required" or "not greater than 0" rather than repeating the field name. The
	// enclosing [Field] supplies the name.
	//
	// A Check is an ordinary function, so anything with the right signature works:
	//
	//	notReserved := func(name string) error {
	//		if reserved[name] {
	//			return fmt.Errorf("%q is reserved", name)
	//		}
	//
	//		return nil
	//	}
	//
	// The constructors in this package ([Required], [GT], and the rest) return one,
	// as do those in the is subpackage. They are constructors rather than checks
	// themselves so they can close over a bound, and so their type parameter is
	// inferred from that bound.
	//
	// A Check that wants to place its failure more precisely than "this field" can
	// return an [Error] or an [Errors] instead of a plain error; see [Fail].
	Check[T any] func(T) error

	// Validator is anything that validates itself, which is the shape a type's own
	// Validate method already has. [Nested] takes one.
	//
	// Nothing has to be registered or declared for a type to qualify. If it has the
	// method, it nests.
	Validator interface {
		Validate() error
	}
)

// Validate runs every rule under ns and reports what failed, or nil when nothing
// did.
//
// Nothing short-circuits. Every rule runs, and every check within a rule runs,
// so one call reports all the problems rather than the first. Results come back
// in the order the rules were given, and within a rule in the order the checks
// were given.
//
// ns is the first segment of every path this call produces, so a failure reads
// "user.email: is required". Paths are built by prepending, one segment per
// level, from the outside in:
//
//	validate.Validate("user",                              // user
//		validate.Group("address",                          // user.address
//			validate.Field("street", a.Street, required),  // user.address.street
//		),
//	)
//
// An empty ns contributes no segment, which is what makes a validator reusable
// at depth: see [Nested] for how a type's own Validate method nests.
//
// The returned error is an [Errors] holding one [Error] per failure. It supports
// errors.Is and errors.As down to the individual failures, so a caller can pull
// out the paths rather than parse the string:
//
//	var verrs validate.Errors
//	if errors.As(err, &verrs) {
//		for _, e := range verrs {
//			log.Println(e.Path(), e.Error())
//		}
//	}
//
// Validate with no rules, or with rules that all pass, returns a nil error
// rather than an empty [Errors], so the usual `if err != nil` is safe.
func Validate(ns string, rules ...Rule) error {
	errs := Group(ns, rules...)()
	if len(errs) == 0 {
		return nil
	}

	return errs
}
