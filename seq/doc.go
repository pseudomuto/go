// Package seq provides adapters for the iter.Seq iterators added in Go 1.23.
//
// Every adapter here is lazy. It returns a new iterator that does no work until
// something ranges over it, and it stops pulling from its source as soon as the
// consumer stops. Callbacks run once per value pulled, and not at all if the
// sequence is never consumed. That makes the adapters cheap to chain: values
// flow through the whole chain one at a time instead of being buffered between
// steps.
//
// [Reduce] is the exception. It is a consumer rather than an adapter, so it runs
// immediately and drains the sequence to produce one value.
//
// For the same operations over slices and maps, see
// [github.com/pseudomuto/go/slices] and [github.com/pseudomuto/go/maps].
package seq
