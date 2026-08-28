package maps

import (
	"cmp"
	"iter"

	"github.com/pseudomuto/go/seq"
)

// UniqKeysBy returns a sequence of m's keys with duplicates removed, where two
// keys count as duplicates when key returns the same value for both.
//
// Map keys are already distinct, so this only drops anything when key collapses
// two of them, folding case for instance. To keep that outcome reproducible,
// UniqKeysBy walks m in ascending key order and keeps the first key it sees for
// each derived value. Without the sort the survivor would depend on Go's
// randomized map iteration order and could differ between identical calls.
//
// The sort runs on each iteration, so the returned sequence can be ranged over
// more than once and always reflects m as it stands at that moment.
func UniqKeysBy[T cmp.Ordered, U any, V comparable](m map[T]U, key func(T) V) iter.Seq[T] {
	return seq.UniqBy(orderedKeys(m), key)
}

// UniqValuesBy returns a sequence of m's values with duplicates removed, where
// two values count as duplicates when key returns the same result for both.
//
// UniqValuesBy walks m in ascending key order and keeps the first value it finds
// for each derived key. The order matters: without it the surviving
// representative would depend on Go's randomized map iteration order and could
// differ between identical calls on the same map.
//
// The sort runs on each iteration, so the returned sequence can be ranged over
// more than once and always reflects m as it stands at that moment.
func UniqValuesBy[T cmp.Ordered, U any, V comparable](m map[T]U, key func(U) V) iter.Seq[U] {
	return seq.UniqBy(orderedValues(m), key)
}
