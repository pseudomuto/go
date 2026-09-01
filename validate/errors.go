package validate

import (
	"errors"
	"fmt"
	"strings"
)

type (
	// Error is a single validation failure: one message, at one path.
	//
	// The path names the failure's position in whatever was validated, as
	// dot-separated segments running from the outside in, so a street inside an
	// address inside a user reads "user.address.street". It is built up rather than
	// set: a [Check] reports a message and, if it wants, a path relative to itself,
	// then every enclosing [Field], [Group] and [Validate] prepends its own segment
	// on the way out. See [Fail] for what a check can say about placement.
	//
	// Error is comparable, so errors.Is matches on path and message together.
	Error struct {
		path string
		msg  string
	}

	// Errors is a list of failures, and is itself an error.
	//
	// This is what [Validate] returns when anything failed. It never comes back
	// empty from [Validate], which returns nil instead.
	//
	// Errors unwraps to its elements, so errors.As reaches both the whole list and
	// any single [Error] in it:
	//
	//	var verrs validate.Errors // the lot
	//	var verr validate.Error   // the first one
	Errors []Error
)

// Fail builds a failure with message msg, at a path relative to wherever the
// [Check] returning it is running.
//
// Placement is relative, not absolute. Whatever path is given here sits below
// the enclosing [Field], which sits below its [Group] and [Validate], so a check
// can name something inside the value it was handed but cannot report against a
// field elsewhere in the struct. For a cross-field rule, hang the [Field] on the
// field the message is about and close over the rest.
//
// With no path, the failure lands on the enclosing field itself, which is what
// returning a plain error already does. Fail earns its keep when a check looks
// inside its value:
//
//	// Inside Field("address", ...) under Validate("user"):
//	return validate.Fail("is malformed", "zip") // user.address.zip: is malformed
//	return validate.Fail("is malformed")        // user.address: is malformed
//
// Empty segments are dropped rather than leaving an empty step in the path.
func Fail(msg string, path ...string) Error {
	return Error{path: joinPath(path...), msg: msg}
}

// Path returns the full dot-separated path to the failure, outermost segment
// first, as it appears in the message.
//
// It is empty only when nothing along the way named anything.
func (e Error) Path() string {
	return e.path
}

// Field returns the last segment of the path, which is the thing the message is
// actually about.
func (e Error) Field() string {
	if i := strings.LastIndex(e.path, "."); i >= 0 {
		return e.path[i+1:]
	}

	return e.path
}

// Namespace returns everything in the path above [Error.Field], so Path is
// Namespace and Field rejoined. It is empty for a failure at the top level.
func (e Error) Namespace() string {
	if i := strings.LastIndex(e.path, "."); i >= 0 {
		return e.path[:i]
	}

	return ""
}

// Error formats the failure as "path: message", or as the message alone when the
// path is empty.
func (e Error) Error() string {
	if e.path == "" {
		return e.msg
	}

	return fmt.Sprintf("%s: %s", e.path, e.msg)
}

// under returns e moved one level down, beneath segment seg. An empty seg leaves
// the path alone, which is what lets [Validate] and [Group] go unnamed.
func (e Error) under(seg string) Error {
	e.path = joinPath(seg, e.path)
	return e
}

// Error formats every failure, one per line, in the order they were collected.
//
// An empty Errors formats as the empty string. That case does not arise from
// [Validate], but Errors is an ordinary slice type that anyone can construct, and
// errors.Join of nothing is nil.
func (ve Errors) Error() string {
	if len(ve) == 0 {
		return ""
	}

	errs := make([]error, len(ve))
	for i, e := range ve {
		errs[i] = e
	}

	return errors.Join(errs...).Error()
}

// Unwrap exposes the failures to errors.Is and errors.As, which is what lets a
// caller ask about one field without walking the list or parsing the message.
//
// The returned slice is fresh, so writing to it does not touch ve.
func (ve Errors) Unwrap() []error {
	errs := make([]error, len(ve))
	for i, e := range ve {
		errs[i] = e
	}

	return errs
}

// under returns ve with every failure moved one level down, beneath segment seg,
// leaving ve itself untouched.
func (ve Errors) under(seg string) Errors {
	out := make(Errors, len(ve))
	for i, e := range ve {
		out[i] = e.under(seg)
	}

	return out
}

// toErrors normalises whatever a [Check] returned into [Errors] positioned
// beneath segment seg, which is the field name the check was registered under.
func toErrors(seg string, err error) Errors {
	if ves, ok := errors.AsType[Errors](err); ok {
		return ves.under(seg)
	}

	if ve, ok := errors.AsType[Error](err); ok {
		return Errors{ve.under(seg)}
	}

	return Errors{Error{msg: err.Error()}.under(seg)}
}

// joinPath builds a path from segs, dropping empty ones so an unnamed level
// contributes nothing rather than an empty step.
func joinPath(segs ...string) string {
	kept := make([]string, 0, len(segs))
	for _, s := range segs {
		if s != "" {
			kept = append(kept, s)
		}
	}

	return strings.Join(kept, ".")
}
