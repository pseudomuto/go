package maps

// Filter returns a new map holding the entries of m for which pred returns true.
//
// pred takes an [Entry] rather than a separate key and value, so it names the two
// sides instead of ordering them. It runs once per entry, including those left
// out of the result.
//
// Unlike most of this package, Filter needs the keys to be comparable only, not
// ordered. A map has no order to be nondeterministic about, so there is nothing
// to stabilise here.
//
// m is left untouched. The result is always non-nil and safe to write to, even
// when nothing matched, since a nil map would panic on the first write.
func Filter[T comparable, U any](m map[T]U, pred func(Entry[T, U]) bool) map[T]U {
	// Ranged directly rather than routed through a seq adapter. No seq primitive
	// produces a map, and going via Entries would require cmp.Ordered keys that
	// this function does not otherwise need.
	out := make(map[T]U)

	for k, v := range m {
		if pred(Entry[T, U]{Key: k, Value: v}) {
			out[k] = v
		}
	}

	return out
}

// Reject returns a new map holding the entries of m for which pred returns
// false.
//
// It is the complement of [Filter] under the same pred: for a pred that answers
// consistently, the two together cover every entry of m exactly once. The same
// notes about [Entry], key constraints, and the non-nil result apply.
func Reject[T comparable, U any](m map[T]U, pred func(Entry[T, U]) bool) map[T]U {
	return Filter(m, func(e Entry[T, U]) bool { return !pred(e) })
}
