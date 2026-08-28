package seq

import "iter"

// Reduce folds seq into a single value. It starts with init and calls fn once
// per value, left to right, passing the running accumulator and the value. The
// final accumulator is the result, so an empty seq yields init unchanged.
//
// Reduce is a consumer, not an adapter: it runs immediately and drains seq
// rather than returning a lazy sequence. It has no early exit, so reducing an
// endless sequence never returns. Filter or otherwise bound the sequence first.
//
// The accumulator type is free to differ from the value type, and Go infers it
// from init, so counting words into a map or joining ints into a string needs no
// explicit type arguments.
func Reduce[T, U any](seq iter.Seq[T], init U, fn func(U, T) U) U {
	acc := init
	for v := range seq {
		acc = fn(acc, v)
	}

	return acc
}
