package chain

import (
	stdslices "slices"

	"github.com/pseudomuto/go/slices"
)

// Slice is a []T with the slices package's helpers hung on it as methods, so a
// composition reads left to right instead of inside out.
//
// It is a defined type rather than a struct, so it *is* a slice: index it, take
// its len, range over it, and pass it to anything expecting a []T with no
// conversion. That is why there is no Unwrap or Len method here.
//
//	got := chain.OfSlice(users).Filter(isActive).UniqBy(byName)
//	fmt.Println(len(got), got[0])
//	var plain []User = got
//
// T is unconstrained. There is no zero-argument Uniq, because deduplication needs
// a comparable key and a method cannot add a constraint its receiver lacks; use
// [Slice.UniqBy] with [Identity] instead.
type Slice[T any] []T

// OfSlice converts sl for chaining. It is the same slice, not a copy, so the two
// share a backing array. Every method that transforms returns a freshly allocated
// result, leaving sl alone.
func OfSlice[T any](sl []T) Slice[T] {
	return Slice[T](sl)
}

// Filter keeps the elements for which pred returns true. See [slices.Filter].
func (s Slice[T]) Filter(pred func(T) bool) Slice[T] {
	return slices.Filter(s, pred)
}

// Reject drops the elements for which pred returns true. See [slices.Reject].
func (s Slice[T]) Reject(pred func(T) bool) Slice[T] {
	return slices.Reject(s, pred)
}

// UniqBy removes duplicates, comparing the key derived by key and keeping the
// first element seen for each. See [slices.UniqBy].
//
// Pass [Identity] to deduplicate on the elements themselves, which requires T to
// be comparable at the call site.
func (s Slice[T]) UniqBy[U comparable](key func(T) U) Slice[T] {
	return slices.UniqBy(s, key)
}

// Map applies fn to every element, and may change the element type. See
// [slices.Map].
func (s Slice[T]) Map[U any](fn func(T) U) Slice[U] {
	return slices.Map(s, fn)
}

// MapErr is [Slice.Map] for an fn that can fail. It returns the mapped slice, or
// the first error fn produced. See [slices.MapErr].
//
// It is all or nothing: on error the slice is nil, discarding whatever mapped
// cleanly before it.
func (s Slice[T]) MapErr[U any](fn func(T) (U, error)) (Slice[U], error) {
	return slices.MapErr(s, fn)
}

// Reduce folds the slice into a single value, starting from init. The accumulator
// may be any type. See [slices.Reduce].
func (s Slice[T]) Reduce[A any](init A, fn func(A, T) A) A {
	return slices.Reduce(s, init, fn)
}

// SortFunc returns the elements ordered by compare, which must return a negative
// number when a sorts before b, zero when they tie, and a positive number when a
// sorts after b. The receiver is left unsorted.
//
// There is no comparator-free Sort, because sorting needs cmp.Ordered on T and a
// method cannot add a constraint its receiver lacks.
func (s Slice[T]) SortFunc(compare func(T, T) int) Slice[T] {
	return stdslices.SortedFunc(stdslices.Values(s), compare)
}

// Seq continues the chain lazily, as a [Seq] over the elements.
func (s Slice[T]) Seq() Seq[T] {
	return Seq[T](stdslices.Values(s))
}
