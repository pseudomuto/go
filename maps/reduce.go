package maps

import (
	"cmp"

	"github.com/pseudomuto/go/seq"
)

// Reduce folds m's entries into a single value. It starts with init and calls fn
// once per entry, passing the running accumulator and the entry. The final
// accumulator is the result, so an empty or nil map yields init unchanged.
//
// fn takes an [Entry] rather than a separate key and value so that it names the
// two sides instead of ordering them. See [ReduceKeys] and [ReduceValues] for
// folding one side on its own.
//
// Reduce walks m in ascending key order, so an fn whose result depends on order
// gives the same answer on every run. That is why the keys have to be ordered
// rather than merely comparable.
func Reduce[T cmp.Ordered, U, V any](m map[T]U, init V, fn func(V, Entry[T, U]) V) V {
	return seq.Reduce(orderedEntries(m), init, fn)
}

// ReduceKeys folds m's keys into a single value. It starts with init and calls
// fn once per key, passing the running accumulator and the key. The final
// accumulator is the result, so an empty or nil map yields init unchanged.
//
// ReduceKeys walks m in ascending key order, so an fn whose result depends on
// order, appending to a slice or building a string say, gives the same answer on
// every run. That is why the keys have to be ordered rather than merely
// comparable.
func ReduceKeys[T cmp.Ordered, U, V any](m map[T]U, init V, fn func(V, T) V) V {
	return seq.Reduce(orderedKeys(m), init, fn)
}

// ReduceValues folds m's values into a single value. It starts with init and
// calls fn once per value, passing the running accumulator and the value. The
// final accumulator is the result, so an empty or nil map yields init unchanged.
//
// ReduceValues walks m in ascending key order, so an fn whose result depends on
// order, appending to a slice or building a string say, gives the same answer on
// every run. That is why the keys have to be ordered rather than merely
// comparable. Note the ordering is by key, not by value.
func ReduceValues[T cmp.Ordered, U, V any](m map[T]U, init V, fn func(V, U) V) V {
	return seq.Reduce(orderedValues(m), init, fn)
}
