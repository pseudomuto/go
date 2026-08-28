package slices

import (
	"slices"

	"github.com/pseudomuto/go/seq"
)

// Reduce folds sl into a single value. It starts with init and calls fn once per
// element, left to right, passing the running accumulator and the element. The
// final accumulator is the result, so an empty or nil slice yields init
// unchanged.
//
// sl is left untouched. The accumulator type is free to differ from the element
// type, and Go infers it from init, so no explicit type arguments are needed.
func Reduce[T, U any](sl []T, init U, fn func(U, T) U) U {
	return seq.Reduce(slices.Values(sl), init, fn)
}
