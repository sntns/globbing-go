# Contributing

Thanks for taking the time. This is a small library with a narrow remit: match
paths the way `gitignore(5)` says to. That constraint decides most questions.

## Getting set up

```console
git clone https://github.com/sntns/globbing-go
cd globbing-go
make check
```

`make check` runs everything CI runs: tidiness, `gofmt`, `go vet`,
`golangci-lint` and the race-enabled test suite. You need Go 1.22 or later and
[golangci-lint](https://golangci-lint.run/welcome/install/) v2.5 or later.
`git` must be on `PATH` for the conformance suite; without it those tests skip
rather than fail.

Useful targets:

| Target | What it does |
| --- | --- |
| `make test` | tests with `-race -shuffle=on` |
| `make cover` | coverage profile plus a per-function summary |
| `make cover-html` | the same, opened in a browser |
| `make bench` | benchmarks with allocation counts |
| `make lint` | golangci-lint only |
| `make doc` | the package documentation, served locally |

## Changing matching behaviour

Git is the reference. If this library and `git check-ignore` disagree, this
library is wrong — including when the current behaviour looks more useful.

Every change to matching needs a case in
[`conformance_test.go`](conformance_test.go), which runs each rule set through
both implementations and compares them path by path. Adding a case is usually
two lines: a rule set in `conformanceCases` and, if needed, a path in
`conformancePaths`.

When you are unsure what git does, ask it:

```console
mkdir /tmp/t && cd /tmp/t && git init -q
printf 'build/\n!build/keep.o\n' > .gitignore
mkdir build && touch build/keep.o
git check-ignore -v --no-index build/keep.o
```

## Tests

Coverage sits at 100% of statements and CI reports it on every pull request. A
drop is not an automatic failure, but new code without a test will be sent back.

Exported API changes need a doc comment and, where it helps a reader, an
example in `example_test.go` — examples run as tests, so they cannot go stale.

## Commits and pull requests

Commit messages follow [Conventional Commits](https://www.conventionalcommits.org):
`fix: `, `feat: `, `docs: `, `test: `, `ci: `, `refactor: `, `build: `. Keep the
subject in the imperative and under 72 characters.

Pull requests should be focused. If you found two unrelated things, two pull
requests are easier to review and to revert.

## Releasing

Releases are git tags; there is nothing to publish. From `main`, with CI green:

1. Update `CHANGELOG.md`: move `Unreleased` entries under the new version and
   date it.
2. Tag and push:

   ```console
   git tag -a v0.2.0 -m 'v0.2.0'
   git push origin v0.2.0
   ```

3. Prime the module proxy so the version is immediately resolvable:

   ```console
   GOPROXY=proxy.golang.org go list -m github.com/sntns/globbing-go@v0.2.0
   ```

The project follows [semantic versioning](https://semver.org). While the major
version is 0, a minor bump may break the API; the changelog says so explicitly
when it does.

## Code of conduct

Participation is covered by the [Code of Conduct](CODE_OF_CONDUCT.md).
