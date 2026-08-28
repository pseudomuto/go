package maps

import (
	"cmp"
	"slices"
)

// Entry is one key and value pair from a map.
//
// Functions that hand whole entries to a callback pass this rather than a
// separate key and value, so the callback names the two sides instead of
// ordering them. On a map[string]string a positional callback compiles with the
// two the wrong way round, and the mistake only shows up in the output.
type Entry[K comparable, V any] struct {
	Key   K
	Value V
}

// Entries returns m's entries as a slice, in ascending key order.
//
// The ordering is part of the contract rather than a convenience. A slice
// implies order, so handing back the runtime's randomized order would make
// Entries(m)[0] a different entry on every run. That is why the keys have to be
// ordered rather than merely comparable.
//
// An empty or nil map returns nil. [FromEntries] is the inverse, so
// FromEntries(Entries(m)) reproduces m.
func Entries[T cmp.Ordered, U any](m map[T]U) []Entry[T, U] {
	return slices.Collect(orderedEntries(m))
}

// FromEntries builds a map from es. Entries are applied in order, so when two
// share a key the later one wins.
//
// The result is always non-nil and safe to write to, even for an empty or nil
// slice. That follows the standard library's maps.Collect rather than this
// package's nil-for-nothing rule, which covers slice returns only: a nil map
// would panic on the first write.
//
// es is left untouched. [Entries] is the inverse, so FromEntries(Entries(m))
// reproduces m.
func FromEntries[T comparable, U any](es []Entry[T, U]) map[T]U {
	m := make(map[T]U, len(es))
	for _, e := range es {
		m[e.Key] = e.Value
	}

	return m
}
