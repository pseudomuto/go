package slices

import (
	"slices"

	"github.com/pseudomuto/go/seq"
)

// Filter returns a new slice holding the elements of sl for which pred returns
// true, in their original order.
//
// pred runs for every element, including those left out of the result. sl is
// left untouched, and a slice with nothing to keep returns nil.
func Filter[T any](sl []T, pred func(T) bool) []T {
	return slices.Collect(seq.Filter(slices.Values(sl), pred))
}

// Reject returns a new slice holding the elements of sl for which pred returns
// false, in their original order.
//
// It is the complement of [Filter] under the same pred: for a pred that answers
// consistently, the two together cover every element of sl exactly once. sl is
// left untouched, and a slice with nothing to keep returns nil.
func Reject[T any](sl []T, pred func(T) bool) []T {
	return slices.Collect(seq.Reject(slices.Values(sl), pred))
}
