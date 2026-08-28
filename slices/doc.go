// Package slices provides slice helpers that complement the standard library's
// slices package.
//
// Every function here takes or returns a slice. Functions that return a slice
// allocate it fresh rather than carving it out of the caller's, and nothing here
// modifies the slice it was given, so a caller can hold on to either one safely.
// A slice-returning function with nothing to return returns nil rather than an
// empty slice.
//
// The lazy adapters these are built on live in [github.com/pseudomuto/go/seq].
// The dependency runs one way: slices imports seq, never the reverse.
//
// This package does not re-export the standard library, so import both when you
// need both.
package slices
