package seq

import "iter"

// Filter returns a sequence of the values in seq for which pred returns true.
// pred runs for every value in seq, including those that never reach the
// consumer.
func Filter[T any](seq iter.Seq[T], pred func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if pred(v) {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// Reject returns a sequence of the values in seq for which pred returns false.
// It is the complement of [Filter] under the same pred: for a pred that answers
// consistently, the two together cover every value in seq exactly once.
func Reject[T any](seq iter.Seq[T], pred func(T) bool) iter.Seq[T] {
	return Filter(seq, func(t T) bool { return !pred(t) })
}
