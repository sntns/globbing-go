package globbing

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

// A ParseError reports a pattern that [Filter.Parse] could not compile,
// together with the line it came from.
type ParseError struct {
	// Line is the 1-based line number within the parsed stream.
	Line int
	// Pattern is the offending line, verbatim.
	Pattern string
	// Err is the underlying error, typically a [*SyntaxError].
	Err error
}

func (e *ParseError) Error() string {
	var se *SyntaxError
	if errors.As(e.Err, &se) {
		return fmt.Sprintf("globbing: line %d: %s", e.Line, se.message())
	}
	return fmt.Sprintf("globbing: line %d: %v", e.Line, e.Err)
}

func (e *ParseError) Unwrap() error { return e.Err }

// A Filter is an ordered set of patterns evaluated together, the way git
// evaluates the rules collected from a set of .gitignore files.
//
// The last pattern that matches a path decides the outcome, so a negated rule
// re-includes a path only if no later rule excludes it again. As in git, a path
// below an excluded directory can never be re-included: the directory rule
// wins, whatever the rules written for its contents.
//
// A Filter is safe for concurrent matching once its patterns are in place;
// adding patterns is not concurrency-safe.
type Filter struct {
	patterns  []*Pattern
	negations int
}

// NewFilter returns a Filter holding the given patterns, in order.
func NewFilter(patterns ...*Pattern) *Filter {
	f := &Filter{}
	f.Add(patterns...)
	return f
}

// Add appends already-compiled patterns to the set.
func (f *Filter) Add(patterns ...*Pattern) {
	for _, p := range patterns {
		if p.negated {
			f.negations++
		}
	}
	f.patterns = append(f.patterns, patterns...)
}

// AddPattern compiles a single pattern relative to base and appends it. It
// returns [ErrNoPattern] if the line carries no rule, leaving the set
// unchanged.
func (f *Filter) AddPattern(base, pattern string) error {
	p, err := CompileBase(base, pattern)
	if err != nil {
		return err
	}
	f.Add(p)
	return nil
}

// Parse reads .gitignore-style lines from r and appends every rule it finds,
// each relative to base. Blank lines and comments are skipped. Parsing stops at
// the first malformed pattern and returns a [*ParseError]; rules read before it
// stay in the set.
func (f *Filter) Parse(base string, r io.Reader) error {
	scanner := bufio.NewScanner(r)
	for line := 1; scanner.Scan(); line++ {
		text := scanner.Text()
		p, err := CompileBase(base, text)
		switch {
		case err == nil:
			f.Add(p)
		case errors.Is(err, ErrNoPattern):
			continue
		default:
			return &ParseError{Line: line, Pattern: text, Err: err}
		}
	}
	return scanner.Err()
}

// Match reports whether path, denoting a file, is excluded by the set.
func (f *Filter) Match(path string) bool {
	return f.match(path, false)
}

// MatchDir is like [Filter.Match] but treats path as a directory, which lets
// directory-only rules such as "build/" match the directory itself.
func (f *Filter) MatchDir(path string) bool {
	return f.match(path, true)
}

func (f *Filter) match(name string, isDir bool) bool {
	name = normalizePath(name)

	// Git never descends into an excluded directory, so a rule can never
	// re-include something below one. Without negations the walk is redundant:
	// a pattern that excludes a directory already covers everything under it.
	if f.negations > 0 {
		for i := 1; i < len(name); i++ {
			if name[i] == '/' && f.decide(name[:i], true) {
				return true
			}
		}
	}
	return f.decide(name, isDir)
}

// decide applies the rules to a single path, the last match winning.
func (f *Filter) decide(name string, isDir bool) bool {
	for i := len(f.patterns) - 1; i >= 0; i-- {
		if p := f.patterns[i]; p.match(name, isDir) {
			return !p.negated
		}
	}
	return false
}

// Patterns returns the patterns in the set, in evaluation order. The returned
// slice is a copy; the patterns themselves are shared.
func (f *Filter) Patterns() []*Pattern {
	out := make([]*Pattern, len(f.patterns))
	copy(out, f.patterns)
	return out
}

// Len returns the number of patterns in the set.
func (f *Filter) Len() int { return len(f.patterns) }
