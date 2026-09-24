// Package is holds checks that ask what a value looks like.
//
// The package name is doing work. A check here is called where it reads as a
// sentence about the field:
//
//	validate.Field("addr", cfg.Addr, is.IP())
//	validate.Field("cidr", cfg.Net, is.CIDRv4())
//
// The parent package keeps the checks whose names already read that way alone:
// [validate.Required], and the comparators [validate.GT] and friends. Formats are
// nouns, and a bare validate.IP() would read as something returning an IP rather
// than something testing for one. Putting the verb in the package name fixes that
// without an Is prefix on every function.
//
// Everything here returns a [validate.Check], so these compose with the rest of
// the package like any other check, and one you write yourself sits alongside
// them with no ceremony.
//
// Messages follow the parent's convention: they complete the sentence
// "<field> ...", so a failure formats as "addr: is not a valid IP address".
package is
