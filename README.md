# globbing-go

[![CI](https://github.com/sntns/globbing-go/actions/workflows/ci.yml/badge.svg)](https://github.com/sntns/globbing-go/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/sntns/globbing-go/branch/main/graph/badge.svg)](https://codecov.io/gh/sntns/globbing-go)
[![Go Reference](https://pkg.go.dev/badge/github.com/sntns/globbing-go.svg)](https://pkg.go.dev/github.com/sntns/globbing-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/sntns/globbing-go)](https://goreportcard.com/report/github.com/sntns/globbing-go)
[![Go version](https://img.shields.io/github/go-mod/go-version/sntns/globbing-go)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Gitignore-style path matching for Go, with no dependencies outside the standard
library.

The package implements [gitignore(5)](https://git-scm.com/docs/gitignore) and is
checked against git itself: [`conformance_test.go`](conformance_test.go) runs
every rule through both this library and `git check-ignore` and fails on any
divergence.

```go
import globbing "github.com/sntns/globbing-go"
```

```console
go get github.com/sntns/globbing-go
```

## Usage

### A single rule

```go
p, err := globbing.Compile("doc/**/*.md")
if err != nil {
    return err
}

p.Match("/doc/api/index.md") // true
p.Match("/src/README.md")    // false
```

### A set of rules

`Filter` evaluates rules the way git does — the last rule that matches decides:

```go
f := globbing.NewFilter()
if err := f.Parse("", strings.NewReader("*.log\n!keep.log\n")); err != nil {
    return err
}

f.Match("/var/app.log")  // true
f.Match("/var/keep.log") // false
```

### Rules from a subdirectory

A `.gitignore` in a subdirectory only governs that subtree. Pass the directory
as the *base*:

```go
f := globbing.NewFilter()
f.Parse("", rootGitignore)               // /.gitignore
f.Parse("/vendor", vendorGitignore)      // /vendor/.gitignore
```

The base is compared on whole path segments, so a base of `/home/ann` never
swallows `/home/annex`.

### Files and directories

Whether a path denotes a directory cannot be read off the path itself, so the
caller says which it is. This only matters for directory-only rules:

```go
f := globbing.NewFilter(globbing.MustCompile("build/"))

f.MatchDir("/build")      // true  — the directory itself
f.Match("/build")         // false — a *file* named build is not covered
f.Match("/build/main.o")  // true  — anything below it is
```

## Pattern syntax

| Syntax | Meaning |
| --- | --- |
| `#…` | comment (only when `#` is the first character of the line) |
| `!…` | negation: re-include a path an earlier rule excluded |
| `/` leading or inner | anchor the rule to its base directory |
| `/` trailing | the rule only applies to directories |
| `*` | any run of characters, never crossing `/` |
| `?` | exactly one character, never `/` |
| `[abc]`, `[a-z]` | one of the listed characters |
| `[!abc]`, `[^abc]` | one character outside the set, never `/` |
| `[[:digit:]]` | one character of the POSIX class |
| `**/x` | `x` at any depth |
| `x/**` | everything below `x` |
| `a/**/b` | `b` below `a`, at any depth |
| `\` | escape the next character, including `#`, `!` and a trailing space |

A pattern with no slash (other than a trailing one) matches an entry of that
name at any depth. A pattern containing a slash is anchored to its base.

Unescaped trailing spaces are dropped. Leading whitespace is significant, so
`  #foo` is a pattern and not a comment — as in git.

## Two rules worth knowing

**A rule covers what it names *and* everything below it.** `*.bak` matches the
file `/a.bak`, and also `/a.bak/inner/file.txt`.

**A path below an excluded directory can never be re-included.** This is git's
behaviour, not an implementation limit — git does not descend into an excluded
directory, so no rule inside it is ever consulted:

```go
// Does not work: /build is excluded, so the negation is never reached.
"build/"
"!build/keep.o"

// Works: the directory itself stays included.
"build/*"
"!build/keep.o"
```

## Errors instead of panics

Compiling is explicit and never panics on user input:

```go
_, err := globbing.Compile("name[a-z")
// globbing: unterminated bracket expression at offset 4 in pattern "name[a-z"
```

- [`ErrNoPattern`](https://pkg.go.dev/github.com/sntns/globbing-go#ErrNoPattern)
  — the line is blank, a comment, or a bare `!`.
- [`*SyntaxError`](https://pkg.go.dev/github.com/sntns/globbing-go#SyntaxError)
  — the pattern is malformed; carries the offset.
- [`*ParseError`](https://pkg.go.dev/github.com/sntns/globbing-go#ParseError)
  — wraps the above with the line number, returned by `Filter.Parse`.

`MustCompile` is available for patterns fixed at build time.

## Paths

Paths are slash-separated and treated as absolute; a path without a leading
slash gets one. They are cleaned before matching, so `a//b`, `a/./b` and `/a/b`
are equivalent. On Windows, run paths through
[`filepath.ToSlash`](https://pkg.go.dev/path/filepath#ToSlash) first.

## Performance

Every pattern is compiled to a regular expression once, at `Compile` time.
Matching allocates nothing:

```
BenchmarkFilterMatch-4   1223446   979.4 ns/op   0 B/op   0 allocs/op
```

(33 patterns, a 6-segment path, on a Xeon @ 2.80 GHz.)

## Compatibility

The module requires Go 1.22 or later, and CI tests it against that floor plus
the two release lines the Go team supports, on Linux, macOS and Windows.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Matching changes need a case in
`conformance_test.go` so they are checked against git.

## License

[MIT](LICENSE)
