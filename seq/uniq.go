package seq

import "iter"

// Uniq returns a sequence with duplicates removed, keeping the first occurrence
// of each value and preserving the order of seq.
//
// Uniq tracks what it has seen in a set built fresh for each iteration, so the
// returned sequence can be ranged over more than once, provided seq can be too.
// That set holds every distinct value seen so far, so memory grows with the
// number of distinct values rather than the length of seq.
func Uniq[T comparable](seq iter.Seq[T]) iter.Seq[T] {
	return UniqBy(seq, func(t T) T { return t })
}

// UniqBy is [Uniq] for values that are not usefully comparable on their own. key
// maps each value to something that is, so a sequence of structs can be deduped
// by name, ID, or whatever else tells them apart. The first value seen for a
// given key wins and later ones are dropped, preserving the order of seq.
//
// Only the key has to be comparable, not T itself, so this works on structs
// holding slices, maps, or functions. key runs once per value in seq, and the
// set of seen keys is built fresh for each iteration, so the returned sequence
// can be ranged over more than once, provided seq can be too.
func UniqBy[T any, U comparable](seq iter.Seq[T], key func(T) U) iter.Seq[T] {
	return func(yield func(T) bool) {
		seen := make(map[U]struct{})
		for v := range seq {
			k := key(v)
			if _, ok := seen[k]; ok {
				continue
			}

			seen[k] = struct{}{}
			if !yield(v) {
				return
			}
		}
	}
}
