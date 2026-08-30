package chain

import (
	"iter"
	stdslices "slices"

	"github.com/pseudomuto/go/seq"
)

// Seq is an [iter.Seq] with the seq package's adapters hung on it as methods, so
// a composition reads left to right instead of inside out.
//
// It is a defined type rather than a struct, so it can be ranged over directly.
// Unlike [Slice] and [Map] it needs an explicit conversion in both directions,
// because it and [iter.Seq] are both named types; [OfSeq] and [Seq.Unwrap] do
// that for you.
//
// T is unconstrained. There is no zero-argument Uniq, because deduplication needs
// a comparable key and a method cannot add a constraint its receiver lacks; use
// [Seq.UniqBy] with [Identity] instead.
type Seq[T any] iter.Seq[T]

// OfSeq converts s for chaining. Nothing is consumed until a terminal method
// runs, so the result stays as lazy as the sequence it came from.
func OfSeq[T any](s iter.Seq[T]) Seq[T] {
	return Seq[T](s)
}

// Filter keeps the values for which pred returns true. See [seq.Filter].
func (s Seq[T]) Filter(pred func(T) bool) Seq[T] {
	return Seq[T](seq.Filter(iter.Seq[T](s), pred))
}

// Reject drops the values for which pred returns true. See [seq.Reject].
func (s Seq[T]) Reject(pred func(T) bool) Seq[T] {
	return Seq[T](seq.Reject(iter.Seq[T](s), pred))
}

// UniqBy removes duplicates, comparing the key derived by key and keeping the
// first value seen for each. See [seq.UniqBy].
//
// Pass [Identity] to deduplicate on the values themselves, which requires T to be
// comparable at the call site.
func (s Seq[T]) UniqBy[U comparable](key func(T) U) Seq[T] {
	return Seq[T](seq.UniqBy(iter.Seq[T](s), key))
}

// Map applies fn to every value, and may change the element type. See [seq.Map].
func (s Seq[T]) Map[U any](fn func(T) U) Seq[U] {
	return Seq[U](seq.Map(iter.Seq[T](s), fn))
}

// MapErr is [Seq.Map] for an fn that can fail, yielding value and error pairs so
// the consumer decides whether to stop. See [seq.MapErr].
//
// This ends the chain at a bare [iter.Seq2], since there is no wrapper for pairs.
// Hand it to [github.com/pseudomuto/go/slices.CollectErr] to get a slice and the
// first error.
func (s Seq[T]) MapErr[U any](fn func(T) (U, error)) iter.Seq2[U, error] {
	return seq.MapErr(iter.Seq[T](s), fn)
}

// Reduce folds the sequence into a single value, starting from init. The
// accumulator may be any type. See [seq.Reduce].
func (s Seq[T]) Reduce[A any](init A, fn func(A, T) A) A {
	return seq.Reduce(iter.Seq[T](s), init, fn)
}

// Collect drains the sequence into a [Slice], which keeps chaining and is usable
// anywhere a []T is expected. An empty sequence collects to nil.
func (s Seq[T]) Collect() Slice[T] {
	return stdslices.Collect(iter.Seq[T](s))
}

// Unwrap returns the underlying [iter.Seq]. This one is not optional: both types
// are named, so plain assignment between them does not compile.
func (s Seq[T]) Unwrap() iter.Seq[T] {
	return iter.Seq[T](s)
}
