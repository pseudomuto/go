package chain

import (
	"cmp"
	"iter"

	"github.com/pseudomuto/go/maps"
)

// Map is a map[K]V with the maps package's helpers hung on it as methods, so a
// composition reads left to right instead of inside out.
//
// It is a defined type rather than a struct, so it *is* a map: index it, take its
// len, range over it, write a composite literal, and pass it to anything
// expecting a map[K]V with no conversion. That is why there is no Unwrap or Len
// method here.
//
//	m := chain.Map[string, int]{"a": 1}
//	fmt.Println(len(m), m["a"])
//	var plain map[string]int = m
//
// K is ordered rather than merely comparable because every order-dependent
// function in [github.com/pseudomuto/go/maps] requires it, and a method cannot
// add a constraint its receiver lacks. Without it this type would lose
// [Map.Reduce], [Map.Entries] and [Map.SortedKeys]. That rules out struct, array
// and pointer keys; those callers use the maps functions directly.
//
// V is unconstrained, so map[string][]string works.
type Map[K cmp.Ordered, V any] map[K]V

// OfMap converts m for chaining. It is the same map, not a copy. Every method
// that transforms returns a freshly built map, leaving m alone.
func OfMap[K cmp.Ordered, V any](m map[K]V) Map[K, V] {
	return Map[K, V](m)
}

// Filter keeps the entries for which pred returns true. See [maps.Filter].
func (m Map[K, V]) Filter(pred func(maps.Entry[K, V]) bool) Map[K, V] {
	return maps.Filter(m, pred)
}

// Reject drops the entries for which pred returns true. See [maps.Reject].
func (m Map[K, V]) Reject(pred func(maps.Entry[K, V]) bool) Map[K, V] {
	return maps.Reject(m, pred)
}

// Map applies fn to every entry and keeps chaining as a [Seq] of the results. See
// [maps.Map].
//
// Results arrive in map iteration order, which Go randomizes. That is fine for a
// sequence, but [Seq.Collect] then materialises that order into a slice. Build
// from [Map.SortedKeys] or [Map.Entries] when order matters.
func (m Map[K, V]) Map[R any](fn func(maps.Entry[K, V]) R) Seq[R] {
	return Seq[R](maps.Map(m, fn))
}

// MapErr is [Map.Map] for an fn that can fail, yielding value and error pairs so
// the consumer decides whether to stop. See [maps.MapErr].
//
// This ends the chain at a bare [iter.Seq2], since there is no wrapper for pairs.
func (m Map[K, V]) MapErr[R any](fn func(maps.Entry[K, V]) (R, error)) iter.Seq2[R, error] {
	return maps.MapErr(m, fn)
}

// MapKeys applies fn to every key and keeps chaining as a [Seq] of the results.
// See [maps.MapKeys].
//
// It yields a sequence of transformed keys, not a map re-keyed by fn, because no
// maps function produces the latter. Results arrive in map iteration order.
func (m Map[K, V]) MapKeys[R any](fn func(K) R) Seq[R] {
	return Seq[R](maps.MapKeys(m, fn))
}

// MapValues applies fn to every value and keeps chaining as a [Seq] of the
// results. See [maps.MapValues]. Results arrive in map iteration order.
func (m Map[K, V]) MapValues[R any](fn func(V) R) Seq[R] {
	return Seq[R](maps.MapValues(m, fn))
}

// UniqKeysBy yields the keys with duplicates removed, comparing the key derived by
// key. See [maps.UniqKeysBy]. Walks the map in ascending key order, so which key
// survives a collision is the same on every run.
func (m Map[K, V]) UniqKeysBy[R comparable](key func(K) R) Seq[K] {
	return Seq[K](maps.UniqKeysBy(m, key))
}

// UniqValuesBy yields the values with duplicates removed, comparing the key
// derived by key. See [maps.UniqValuesBy]. Walks the map in ascending key order,
// so which value survives a collision is the same on every run.
func (m Map[K, V]) UniqValuesBy[R comparable](key func(V) R) Seq[V] {
	return Seq[V](maps.UniqValuesBy(m, key))
}

// SortedKeys keeps chaining as a [Slice] of the keys in ascending order. See
// [maps.SortedKeys].
func (m Map[K, V]) SortedKeys() Slice[K] {
	return maps.SortedKeys(m)
}

// SortedKeysFunc keeps chaining as a [Slice] of the keys ordered by compare. See
// [maps.SortedKeysFunc].
func (m Map[K, V]) SortedKeysFunc(compare func(K, K) int) Slice[K] {
	return maps.SortedKeysFunc(m, compare)
}

// SortedValuesFunc keeps chaining as a [Slice] of the values ordered by compare.
// See [maps.SortedValuesFunc].
//
// There is no comparator-free SortedValues, because that needs cmp.Ordered on V
// and a method cannot add a constraint its receiver lacks.
func (m Map[K, V]) SortedValuesFunc(compare func(V, V) int) Slice[V] {
	return maps.SortedValuesFunc(m, compare)
}

// Entries keeps chaining as a [Slice] of the entries in ascending key order. See
// [maps.Entries].
func (m Map[K, V]) Entries() Slice[maps.Entry[K, V]] {
	return maps.Entries(m)
}

// Reduce folds the entries into a single value, starting from init. The
// accumulator may be any type. See [maps.Reduce].
func (m Map[K, V]) Reduce[A any](init A, fn func(A, maps.Entry[K, V]) A) A {
	return maps.Reduce(m, init, fn)
}

// ReduceKeys folds the keys into a single value, starting from init. See
// [maps.ReduceKeys].
func (m Map[K, V]) ReduceKeys[A any](init A, fn func(A, K) A) A {
	return maps.ReduceKeys(m, init, fn)
}

// ReduceValues folds the values into a single value, starting from init. See
// [maps.ReduceValues].
func (m Map[K, V]) ReduceValues[A any](init A, fn func(A, V) A) A {
	return maps.ReduceValues(m, init, fn)
}
