package maps

import (
	"cmp"
	"iter"
	"maps"
	"slices"
)

// SortedKeys returns m's keys in ascending order. An empty or nil map returns
// nil.
func SortedKeys[T cmp.Ordered, U any](m map[T]U) []T {
	return slices.Sorted(maps.Keys(m))
}

// SortedKeysFunc returns m's keys ordered by compare, which must return a
// negative number when a sorts before b, zero when they tie, and a positive
// number when a sorts after b.
//
// Because the caller supplies the ordering, the keys only have to be
// comparable, not ordered. An empty or nil map returns nil.
func SortedKeysFunc[T comparable, U any](m map[T]U, compare func(T, T) int) []T {
	return slices.SortedFunc(maps.Keys(m), compare)
}

// SortedValues returns m's values in ascending order. Values that repeat are all
// kept, since this sorts rather than dedupes. An empty or nil map returns nil.
func SortedValues[T comparable, U cmp.Ordered](m map[T]U) []U {
	return slices.Sorted(maps.Values(m))
}

// SortedValuesFunc returns m's values ordered by compare, which must return a
// negative number when a sorts before b, zero when they tie, and a positive
// number when a sorts after b.
//
// Because the caller supplies the ordering, the values themselves do not have to
// be ordered, which is what makes this usable on structs. An empty or nil map
// returns nil.
func SortedValuesFunc[T comparable, U any](m map[T]U, compare func(U, U) int) []U {
	return slices.SortedFunc(maps.Values(m), compare)
}

// orderedEntries returns a sequence of m's entries in ascending key order. It
// sorts on each iteration rather than once at call time, so the sequence stays
// re-iterable and never serves entries from a stale snapshot.
//
// orderedKeys and orderedValues both project off this, which keeps the sort in
// one place.
func orderedEntries[T cmp.Ordered, U any](m map[T]U) iter.Seq[Entry[T, U]] {
	return func(yield func(Entry[T, U]) bool) {
		for _, k := range slices.Sorted(maps.Keys(m)) {
			if !yield(Entry[T, U]{Key: k, Value: m[k]}) {
				return
			}
		}
	}
}

// orderedKeys returns a sequence of m's keys in ascending order.
func orderedKeys[T cmp.Ordered, U any](m map[T]U) iter.Seq[T] {
	return func(yield func(T) bool) {
		for e := range orderedEntries(m) {
			if !yield(e.Key) {
				return
			}
		}
	}
}

// orderedValues returns a sequence of m's values ordered by their keys.
func orderedValues[T cmp.Ordered, U any](m map[T]U) iter.Seq[U] {
	return func(yield func(U) bool) {
		for e := range orderedEntries(m) {
			if !yield(e.Value) {
				return
			}
		}
	}
}
