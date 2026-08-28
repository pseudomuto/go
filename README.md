# go

[![CI](https://img.shields.io/github/actions/workflow/status/pseudomuto/go/ci.yaml?branch=main)](https://github.com/pseudomuto/go/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/pseudomuto/go.svg)](https://pkg.go.dev/github.com/pseudomuto/go)
[![Coverage](https://img.shields.io/codecov/c/github/pseudomuto/go)](https://codecov.io/gh/pseudomuto/go)

A collection of small, self-contained packages for Go applications. Each is independent, so you import only what you
need.

Today that means helpers for iterators, slices, and maps. The module is not limited to those: it is a home for whatever
turns out to be (whatever I deem) worth reusing across projects, and each new package gets its own top-level directory.

## Install

```bash
go get github.com/pseudomuto/go
```

Requires Go 1.26.5 or later, per `go.mod`.

## Getting started

```go
package main

import (
	"fmt"
	"slices"

	"github.com/pseudomuto/go/seq"
)

func main() {
	words := slices.Values([]string{"go", "is", "fun", "go", "again"})

	// Nothing runs until the range loop starts pulling.
	short := seq.Filter(seq.Uniq(words), func(w string) bool { return len(w) < 4 })

	for w := range short {
		fmt.Println(w)
	}
	// go
	// is
	// fun
}
```

Same job against a slice, without the iterator plumbing:

```go
import "github.com/pseudomuto/go/slices"

unique := slices.UniqBy(users, func(u User) string { return u.Name })
```

## Packages

What is here today. These three form one family, with a rule for where things go: `seq` holds the lazy adapters over
`iter.Seq`, while `slices` and `maps` hold anything with a slice or a map on either side. The dependency runs one way,
so `slices` and `maps` import `seq` and never the reverse.

### seq

Lazy adapters over `iter.Seq`. Each returns a new iterator that does no work until something ranges over it, and stops
pulling from its source the moment the consumer stops.

| Function | What it does                                                                   |
| -------- | ------------------------------------------------------------------------------ |
| `Map`    | Applies `fn` to every value                                                    |
| `MapErr` | `Map` for a fallible `fn`; yields `(value, error)` pairs so the caller decides |
| `Filter` | Keeps values where `pred` returns true                                         |
| `Reject` | Drops values where `pred` returns true, the complement of `Filter`             |
| `Uniq`   | Removes duplicates, keeping the first occurrence                               |
| `UniqBy` | Same, comparing a derived key, so the values themselves need not be comparable |
| `Reduce` | Folds the sequence into a single value. Eager, unlike everything else here     |

### slices

| Function     | What it does                                                                      |
| ------------ | --------------------------------------------------------------------------------- |
| `Map`        | Applies `fn` to every element. Allocates once, since the length is known          |
| `MapErr`     | `Map` for a fallible `fn`. All or nothing: the first error, and no partial result |
| `Filter`     | Keeps elements where `pred` returns true                                          |
| `Reject`     | Drops elements where `pred` returns true                                          |
| `Uniq`       | Removes duplicates, keeping the first occurrence                                  |
| `UniqBy`     | Same, comparing a derived key                                                     |
| `Reduce`     | Folds the slice into a single value                                               |
| `CollectErr` | Collects a `Seq2` of `(value, error)` into a slice, stopping at the first error   |

### maps

| Function                                 | What it does                                       |
| ---------------------------------------- | -------------------------------------------------- |
| `Entries` / `FromEntries`                | Convert between a map and a sorted `[]Entry`       |
| `Map`                                    | Apply `fn` to each entry                           |
| `MapErr`                                 | `Map` for a fallible `fn`; yields `(value, error)` |
| `MapKeys` / `MapValues`                  | Apply `fn` to each key or value                    |
| `Filter` / `Reject`                      | Keep or drop entries by predicate, into a new map  |
| `SortedKeys` / `SortedValues`            | Keys or values as a sorted slice                   |
| `SortedKeysFunc` / `SortedValuesFunc`    | Same, ordered by a caller-supplied comparator      |
| `UniqKeysBy` / `UniqValuesBy`            | Deduplicate keys or values by a derived key        |
| `Reduce` / `ReduceKeys` / `ReduceValues` | Fold entries, keys, or values into a single value  |

`Reduce` hands each entry to its callback as an `Entry{Key, Value}` rather than as two positional arguments, so a
`map[string]string` callback cannot silently transpose them.

## Design notes

Contracts for the three packages above, worth knowing before you reach for something. They are not module-wide rules:
anything added later states its own contracts in its package doc.

**Laziness.** Every `seq` adapter is lazy and composes without buffering, so values flow through a whole chain one at a
time. Callbacks run once per value pulled, and not at all if the sequence is never consumed. `Reduce` is the exception:
it is a consumer, so it runs immediately and drains its source.

**Map iteration order.** Go randomizes it, and these packages split along that line. `MapKeys` and `MapValues` pass the
runtime's order straight through, so treat their results as a set. Everything else imposes an order: the `Sorted`
functions sort explicitly, and `Entries`, `UniqKeysBy`, `UniqValuesBy` and the `Reduce` family all walk the map in
ascending key order. That is why those take `cmp.Ordered` keys rather than merely `comparable` ones. It buys you a
result that is identical on every run, including which entry survives a collision.

**No surprise mutation.** Nothing modifies the slice or map it was given, and slice results are freshly allocated rather
than carved out of the caller's backing array.

**Empty results.** A slice-returning function with nothing to return returns nil. `FromEntries` is the exception and
always hands back a writable map, since a nil map panics on the first write.

**No stdlib re-exports.** These packages shadow `slices` and `maps` by name but do not wrap them, so import both when
you need both.

## Contributing

See [CONTRIBUTING.md](./.github/CONTRIBUTING.md) for setup, the conventions review will hold you to, and what CI checks.

## License

MIT. See [LICENSE](./LICENSE).
