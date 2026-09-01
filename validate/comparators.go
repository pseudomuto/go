package validate

import (
	"cmp"
	"errors"
	"fmt"
)

// GT reports values that are not strictly greater than mark.
//
// The boundary itself fails: GT(0) rejects 0 and accepts anything above it. Use
// [GTE] when the boundary should pass.
func GT[T cmp.Ordered](mark T) Check[T] {
	return ordered(mark, "not greater than %v", func(v T) bool { return v > mark })
}

// GTE reports values that are not greater than or equal to mark.
//
// The boundary passes: GTE(0) accepts 0. Use [GT] when the boundary should fail.
func GTE[T cmp.Ordered](mark T) Check[T] {
	return ordered(mark, "not greater than or equal to %v", func(v T) bool { return v >= mark })
}

// LT reports values that are not strictly less than mark.
//
// The boundary itself fails: LT(10) rejects 10 and accepts anything below it.
// Use [LTE] when the boundary should pass.
func LT[T cmp.Ordered](mark T) Check[T] {
	return ordered(mark, "not less than %v", func(v T) bool { return v < mark })
}

// LTE reports values that are not less than or equal to mark.
//
// The boundary passes: LTE(10) accepts 10. Use [LT] when the boundary should
// fail.
func LTE[T cmp.Ordered](mark T) Check[T] {
	return ordered(mark, "not less than or equal to %v", func(v T) bool { return v <= mark })
}

// Required reports the zero value of T.
//
// What counts as absent is whatever Go calls the zero value, so this rejects "",
// 0, false, and the nil pointer, and it cannot tell an explicit zero from an
// unset field. On a field where 0 or false is a legitimate value, reach for a
// pointer and check that instead, or use [GTE] and friends to state the real
// bound.
//
// T must be comparable, which rules out slices, maps, and functions. For a slice
// field that must be non-empty, compare its length: GT(0) over len(v), or a
// hand-written [Check].
func Required[T comparable]() Check[T] {
	return func(v T) error {
		var zero T
		if v == zero {
			return errors.New("is required")
		}

		return nil
	}
}

// Unique reports a slice holding the same value twice.
//
// It stops at the first duplicate rather than collecting them all, so the
// message names one offending value even when there are several. Order is
// irrelevant: duplicates need not be adjacent. A nil or single-element slice
// always passes.
//
// Comparison is Go's ==, so this works on the same types [Required] does.
func Unique[T comparable]() Check[[]T] {
	return func(vs []T) error {
		seen := make(map[T]struct{}, len(vs))
		for _, v := range vs {
			if _, dup := seen[v]; dup {
				return fmt.Errorf("contains duplicate value: %v", v)
			}

			seen[v] = struct{}{}
		}

		return nil
	}
}

// ordered builds a [Check] that passes when pass reports true, and otherwise
// fails with msg, which must hold a single verb for mark.
func ordered[T cmp.Ordered](mark T, msg string, pass func(T) bool) Check[T] {
	return func(v T) error {
		if pass(v) {
			return nil
		}

		return fmt.Errorf(msg, mark)
	}
}
