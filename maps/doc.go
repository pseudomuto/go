// Package maps provides map helpers that complement the standard library's maps
// package.
//
// Every function here takes or returns a map. Those that return a slice allocate
// it fresh, and nothing here modifies the map or slice it was given. A function
// with no values to return returns nil rather than an empty slice. The
// map-returning functions, [FromEntries], [Filter] and [Reject], are the
// exception: they always hand back a writable map, since a nil one would panic
// on the first write.
//
// Go randomizes map iteration order, and this package splits along that line.
// [Map], [MapErr], [MapKeys] and [MapValues] pass through whatever order the
// runtime hands them, so their results are only safe to treat as a set. Everything else imposes an
// order: the Sorted functions sort explicitly, while [Entries], [UniqKeysBy],
// [UniqValuesBy], [Reduce], [ReduceKeys], and [ReduceValues] all walk the map in
// ascending key order, so both the entry that survives a collision and the order
// values fold in are the same on every run.
//
// [Map], [MapErr], [Reduce], [Filter] and [Reject] hand each entry to their
// callback as an [Entry] so the callback names the key and the value rather than
// relying on their order.
//
// [Filter] and [Reject] sit outside the ordering split above. They return maps,
// which have no order to stabilise, so they need merely comparable keys rather
// than ordered ones. [Entries] and
// [FromEntries] convert between a map and a slice of [Entry] in either
// direction.
//
// The lazy adapters these are built on live in [github.com/pseudomuto/go/seq].
// The dependency runs one way: maps imports seq, never the reverse.
//
// This package does not re-export the standard library, so import both when you
// need both.
package maps
