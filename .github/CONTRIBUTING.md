# Contributing

Thanks for taking a look. This module is a collection of small, independent packages, so the bar is less about process
and more about keeping each package coherent and its documented contracts honest.

## Getting set up

You need [git](https://git-scm.com) and [mise](https://mise.jdx.dev). mise pins everything else, so you do not need Go
installed separately.

```bash
git clone https://github.com/pseudomuto/go.git
cd go
mise trust     # mise refuses to run hooks from an untrusted config
mise install   # installs Go, golangci-lint, lefthook, and the git hooks
mise run test
```

`mise install` runs `lefthook install` for you, so a fresh clone gets working git hooks. The `pre-push` hook runs the
linter; nothing runs on commit.

If `mise install` skipped the hooks, you forgot `mise trust`. Run it and install again.

## Everyday commands

| Command            | What it does                                             |
| ------------------ | -------------------------------------------------------- |
| `mise run test`    | `go test -race -cover ./...`                             |
| `mise run lint`    | `golangci-lint run ./...`                                |
| `mise run format`  | `go fix ./...` then `golangci-lint fmt`                  |
| `mise run test:ci` | What CI runs. Writes `coverage.out`, which is gitignored |

## Before you open a PR

```bash
mise run format
mise run lint
mise run test
```

Then check:

- [ ] Coverage is still 90%+ of statements. `go test -coverprofile=/tmp/c.out ./... && go tool cover -func=/tmp/c.out`
- [ ] Every new exported function has a doc comment and a runnable `Example`
- [ ] Any behaviour your doc comment promises has a test that fails without it
- [ ] Package docs still tell the truth, if you added something that breaks a stated invariant

That last one bites more often than you would expect. Adding the first eager function to a package documented as lazy,
or the first map-returning function to a package documented as returning nil for nothing, means the package doc needs
amending too.

## Repository layout

One package per top-level directory, each independent and importable on its own. A new package means a new directory at
the root, with its own `doc.go` and its own tests. There is no shared `internal` package; adding one should be a
deliberate decision rather than a convenience.

Every package states its contracts in its own package doc. The conventions below are split the same way: what holds
everywhere, then what belongs to the collections packages that happen to be here today.

## Conventions for every package

**Naming suffixes.** These mean specific things, and are worth keeping consistent across packages:

| Suffix | The caller supplies      | Example                       |
| ------ | ------------------------ | ----------------------------- |
| `Err`  | A callback that can fail | `seq.MapErr`, `slices.MapErr` |
| `By`   | A key extractor          | `seq.UniqBy(s, key)`          |
| `Func` | A comparator             | `maps.SortedKeysFunc`         |

`By` and `Func` are not interchangeable. Standard library `*Func` variants take comparators, so using `Func` for a key
extractor will mislead people.

**The package doc carries the contracts.** Anything a caller could get wrong belongs there: what is lazy, what
allocates, what order results arrive in, what happens on empty input. If you add something that breaks a stated
invariant, amend the statement in the same change.

**Empty results.** A slice-returning function with nothing to return returns nil. A map-returning function always
returns a writable map, because a nil map panics on the first write.

**No surprise mutation.** Nothing modifies the slice or map it was given, and slice results are freshly allocated rather
than carved out of the caller's backing array.

**Do not shadow the standard library without a reason.** `slices` and `maps` already do, deliberately, and pay for it by
needing an import alias wherever a file wants both. Do not add a third unless the name really is the right one.

## The collections packages

`seq`, `slices`, `maps` and `chain` are one family, and they have a placement rule of their own:

| Package  | Owns                                                   |
| -------- | ------------------------------------------------------ |
| `seq`    | Lazy adapters over `iter.Seq`. Seq in, Seq or Seq2 out |
| `slices` | Anything with a slice on either side                   |
| `maps`   | Anything with a map on either side                     |
| `chain`  | The other three's helpers, as methods on the containers |

The dependency runs one way. `slices` and `maps` import `seq`; `seq` imports neither. Keep it that way or you get an
import cycle the first time someone adds a conversion.

`seq` holds the primitives, and `slices` and `maps` wrap them where one fits. Plenty of operations have no `seq` analog,
though: producing a map, collecting a `Seq2` into a slice, sorting. Those use the standard library or a local loop,
which is fine and needs no explanation.

What does need explaining is declining to delegate when you could have. `maps.Map` and `maps.MapErr` range the map
directly, because routing through `maps.Keys` costs a second hash lookup per entry, measured at roughly 2x on a
100k-entry map. `maps.Filter` hand-rolls because routing through `Entries` would drag in a `cmp.Ordered` constraint it
does not otherwise need. All three carry a comment saying so. If you skip delegation, leave the reason and the evidence
in the code, or the next person will helpfully clean it up and hand the time back.

**Entry, not two positional arguments.** Callbacks that see a whole map entry take a `maps.Entry`, so they name the key
and the value instead of ordering them. On a `map[string]string` a transposed positional callback compiles fine and
fails silently.

**Laziness in `seq`.** Adapters do no work until something ranges over them, and stop pulling from their source as soon
as the consumer stops. Callbacks run once per value pulled. `Reduce` is the one consumer, and it says so.

**Determinism in `maps`.** Go randomizes map iteration order. `Map`, `MapErr`, `MapKeys` and `MapValues` pass that order
through, so their results are only safe to treat as a set. Everything else imposes an order, walking the map in
ascending key order, which is why those take `cmp.Ordered` keys rather than merely `comparable` ones. If you add a
function whose result depends on order, sort; do not hand back a value that differs between runs.

**`chain` adds no algorithms.** Every method is a one-line delegation to an exported function in `seq`, `slices`, `maps`
or the standard library. If a method would need new logic, that function belongs in the container package first, as its
own change. This is mechanically checkable: no method body in `chain` should contain a `for`, an `if`, or more than one
statement.

**`chain` never returns a randomly ordered slice.** Materialising map order into a slice is the nondeterminism the
`maps` package was fixed to avoid, so the map-to-slice crossing is offered only in sorted form. Sequences are exempt,
since an `iter.Seq` is unordered by definition.

**`chain` types are defined types, not structs.** `Slice` is `[]T` and `Map` is `map[K]V`, so they are indexable,
rangeable and assignable to the plain containers with no conversion. That is why they have no `Unwrap` or `Len`. `Seq`
is the exception: it and `iter.Seq` are both named, so it needs an explicit conversion, which `OfSeq` and `Seq.Unwrap`
provide. They cannot be aliases, because an alias cannot have methods.

**`chain` needs Go 1.27.** Methods with type parameters landed there, which is what lets `Map` change the element type
and `Reduce` take any accumulator. Note the linter has to match: `golangci-lint` built against Go 1.26 refuses a module
targeting 1.27 outright, so run it through `mise` rather than whatever is on your `PATH`.

**Methods still cannot add a constraint the receiver lacks.** That is why there is no zero-argument `Uniq` (use
`UniqBy` with `chain.Identity`), no comparator-free `Sort`, and why `chain.Map` requires ordered keys. If you find
yourself wanting one of those, it is a language limit rather than an oversight. `chain/doc.go` explains it with the
exact compiler errors, and is the place to point people.

## Testing conventions

- One test file per source file, named after it. `filter.go` gets `filter_test.go`
- Test functions in the same order as the functions they cover
- Tests live in the `_test` package (`package seq_test`) so they exercise the public API
- Table-driven with a named subtest per case, and `t.Parallel()` on both the parent and each subtest
- testify `require`, not `assert`, and build it from the subtest's own `t`
- `Example` functions are tests. They run in CI, so their `// Output:` has to be right
- Where output order is randomized, sort before comparing, or assert a count

Assertion messages are worth writing. `"the result shares a backing array with the input"` tells the next person what
broke; a bare failed comparison does not.

A passing test proves nothing until you have seen it fail. Break the implementation deliberately and check the test
catches it, especially for invariants that are invisible in the code, like "does not mutate its input" or "the returned
sequence can be ranged over twice".

## CI

Two jobs on every PR against `main`, and on every push to `main`: `test` and `lint`, each capped at 5 minutes. Both run
the same mise tasks you run locally, so a green local run should mean a green CI run.

Coverage goes to Codecov. GitHub Actions are pinned to commit SHAs with the version in a trailing comment; if you bump
one, update both.
