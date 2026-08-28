package slices

import (
	"slices"

	"github.com/pseudomuto/go/seq"
)

// Map returns a new slice holding the result of fn applied to each element of
// sl, in order. sl is left untouched, and an empty or nil slice returns nil.
//
// The result is allocated once with exactly enough capacity, since its length is
// known up front. fn runs once per element.
func Map[T, U any](sl []T, fn func(T) U) []U {
	if len(sl) == 0 {
		return nil
	}

	return slices.AppendSeq(make([]U, 0, len(sl)), seq.Map(slices.Values(sl), fn))
}

// MapErr is [Map] for an fn that can fail. It returns the mapped slice, or the
// first error fn produced.
//
// Like [CollectErr], which it is built on, MapErr is all or nothing: on error it
// returns a nil slice and that error, discarding whatever mapped cleanly before
// it. fn stops being called at the first failure, so later elements are never
// visited.
//
// sl is left untouched. An empty or nil slice returns nil and no error.
func MapErr[T, U any](sl []T, fn func(T) (U, error)) ([]U, error) {
	return CollectErr(seq.MapErr(slices.Values(sl), fn))
}
