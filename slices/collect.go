package slices

import "iter"

// CollectErr collects the values from seq into a slice, stopping at the first
// error.
//
// It is all or nothing: on error CollectErr returns a nil slice along with that
// error, discarding the values it had already collected. That way a caller who
// ignores the error gets an obviously empty result rather than silently
// truncated data. To keep the values that did succeed, range over the
// [iter.Seq2] yourself and decide pair by pair.
//
// The error is returned unwrapped, so errors.Is and errors.As match against
// whatever the producer put in the pair.
func CollectErr[T any](seq iter.Seq2[T, error]) ([]T, error) {
	var out []T

	for v, err := range seq {
		if err != nil {
			return nil, err
		}

		out = append(out, v)
	}

	return out, nil
}
