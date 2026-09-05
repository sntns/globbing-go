package globbing

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
)

var filterTests = []struct {
	name     string
	patterns []struct{ base, pattern string }
	match    []string
	notMatch []string
}{
	{
		name: "several directory rules at the root",
		patterns: []struct{ base, pattern string }{
			{"", "name/"},
			{"", "sntns/"},
		},
		match: []string{
			"/name/file.txt",
			"/name/log/name.log",
			"/sntns/file.txt",
			"/sntns/log/name.log",
		},
		notMatch: []string{
			"/name.log",
			"/sntns.log",
		},
	},
	{
		name: "unanchored rules apply below their base",
		patterns: []struct{ base, pattern string }{
			{"", "name/"},
			{"", "sntns/"},
			{"/home", "sntns/"},
		},
		match: []string{
			"/name/file.txt",
			"/sntns/file.txt",
			"/home/sntns/file.txt",
			"/home/name/file.txt",
		},
		notMatch: []string{
			"/name.log",
			"/home/sntns.log",
		},
	},
	{
		name: "anchored rules stay at their base",
		patterns: []struct{ base, pattern string }{
			{"", "/name/"},
			{"", "/sntns/"},
			{"/home", "/sntns/"},
		},
		match: []string{
			"/name/file.txt",
			"/sntns/log/name.log",
			"/home/sntns/file.txt",
		},
		notMatch: []string{
			"/home/name/file.txt",
			"/name.log",
			"/home/sntns.log",
		},
	},
	{
		name: "a negation re-includes what an earlier rule excluded",
		patterns: []struct{ base, pattern string }{
			{"", "*.log"},
			{"/home", "!debug.log"},
		},
		match: []string{
			"/app.log",
			"/other/debug.log",
		},
		notMatch: []string{
			"/home/debug.log",
			"/home/sub/debug.log",
		},
	},
	{
		name: "a rule below an excluded directory has no effect",
		patterns: []struct{ base, pattern string }{
			{"/home", "/sntns/"},
			{"/home", "!/sntns/log/"},
		},
		match: []string{
			"/home/sntns/file.txt",
			"/home/sntns/log/name.log",
		},
	},
	{
		name: "re-including below a directory needs the directory itself kept",
		patterns: []struct{ base, pattern string }{
			{"", "/*"},
			{"", "!/foo"},
			{"", "/foo/*"},
			{"", "!/foo/bar"},
		},
		match: []string{
			"/other",
			"/foo/baz",
		},
		notMatch: []string{
			"/foo/bar",
			"/foo/bar/deep.txt",
		},
	},
	{
		name: "the last matching rule wins",
		patterns: []struct{ base, pattern string }{
			{"", "*.log"},
			{"", "!debug.log"},
			{"", "/vendor/debug.log"},
		},
		match: []string{
			"/app.log",
			"/vendor/debug.log",
		},
		notMatch: []string{
			"/debug.log",
			"/sub/debug.log",
		},
	},
	{
		name: "an earlier negation does not survive a later exclusion",
		patterns: []struct{ base, pattern string }{
			{"", "!keep.txt"},
			{"", "*.txt"},
		},
		match: []string{
			"/keep.txt",
		},
	},
	{
		name:     "an empty set matches nothing",
		patterns: nil,
		notMatch: []string{"/anything", "/a/b/c"},
	},
}

func TestFilterMatch(t *testing.T) {
	for _, tt := range filterTests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFilter()
			for _, p := range tt.patterns {
				if err := f.AddPattern(p.base, p.pattern); err != nil {
					t.Fatalf("AddPattern(%q, %q) = %v", p.base, p.pattern, err)
				}
			}
			for _, name := range tt.match {
				if !f.Match(name) {
					t.Errorf("Match(%q) = false, want true", name)
				}
			}
			for _, name := range tt.notMatch {
				if f.Match(name) {
					t.Errorf("Match(%q) = true, want false", name)
				}
			}
		})
	}
}

func TestNewFilterKeepsOrder(t *testing.T) {
	f := NewFilter(MustCompile("*.log"), MustCompile("!keep.log"))
	if got, want := f.Len(), 2; got != want {
		t.Fatalf("Len() = %d, want %d", got, want)
	}
	if got, want := f.Patterns()[1].String(), "!keep.log"; got != want {
		t.Errorf("Patterns()[1] = %q, want %q", got, want)
	}
	if f.Match("/keep.log") {
		t.Error("Match(/keep.log) = true, want false")
	}
	if !f.Match("/other.log") {
		t.Error("Match(/other.log) = false, want true")
	}
}

