// Package chain hangs the seq, slices and maps helpers on the containers
// themselves as methods, so a composition reads left to right instead of inside
// out.
//
//	// nested
//	total := maps.ReduceValues(maps.Filter(m, isActive), 0, add)
//
//	// chained
//	total := chain.OfMap(m).Filter(isActive).ReduceValues(0, add)
//
// Convert a value with [OfSeq], [OfSlice] or [OfMap], chain methods on it, and
// end wherever you like.
//
// This package needs Go 1.27, which allows methods to declare type parameters.
// Before that, no method could change the element type, and the chain had to
// break at every such step.
//
// # The chain does not break at a type change
//
// [Slice.Map], [Seq.Map], [Map.MapKeys] and friends may return a different
// element type, and the Reduce methods take any accumulator, so a whole pipeline
// stays in one expression:
//
//	report := chain.OfMap(users).
//		Filter(isActive).
//		SortedKeys().                                  // -> Slice[string]
//		Map(func(k string) User { return lookup(k) }). // -> Slice[User]
//		UniqBy(func(u User) string { return u.Email }).
//		Reduce(Summary{}, addRow)                      // -> Summary
//
// # These are the containers, not wrappers around them
//
// [Slice] is defined as []T and [Map] as map[K]V, so they *are* a slice and a map.
// Index them, take their len, range over them, write composite literals, and pass
// them anywhere a plain []T or map[K]V is expected, with no conversion and no
// Unwrap:
//
//	got := chain.OfMap(raw).Filter(isActive).SortedKeys()
//
//	fmt.Println(len(got), got[0])
//	var plain []string = got
//	m := chain.Map[string, int]{"a": 1}
//
// [Seq] is the exception. It is defined over [iter.Seq], and Go will not assign
// between two named types, so it needs an explicit conversion in each direction.
// [OfSeq] and [Seq.Unwrap] are those conversions. Ranging over a [Seq] directly
// works fine.
//
// These are defined types rather than aliases because an alias still cannot have
// methods: "cannot define new methods on generic alias type".
//
// # What is still out of reach
//
// A method may declare its own type parameters, but it cannot add a constraint to
// one the receiver already holds. So an operation needing something of the
// element type itself, rather than of a key derived from it, cannot be a method:
//
//   - No zero-argument Uniq. Deduplication needs a comparable key, and [Slice] and
//     [Seq] deliberately leave T unconstrained so they can hold anything. Use
//     [Slice.UniqBy] or [Seq.UniqBy], passing [Identity] to dedupe on the values
//     themselves.
//   - No comparator-free Sort, and no SortedValues. Those need cmp.Ordered. Use
//     [Slice.SortFunc] and [Map.SortedValuesFunc].
//   - [Map] does require ordered keys, because the deterministic half of the maps
//     package does. Struct, array and pointer keys are out; those callers use the
//     maps functions directly.
//
// Interface methods also still cannot have type parameters, so none of this is
// reachable through an interface.
//
// # Ordering
//
// No method here returns a randomly ordered slice. Go randomizes map iteration
// order, and materialising that into a slice is the nondeterminism the maps
// package was careful to remove, so the map-to-slice crossings are sorted:
// [Map.SortedKeys], [Map.SortedKeysFunc], [Map.SortedValuesFunc] and
// [Map.Entries].
//
// Sequences are exempt, since an [iter.Seq] is unordered by definition.
// [Map.Map], [Map.MapKeys] and [Map.MapValues] return a [Seq] in map order, and
// [Seq.Collect] on one of those does produce an unordered slice. That is the
// caller asking for it explicitly, exactly as slices.Collect(maps.MapKeys(...))
// is today.
//
// # This package adds no behaviour
//
// Every method is a one-line delegation to an exported function in
// [github.com/pseudomuto/go/seq], [github.com/pseudomuto/go/slices],
// [github.com/pseudomuto/go/maps] or the standard library. The contracts,
// including laziness, allocation and determinism, are documented on those
// functions and hold unchanged here. If an operation would need logic this
// package does not have, it belongs in the container package first.
package chain
