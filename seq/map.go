package seq

import "iter"

// Map returns a sequence of the values in seq with fn applied to each.
func Map[T, U any](seq iter.Seq[T], fn func(T) U) iter.Seq[U] {
	return func(yield func(U) bool) {
		for v := range seq {
			if !yield(fn(v)) {
				return
			}
		}
	}
}

// MapErr is [Map] for a fn that can fail. It returns a sequence of value and
// error pairs, one per value in seq.
//
// When fn fails, the pair holds the zero value of U and fn's error, in the
// position of the value that failed. MapErr itself never stops on an error, it
// keeps pulling from seq, so the consumer chooses whether to break on the first
// failure or skip it and carry on.
func MapErr[T, U any](seq iter.Seq[T], fn func(T) (U, error)) iter.Seq2[U, error] {
	return func(yield func(U, error) bool) {
		for v := range seq {
			if !yield(fn(v)) {
				return
			}
		}
	}
}
