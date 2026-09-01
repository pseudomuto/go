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
| `mise run build`   | goreleaser snapshot. Writes `dist/`, which is gitignored |

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
everywhere, then what belongs to the collections packages, then what belongs to `validate`.

## Conventions for every package

**Naming suffixes.** These mean specific things, and are worth keeping consistent across packages:

| Suffix | The caller supplies                                          | Example                                    |
| ------ | ------------------------------------------------------------ | ------------------------------------------ |
| `Err`  | A callback that can fail                                     | `seq.MapErr`, `slices.MapErr`              |
| `By`   | A key extractor                                              | `seq.UniqBy(s, key)`                       |
| `Func` | A function the base form does not take, usually a comparator | `maps.SortedKeysFunc`, `validate.WhenFunc` |

`By` and `Func` are not interchangeable. A key extractor always takes `By`: the standard library's sorting family reads
`Func` as a comparator, so using it there will mislead people. `Func` is the fallback for anything else the caller hands
over, like the rule builder in `validate.WhenFunc`.

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

## The `validate` package

`validate` is unrelated to the four above and imports none of them. Its premise is that everything is spelled out at the
call site, which makes a few things load-bearing:

**No reflection, ever.** The field name and the value both arrive as arguments. That is what makes a mismatched check a
compile error and what keeps struct tags out of it. Needing `reflect` is a change to the premise rather than an
implementation detail, so raise it before writing it.

**Paths compose by prepending, one segment per level.** Segments accumulate and never merge or replace one another, so
two levels that pick the same name both appear, and an empty name contributes nothing. Do not add a case that collapses
or rewrites a segment, however much nicer the output looks. The predictability is the feature.

**New rule constructors should be sugar.** `Validate` is a `Group` at the root, `EachNested` is `Each` plus `Nested`,
and `When` is a `Group` handed no rules. Before writing a constructor that loops over `Errors` itself, check whether
composing the existing ones gets you there.

**Check messages complete the sentence "<field> ...".** `"is required"` and `"not greater than 0"`, never
`"name is required"`. `Field` supplies the name, so a message that repeats it reads doubled.

**Nothing short-circuits.** Every rule runs and every check within a rule runs, so one call reports every problem rather
than the first. A false `When` drops the rules it guards and nothing else.

## Testing conventions

- One test file per source file, named after it. `filter.go` gets `filter_test.go`
- Test functions in the same order as the functions they cover
- Tests live in the `_test` package (`package seq_test`) so they exercise the public API
- Table-driven with a named subtest per case, and `t.Parallel()` on both the parent and each subtest
- testify `require`, not `assert`, and build it from the subtest's own `t`
- `Example` functions are tests. They run in CI, so their `// Output:` has to be right
- Where output order is randomized, sort before comparing, or assert a count
- Fixtures and helpers shared by several of a package's test files live in one of them rather than being duplicated.
  `validate` keeps its structs and `pathsOf` in `validate_test.go`

Assertion messages are worth writing. `"the result shares a backing array with the input"` tells the next person what
broke; a bare failed comparison does not.

A passing test proves nothing until you have seen it fail. Break the implementation deliberately and check the test
catches it, especially for invariants that are invisible in the code, like "does not mutate its input" or "the returned
sequence can be ranged over twice".

## CI

Three jobs on every PR against `main`, and on every push to `main`: `build`, `test` and `lint`, each capped at 5
minutes. All three run the same mise tasks you run locally, so a green local run should mean a green CI run. `build` is
the goreleaser snapshot, which catches a broken `.goreleaser.yaml` on the PR rather than halfway through a release.

Coverage goes to Codecov. GitHub Actions are pinned to commit SHAs with the version in a trailing comment; if you bump
one, update both.

## Cutting a release

Releases are manual and need write access. In the Actions tab, pick the **release** workflow, run it against `main`, and
choose a bump:

| Bump    | What it does                                            |
| ------- | ------------------------------------------------------- |
| `auto`  | `svu next --always` picks the bump from the commit log  |
| `patch` | `svu patch`                                             |
| `minor` | `svu minor`                                             |
| `major` | `svu major`                                             |

The workflow resolves the version, creates and pushes an annotated tag, then hands off to goreleaser. Do not tag by
hand. The tag is the workflow's output, and tagging out of band leaves the two out of step.

### What `auto` reads

`svu next` walks the commits since the last tag and maps conventional commit prefixes onto a bump:

| Commit                                   | Bump  |
| ---------------------------------------- | ----- |
| `feat:`                                  | minor |
| `fix:`                                   | patch |
| `feat!:` or a `BREAKING CHANGE:` trailer | major |
| anything else (`build:`, `docs:`, ...)   | none  |

`--always` turns that last row into a patch bump, so a release carrying only chores still gets a version instead of
reusing the current one. Nothing enforces the prefixes at commit time, so `auto` is only as honest as the log. Run `svu
current` and `svu next` locally first if you want to see what it will pick.

This module is still on `v0`, and `auto` will happily take a breaking change to `v1.0.0`. Choose the bump explicitly if
that is not what you meant.

### What ends up in the release

A GitHub release, with a changelog GitHub generates from the commits in range, and no attached files. This is a library,
so `go get` is the install path and there is nothing to download. `mise run build` runs goreleaser locally in snapshot
mode if you want to watch the config execute without publishing anything.

### If the release job fails

The tag is pushed before goreleaser runs, so a mid-job failure can leave a tag behind with no release attached, and
re-running the workflow will then fail on the tag already existing. Clear it and start over:

```bash
git push --delete origin vX.Y.Z
git tag -d vX.Y.Z
```
