package slices

import (
	"slices"

	"github.com/pseudomuto/go/seq"
)

// Uniq returns a new slice with duplicates removed, keeping the first
// occurrence of each value and preserving the order of sl.
//
// sl is left untouched, and the result never shares a backing array with it. An
// empty or nil slice returns nil.
func Uniq[T comparable](sl []T) []T {
	return slices.Collect(seq.Uniq(slices.Values(sl)))
}

// UniqBy is [Uniq] for values that are not usefully comparable on their own. key
// maps each value to something that is, so a slice of structs can be deduped by
// name, ID, or whatever else tells them apart. The first value seen for a given
// key wins and later ones are dropped, preserving the order of sl.
//
// Only the key has to be comparable, not T, so this works on structs holding
// slices, maps, or functions. key runs once per value in sl, duplicates
// included.
//
// sl is left untouched, and the result never shares a backing array with it. An
// empty or nil slice returns nil.
func UniqBy[T any, U comparable](sl []T, key func(T) U) []T {
	return slices.Collect(seq.UniqBy(slices.Values(sl), key))
}