func TestFilterPatternsIsACopy(t *testing.T) {
	f := NewFilter(MustCompile("*.log"))
	got := f.Patterns()
	got[0] = MustCompile("*.tmp")
	if f.Patterns()[0].String() != "*.log" {
		t.Error("mutating the slice returned by Patterns() changed the filter")
	}
}

func TestFilterMatchDir(t *testing.T) {
	f := NewFilter()
	if err := f.AddPattern("", "build/"); err != nil {
		t.Fatalf("AddPattern: %v", err)
	}
	if !f.MatchDir("/build") {
		t.Error("MatchDir(/build) = false, want true")
	}
	if f.Match("/build") {
		t.Error("Match(/build) = true, want false")
	}
}

func TestFilterAddPatternError(t *testing.T) {
	f := NewFilter()
	if err := f.AddPattern("", "name[a-z"); err == nil {
		t.Fatal("AddPattern with an invalid pattern returned nil error")
	}
	if err := f.AddPattern("", "# comment"); !errors.Is(err, ErrNoPattern) {
		t.Fatalf("AddPattern with a comment = %v, want ErrNoPattern", err)
	}
	if f.Len() != 0 {
		t.Errorf("Len() = %d, want 0: a rejected pattern must not be stored", f.Len())
	}
}

const gitignoreSample = `
# build output
build/
*.o

!keep.o

# an escaped bang is a literal name
\!important
`

func TestFilterParse(t *testing.T) {
	f := NewFilter()
	if err := f.Parse("/project", strings.NewReader(gitignoreSample)); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got, want := f.Len(), 4; got != want {
		t.Fatalf("Len() = %d, want %d (%v)", got, want, f.Patterns())
	}

	for _, name := range []string{"/project/build/out", "/project/main.o", "/project/!important"} {
		if !f.Match(name) {
			t.Errorf("Match(%q) = false, want true", name)
		}
	}
	for _, name := range []string{"/project/keep.o", "/project/main.c", "/other/main.o"} {
		if f.Match(name) {
			t.Errorf("Match(%q) = true, want false", name)
		}
	}
}

func TestFilterParseIndentedHashIsNotAComment(t *testing.T) {
	f := NewFilter()
	if err := f.Parse("", strings.NewReader("  # not a comment\n")); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got, want := f.Len(), 1; got != want {
		t.Fatalf("Len() = %d, want %d: git only treats a leading # as a comment", got, want)
	}
	if !f.Match("/  # not a comment") {
		t.Error("the indented line was not kept as a literal pattern")
	}
}

func TestFilterParseCRLF(t *testing.T) {
	f := NewFilter()
	if err := f.Parse("", strings.NewReader("*.log\r\n!keep.log\r\n")); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !f.Match("/app.log") {
		t.Error("Match(/app.log) = false, want true: the CR was not stripped")
	}
	if f.Match("/keep.log") {
		t.Error("Match(/keep.log) = true, want false")
	}
}

func TestFilterParseError(t *testing.T) {
	f := NewFilter()
	err := f.Parse("", strings.NewReader("*.log\n\nname[a-z\n*.tmp\n"))

	var pe *ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("Parse error = %v, want *ParseError", err)
	}
	if got, want := pe.Line, 3; got != want {
		t.Errorf("Line = %d, want %d", got, want)
	}
	if got, want := pe.Pattern, "name[a-z"; got != want {
		t.Errorf("Pattern = %q, want %q", got, want)
	}
	var se *SyntaxError
	if !errors.As(err, &se) {
		t.Errorf("Parse error does not unwrap to *SyntaxError")
	}
	want := `globbing: line 3: unterminated bracket expression at offset 4 in pattern "name[a-z"`
	if got := err.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
	if got, want := f.Len(), 1; got != want {
		t.Errorf("Len() = %d, want %d: rules read before the error are kept", got, want)
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func TestFilterParseReadError(t *testing.T) {
	f := NewFilter()
	if err := f.Parse("", failingReader{}); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Errorf("Parse error = %v, want io.ErrUnexpectedEOF", err)
	}
}

func TestParseErrorMessage(t *testing.T) {
	pe := &ParseError{Line: 7, Pattern: "x[", Err: errors.New("boom")}
	if got, want := pe.Error(), "globbing: line 7: boom"; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
	if !errors.Is(pe, pe.Err) {
		t.Error("ParseError does not unwrap to its cause")
	}
}

func BenchmarkFilterMatch(b *testing.B) {
	f := NewFilter()
	for i := range 32 {
		if err := f.AddPattern("", fmt.Sprintf("pkg%d/**/*.tmp", i)); err != nil {
			b.Fatal(err)
		}
	}
	if err := f.AddPattern("", "**/vendor/*.log"); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Match("/pkg17/a/b/c/vendor/output.log")
	}
}
