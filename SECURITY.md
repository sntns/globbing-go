# Security policy

## Supported versions

Only the latest released minor version receives fixes.

| Version | Supported |
| --- | --- |
| 0.2.x | yes |
| 0.1.x | no |

## Reporting a vulnerability

Report privately through
[GitHub's advisory form](https://github.com/sntns/globbing-go/security/advisories/new).
Please do not open a public issue for a vulnerability.

Include the pattern and path that trigger the problem, and what you observed. We
aim to acknowledge within five working days and to agree a disclosure timeline
with you from there.

## What counts

This library parses untrusted text — a `.gitignore` file is often user-supplied
— and compiles it to a regular expression. The following are in scope:

- **A panic on any input.** `Compile` and `Filter.Parse` must return an error,
  never panic, whatever the input. `MustCompile` panicking on a bad pattern is
  by design and is not a vulnerability.
- **Catastrophic backtracking.** Go's `regexp` runs in time linear in the input
  and does not backtrack, so this should not be reachable; a demonstration that
  it is would be a real finding.
- **A pattern escaping its base.** A rule declared for `/home/ann` must never
  match outside it, whatever the pattern or the path — including through `..`
  segments, which are cleaned before matching.
- **A pattern matching more than it should**, where a caller relying on the
  filter to exclude a path would leak it.

Out of scope: divergences from git that have no security consequence. Those are
ordinary bugs — please open an issue.

## Design notes

The library has no dependencies outside the standard library and performs no
I/O: it reads from an `io.Reader` you hand it and matches strings. It does not
touch the filesystem, so it cannot follow a symlink or read a file you did not
give it.
