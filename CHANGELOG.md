# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project follows
[semantic versioning](https://semver.org).

## [Unreleased]

## [0.2.0] - 2026-09-05

This release rewrites the matching engine for conformance with
[gitignore(5)](https://git-scm.com/docs/gitignore) and reworks the public API.
Both are breaking. A conformance suite now compares every rule against
`git check-ignore`, so the two implementations cannot drift apart unnoticed.

### Changed — matching semantics

Several patterns match differently than in 0.1.0. In each case the new
behaviour is what git does.

- **A pattern without a slash matches a whole path component, not a
  substring.** `name` used to match `/name.log` and `/lib/name.log`; it now
  matches the entry `name` and everything below it, at any depth. To keep the
  old reach, write `*name*`.
- **`?` matches exactly one character.** It used to match zero or one, so
  `name?.file` matched `/name.file`. The nearest equivalent is `name*.file`,
  which also accepts more than one character.
- **`*` and `?` no longer cross a separator.** `a*b` cannot match `a/x/b`; use
  `a/**/b`.
- **A negated set such as `[!abc]` no longer matches `/`.**
- **A pattern containing a slash is anchored, whether or not it looks like a
  file name.** Anchoring used to depend on the pattern containing a `.`, so
  `lib/name` was unanchored while `lib/name.file` was anchored. Both are now
  anchored, as git specifies.
- **The last matching rule decides.** A negation used to win outright, whatever
  followed it; `*.log`, `!keep.log`, `/vendor/keep.log` now excludes the vendored
  file.
- **A path below an excluded directory can never be re-included**, matching
  git's refusal to descend into an excluded directory. `build/` followed by
  `!build/keep.o` no longer re-includes the file; write `build/*` instead of
  `build/`.
- **A base is compared on path segments.** A base of `/home/ann` no longer
  matches paths under `/home/annex`.
- **Leading whitespace is significant.** Lines were trimmed before parsing, so
  `  #x` was read as a comment; git reads it as a pattern. Unescaped *trailing*
  spaces are still dropped, and `\ ` keeps one.

### Added

- `Pattern`, a single compiled rule, with `Compile`, `CompileBase` and
  `MustCompile` constructors and `Match`, `MatchDir`, `Negated`, `DirOnly`,
  `Base` and `String` methods.
- `MatchDir` on both `Pattern` and `Filter`, so directory-only rules can match
  the directory itself and not only its contents.
- Support for `\` escapes, POSIX character classes (`[[:digit:]]`), `[^...]` as
  a synonym for `[!...]`, and trailing `/**`.
- `ErrNoPattern`, `*SyntaxError` and `*ParseError`, so malformed input is
  reported instead of panicking. `*ParseError` carries the line number and
  unwraps to its cause.
- `Filter.Add`, `Filter.Patterns` and `Filter.Len`.
- `conformance_test.go`, which runs every rule set through both this library
  and `git check-ignore` and fails on any divergence. It skips when `git` is
  not on `PATH`.
- Package documentation, runnable examples, a `Makefile`, `golangci-lint`
  configuration and GitHub Actions for lint, tests across three Go versions and
  three operating systems, coverage and CodeQL.
- MIT license, contribution guide, security policy and code of conduct.

### Changed — API

- `NewFilter` now takes `...*Pattern` and returns `*Filter`. It previously took
  an unexported type, which made it unusable outside the package.
- `Filter.AddPattern` returns an `error` instead of silently accepting anything.
- `Filter.Parse` returns a `*ParseError` naming the offending line; rules read
  before the error are kept.
- The `prefix` argument is now called `base`, and is normalized on the way in.

### Fixed

- An invalid pattern no longer panics through `regexp.MustCompile`.
- A pattern's compiled regular expression is now cached. Memoization was written
  against a value receiver, so it was discarded and the expression recompiled on
  every call. Matching now allocates nothing.
- Regular-expression metacharacters in a pattern (`+`, `(`, `|`, …) are matched
  literally instead of being interpreted.

### Removed

- The unexported `filter` type is no longer part of any exported signature.

## [0.1.0] - 2025-02-24

Initial release.

[Unreleased]: https://github.com/sntns/globbing-go/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/sntns/globbing-go/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/sntns/globbing-go/releases/tag/v0.1.0
