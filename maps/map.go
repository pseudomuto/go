package maps

import (
	"iter"
	"maps"

	"github.com/pseudomuto/go/seq"
)

// Map returns a sequence of m's entries with fn applied to each.
//
// fn takes an [Entry] rather than a separate key and value, so it names the two
// sides instead of ordering them. See [MapKeys] and [MapValues] for mapping one
// side on its own.
//
// The results arrive in map iteration order, which Go randomizes, so treat them
// as a bag rather than a sequence. Sort the result, or build from [Entries],
// when the order matters.
func Map[T comparable, U, V any](m map[T]U, fn func(Entry[T, U]) V) iter.Seq[V] {
	// Ranged directly rather than routed through seq.Map over maps.Keys. Going
	// via keys costs a second hash lookup per entry.
	return func(yield func(V) bool) {
		for k, v := range m {
			if !yield(fn(Entry[T, U]{Key: k, Value: v})) {
				return
			}
		}
	}
}

// MapErr is [Map] for an fn that can fail. It returns a sequence of value and
// error pairs, one per entry in m.
//
// When fn fails, the pair holds the zero value of V and fn's error. MapErr never
// stops on an error, it keeps going, so the consumer chooses whether to break on
// the first failure or skip it and carry on. To collect into a slice instead,
// stopping at the first error, hand the result to slices.CollectErr.
//
// The results arrive in map iteration order, which Go randomizes, so treat them
// as a bag rather than a sequence.
func MapErr[T comparable, U, V any](m map[T]U, fn func(Entry[T, U]) (V, error)) iter.Seq2[V, error] {
	// Ranged directly rather than routed through seq.Map over maps.Keys. Going
	// via keys costs a second hash lookup per entry.
	return func(yield func(V, error) bool) {
		for k, v := range m {
			if !yield(fn(Entry[T, U]{Key: k, Value: v})) {
				return
			}
		}
	}
}

// MapKeys returns a sequence of m's keys with fn applied to each.
//
// The keys arrive in map iteration order, which Go randomizes, so treat the
// result as a set rather than a sequence. Reach for [SortedKeys] or
// [SortedKeysFunc] when the order matters.
func MapKeys[T comparable, U, V any](m map[T]U, fn func(T) V) iter.Seq[V] {
	return seq.Map(maps.Keys(m), fn)
}

// MapValues returns a sequence of m's values with fn applied to each.
//
// The values arrive in map iteration order, which Go randomizes, so treat the
// result as a bag rather than a sequence. Reach for [SortedValues] or
// [SortedValuesFunc] when the order matters.
func MapValues[T comparable, U, V any](m map[T]U, fn func(U) V) iter.Seq[V] {
	return seq.Map(maps.Values(m), fn)
}
