# go

[![CI](https://img.shields.io/github/actions/workflow/status/pseudomuto/go/ci.yaml?branch=main)](https://github.com/pseudomuto/go/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/pseudomuto/go.svg)](https://pkg.go.dev/github.com/pseudomuto/go)
[![Coverage](https://img.shields.io/codecov/c/github/pseudomuto/go)](https://codecov.io/gh/pseudomuto/go)

A collection of small, self-contained packages for Go applications. Each is independent, so you import only what you
need.

Today that means helpers for iterators, slices and maps, plus a validation package. The module is not limited to those:
it is a home for whatever turns out to be (whatever I deem) worth reusing across projects, and each new package gets its
own top-level directory.

## Install

```bash
go get github.com/pseudomuto/go
```

Requires Go 1.27 or later, per `go.mod`. The `chain` package needs 1.27 specifically, for methods with type parameters.

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

What is here today. `seq`, `slices` and `maps` form one family with a rule for where things go: `seq` holds the lazy
adapters over `iter.Seq`, while `slices` and `maps` hold anything with a slice or a map on either side. The dependency
runs one way, so `slices` and `maps` import `seq` and never the reverse. `chain` puts the same helpers on the containers
themselves, as methods.

`validate` stands on its own and shares nothing with those four.

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

### chain

The same helpers as the three packages above, hung on the containers themselves as methods, for when the nesting gets
awkward.

```go
total := chain.OfMap(m).
	Filter(isActive).
	SortedKeys().          // -> chain.Slice[K]
	Map(strings.ToUpper).
	Seq().                 // -> chain.Seq[K]
	Reduce("", concat)
```

`Slice` is defined as `[]T` and `Map` as `map[K]V`, so they _are_ the containers. Index them, take their `len`, range
over them, and pass them anywhere a plain slice or map is expected, with no unwrapping:

```go
got := chain.OfMap(raw).Filter(isActive).SortedKeys()

fmt.Println(len(got), got[0])
var plain []string = got
```

`Seq` is the exception: it and `iter.Seq` are both named types, so it needs an explicit conversion in each direction.
`OfSeq` and `Seq.Unwrap` are those conversions. Ranging over a `Seq` directly works either way.

`chain` adds no behaviour of its own. Every method is a one-line delegation, so the contracts below still hold.

Because Go 1.27 lets methods declare type parameters, the chain does not break at a type change. `Map` may return a
different element type and `Reduce` takes any accumulator, so a whole pipeline stays in one expression:

```go
report := chain.OfMap(users).
	Filter(isActive).
	SortedKeys().                                  // -> Slice[string]
	Map(func(k string) User { return lookup(k) }). // -> Slice[User]
	UniqBy(func(u User) string { return u.Email }).
	Reduce(Summary{}, addRow)                      // -> Summary
```

> [!NOTE]
>
> What a method still cannot do is add a constraint its receiver lacks. `Slice` and `Seq` leave their element type
> unconstrained so they can hold anything, so there is no zero-argument `Uniq`: pass `chain.Identity` to `UniqBy` to
> dedupe on the values themselves. Likewise no comparator-free `Sort`, since that needs `cmp.Ordered`. The package doc
> lists the full set.

### validate

Declarative struct validation with no reflection. You name the field and hand over its value, so there are no struct
tags to parse, nothing is looked up by name at runtime, and a check that does not fit the field's type is a compile
error.

```go
func (u User) Validate() error {
	return validate.Validate("user",
		validate.Field("name", u.Name, validate.Required[string]()),
		validate.Field("age", u.Age, validate.GTE(0), validate.LT(150)),
		validate.Nested("address", u.Address),
		validate.When(u.Notify,
			validate.Field("email", u.Email, validate.Required[string]()),
		),
	)
}

// user.name: is required
// user.age: not less than 150
// user.address.street: is required
// user.email: is required
```

A `Check` is any `func(T) error`. A `Rule` binds values to checks, and `Validate` runs rules and returns nil, or an
`Errors` holding one `Error` per failure.

| Rule              | What it does                                                          |
| ----------------- | --------------------------------------------------------------------- |
| `Field`           | Binds a name and a value to the checks it has to pass                 |
| `Group`           | Nests a set of rules under a name                                     |
| `Nested`          | Hands a value to its own `Validate` method                            |
| `Each`            | Applies a rule to every element of a slice, under a `name[i]` segment |
| `EachNested`      | `Each` plus `Nested`, for a slice whose elements validate themselves  |
| `When` / `Unless` | Guards rules behind a condition                                       |
| `WhenFunc`        | `When`, but the rule is only built if the condition holds             |

| Check                              | What it does                                                           |
| ---------------------------------- | ---------------------------------------------------------------------- |
| `Required`                         | Rejects the zero value                                                 |
| `GT` / `GTE` / `LT` / `LTE`        | Compares against a bound                                               |
| `Unique`                           | Rejects a slice holding the same value twice                           |
| `IsIP` / `IsIPv4` / `IsIPv6`       | Parses as an IP address, optionally pinned to one family               |
| `IsCIDR` / `IsCIDRv4` / `IsCIDRv6` | Parses as a CIDR prefix; the `vN` forms also require a network address |

Every failure carries a path, built by prepending one segment per level as it travels back out. Nothing is
special-cased, which is what makes it predictable:

```
a check reports          is required
Field("street", ...)     street: is required
Group("address", ...)    address.street: is required
Validate("user", ...)    user.address.street: is required
```

`Errors` is a list, not a string. It unwraps to its elements, so `errors.As` reaches either the whole list or a single
`Error`, and each one exposes its path separately from its message. Callers that attach failures to form inputs never
have to parse anything:

```go
var verrs validate.Errors
if errors.As(err, &verrs) {
	for _, e := range verrs {
		fmt.Println(e.Path(), "->", e.Error())
	}
}
```

## Design notes

Contracts for `seq`, `slices` and `maps`, worth knowing before you reach for something. `chain` delegates to them, so
they hold there too. They are not module-wide rules: `validate` states its own in its package doc, as will anything
added later.

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
