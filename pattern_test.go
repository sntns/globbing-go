package globbing

import (
	"errors"
	"strings"
	"testing"
)

var patternTests = []struct {
	name     string
	base     string
	pattern  string
	match    []string
	notMatch []string
	dirs     []string
	notDirs  []string
}{
	{
		name:    "bare name matches an entry at any depth",
		pattern: "name",
		match: []string{
			"/name",
			"/name/file.txt",
			"/lib/name",
			"/lib/name/file.txt",
		},
		notMatch: []string{
			"/name.log",
			"/lib/name.log",
			"/surname",
		},
	},
	{
		name:    "bare name under a base",
		base:    "/home/sntns",
		pattern: "name",
		match: []string{
			"/home/sntns/name",
			"/home/sntns/name/file.txt",
			"/home/sntns/lib/name/file.txt",
		},
		notMatch: []string{
			"/name",
			"/name/file.txt",
			"/home/not_sntns/name",
			"/home/sntnsx/name",
			"/home/sntns",
		},
	},
	{
		name:    "trailing slash restricts to directories",
		pattern: "name/",
		match: []string{
			"/name/file.txt",
			"/name/log/name.log",
			"/lib/name/file.txt",
		},
		notMatch: []string{
			"/name.log",
			"/name",
		},
		dirs:    []string{"/name", "/lib/name"},
		notDirs: []string{"/name.log"},
	},
	{
		name:    "trailing slash under a base",
		base:    "/home/sntns",
		pattern: "name/",
		match: []string{
			"/home/sntns/name/file.txt",
			"/home/sntns/name/log/name.log",
		},
		notMatch: []string{
			"/home/sntns/name.log",
			"/name/file.txt",
			"/home/not_sntns/name/file.txt",
		},
	},
	{
		name:    "literal file name at any depth",
		pattern: "name.log",
		match: []string{
			"/name.log",
			"/lib/name.log",
		},
		notMatch: []string{
			"/name.log.txt",
			"/prefix-name.log",
			"/name/file.txt",
		},
	},
	{
		name:    "literal file name under a base",
		base:    "/home/sntns",
		pattern: "name.log",
		match: []string{
			"/home/sntns/name.log",
			"/home/sntns/lib/name.log",
		},
		notMatch: []string{
			"/home/sntns/name.log.txt",
			"/name.log",
			"/home/not_sntns/lib/name.log",
		},
	},
	{
		name:    "leading slash anchors to the base",
		pattern: "/name.file",
		match: []string{
			"/name.file",
		},
		notMatch: []string{
			"/lib/name.file",
		},
	},
	{
		name:    "leading slash anchors under a base",
		base:    "/home/sntns",
		pattern: "/name.file",
		match: []string{
			"/home/sntns/name.file",
		},
		notMatch: []string{
			"/home/sntns/lib/name.file",
			"/name.file",
			"/home/not_sntns/name.file",
		},
	},
	{
		name:    "an inner slash anchors too",
		pattern: "lib/name.file",
		match: []string{
			"/lib/name.file",
		},
		notMatch: []string{
			"/name.file",
			"/test/lib/name.file",
		},
	},
	{
		name:    "an inner slash without a dot still anchors",
		pattern: "lib/name",
		match: []string{
			"/lib/name",
			"/lib/name/file.txt",
		},
		notMatch: []string{
			"/test/lib/name",
			"/name",
		},
	},
	{
		name:    "inner slash under a base",
		base:    "/home/sntns",
		pattern: "lib/name.file",
		match: []string{
			"/home/sntns/lib/name.file",
		},
		notMatch: []string{
			"/home/sntns/test/lib/name.file",
			"/lib/name.file",
		},
	},
	{
		name:    "leading doublestar matches at any depth",
		pattern: "**/lib/name.file",
		match: []string{
			"/lib/name.file",
			"/test/lib/name.file",
			"/a/b/c/lib/name.file",
		},
		notMatch: []string{
			"/name.file",
			"/lib/sub/name.file",
		},
	},
	{
		name:    "leading doublestar under a base",
		base:    "/home/sntns",
		pattern: "**/lib/name.file",
		match: []string{
			"/home/sntns/lib/name.file",
			"/home/sntns/test/lib/name.file",
		},
		notMatch: []string{
			"/home/sntns/name.file",
			"/lib/name.file",
		},
	},
	{
		name:    "doublestar before a directory name",
		pattern: "**/name",
		match: []string{
			"/name/log.file",
			"/lib/name/log.file",
			"/name/lib/log.file",
		},
		notMatch: []string{
			"/name.log",
		},
	},
	{
		name:    "inner doublestar spans zero or more directories",
		pattern: "/lib/**/name",
		match: []string{
			"/lib/name/log.file",
			"/lib/test/name/log.file",
			"/lib/test/ver1/name/log.file",
		},
		notMatch: []string{
			"/name/log.file",
			"/other/lib/name/log.file",
		},
	},
	{
		name:    "trailing doublestar covers everything below",
		pattern: "/lib/**",
		match: []string{
			"/lib/name",
			"/lib/a/b/c",
		},
		notMatch: []string{
			"/lib",
			"/other/lib/name",
		},
	},
	{
		name:    "star does not cross a separator",
		pattern: "*.file",
		match: []string{
			"/name.file",
			"/lib/name.file",
		},
		notMatch: []string{
			"/lib/name.file.txt",
			"/lib/sub.file2/x",
		},
	},
	{
		name:    "a rule also covers everything below what it names",
		pattern: "*.bak",
		match: []string{
			"/a.bak",
			"/a.bak/inner/file.txt",
		},
	},
	{
		name:    "star under a base",
		base:    "/home/sntns",
		pattern: "*.file",
		match: []string{
			"/home/sntns/name.file",
			"/home/sntns/lib/name.file",
		},
		notMatch: []string{
			"/name.file",
			"/home/not_sntns/name.file",
		},
	},
	{
		name:    "star with a directory suffix",
		pattern: "*name/",
		match: []string{
			"/lastname/log.file",
			"/firstname/log.file",
			"/lib/lastname/log.file",
		},
		notMatch: []string{
			"/lastname.txt",
		},
	},
	{
		name:    "question mark matches exactly one character",
		pattern: "name?.file",
		match: []string{
			"/names.file",
			"/name1.file",
			"/name2.file",
		},
		notMatch: []string{
			"/name.file",
			"/names1.file",
			"/name/.file",
		},
	},
	{
		name:    "question mark under a base",
		base:    "/home/sntns",
		pattern: "name?.file",
		match: []string{
			"/home/sntns/names.file",
			"/home/sntns/name1.file",
		},
		notMatch: []string{
			"/home/sntns/name.file",
			"/names.file",
		},
	},
	{
		name:    "character range",
		pattern: "name[a-z].file",
		match: []string{
			"/names.file",
			"/nameb.file",
		},
		notMatch: []string{
			"/name1.file",
			"/name.file",
		},
	},
	{
		name:    "character set",
		pattern: "name[abc].file",
		match: []string{
			"/namea.file",
			"/nameb.file",
		},
		notMatch: []string{
			"/names.file",
		},
	},
	{
		name:    "negated character set",
		pattern: "name[!abc].file",
		match: []string{
			"/names.file",
			"/namex.file",
		},
		notMatch: []string{
			"/namea.file",
			"/namesb.file",
			"/name/.file",
		},
	},
	{
		name:    "caret negated character set",
		pattern: "name[^abc].file",
		match: []string{
			"/namex.file",
		},
		notMatch: []string{
			"/namea.file",
		},
	},
	{
		name:    "posix character class",
		pattern: "name[[:digit:]].file",
		match: []string{
			"/name1.file",
		},
		notMatch: []string{
			"/namea.file",
		},
	},
	{
		name:    "closing bracket as the first class member",
		pattern: "name[]a].file",
		match: []string{
			"/name].file",
			"/namea.file",
		},
		notMatch: []string{
			"/nameb.file",
		},
	},
	{
		name:    "escaped members inside a bracket expression",
		pattern: `name[a\]\-x].file`,
		match: []string{
			"/namea.file",
			"/name].file",
			"/name-.file",
			"/namex.file",
		},
		notMatch: []string{
			"/nameb.file",
		},
	},
	{
		name:    "non-ascii members inside a bracket expression",
		pattern: "caf[éè].txt",
		match: []string{
			"/café.txt",
			"/cafè.txt",
		},
		notMatch: []string{
			"/cafe.txt",
		},
	},
	{
		name:    "escaped letter inside a bracket expression stays literal",
		pattern: `name[\d].file`,
		match: []string{
			"/named.file",
		},
		notMatch: []string{
			"/name1.file",
		},
	},
	{
		name:    "a range inside a bracket expression",
		pattern: "name[0-9-].file",
		match: []string{
			"/name5.file",
			"/name-.file",
		},
		notMatch: []string{
			"/namea.file",
		},
	},
	{
		name:    "regexp metacharacters are literals",
		pattern: "v1.0+build(2).txt",
		match: []string{
			"/v1.0+build(2).txt",
		},
		notMatch: []string{
			"/v100build2.txt",
			"/v1x0+build(2).txt",
		},
	},
	{
		name:    "backslash escapes a wildcard",
		pattern: `name\*.file`,
		match: []string{
			"/name*.file",
		},
		notMatch: []string{
			"/names.file",
		},
	},
	{
		name:    "backslash escapes a leading bang",
		pattern: `\!important.txt`,
		match: []string{
			"/!important.txt",
		},
		notMatch: []string{
			"/important.txt",
		},
	},
	{
		name:    "backslash escapes a leading hash",
		pattern: `\#notacomment`,
		match: []string{
			"/#notacomment",
		},
	},
	{
		name:    "unescaped trailing spaces are dropped",
		pattern: "name.file   ",
		match: []string{
			"/name.file",
		},
	},
	{
		name:    "escaped trailing space is kept",
		pattern: `name.file\ `,
		match: []string{
			"/name.file ",
		},
		notMatch: []string{
			"/name.file",
		},
	},
	{
		name:    "negation still reports a match",
		pattern: "!name.log",
		match: []string{
			"/name.log",
		},
	},
	{
		name:    "base is compared on segment boundaries",
		base:    "/home/ann",
		pattern: "notes.txt",
		match: []string{
			"/home/ann/notes.txt",
		},
		notMatch: []string{
			"/home/annex/notes.txt",
			"/home/ann",
		},
	},
	{
		name:    "paths are cleaned before matching",
		pattern: "/lib/name.file",
		match: []string{
			"lib/name.file",
			"/lib//name.file",
			"/lib/./name.file",
			"/lib/sub/../name.file",
		},
	},
	{
		name:    "unicode is matched by rune",
		pattern: "caf?/",
		match: []string{
			"/café/menu.txt",
		},
		notMatch: []string{
			"/caf/menu.txt",
		},
	},
}

func TestPatternMatch(t *testing.T) {
	for _, tt := range patternTests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := CompileBase(tt.base, tt.pattern)
			if err != nil {
				t.Fatalf("CompileBase(%q, %q) = %v", tt.base, tt.pattern, err)
			}
			for _, name := range tt.match {
				if !p.Match(name) {
					t.Errorf("Match(%q) = false, want true (pattern %q, regexp %s)", name, tt.pattern, p.file)
				}
			}
			for _, name := range tt.notMatch {
				if p.Match(name) {
					t.Errorf("Match(%q) = true, want false (pattern %q, regexp %s)", name, tt.pattern, p.file)
				}
			}
			for _, name := range tt.dirs {
				if !p.MatchDir(name) {
					t.Errorf("MatchDir(%q) = false, want true (pattern %q, regexp %s)", name, tt.pattern, p.dir)
				}
			}
			for _, name := range tt.notDirs {
				if p.MatchDir(name) {
					t.Errorf("MatchDir(%q) = true, want false (pattern %q, regexp %s)", name, tt.pattern, p.dir)
				}
			}
		})
	}
}

func TestPatternMatchDirImpliesMatchOfContents(t *testing.T) {
	p := MustCompile("build/")
	if !p.MatchDir("/build") {
		t.Error("MatchDir(/build) = false, want true")
	}
	if p.Match("/build") {
		t.Error("Match(/build) = true, want false: a directory rule does not match a file of that name")
	}
	if !p.Match("/build/out.o") {
		t.Error("Match(/build/out.o) = false, want true")
	}
}

func TestCompileNoPattern(t *testing.T) {
	for _, pattern := range []string{"", "   ", "#", "# comment", "!", "/", "!/"} {
		t.Run(pattern, func(t *testing.T) {
			if _, err := Compile(pattern); !errors.Is(err, ErrNoPattern) {
				t.Errorf("Compile(%q) error = %v, want ErrNoPattern", pattern, err)
			}
		})
	}
}

func TestCompileSyntaxError(t *testing.T) {
	tests := []struct {
		pattern string
		reason  string
	}{
		{`name[a-z`, "unterminated bracket expression"},
		{`name\`, "trailing backslash"},
		{`name[a\`, "trailing backslash"},
		{`name[[:digit`, "unterminated character class name"},
	}
	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			_, err := Compile(tt.pattern)
			var se *SyntaxError
			if !errors.As(err, &se) {
				t.Fatalf("Compile(%q) error = %v, want *SyntaxError", tt.pattern, err)
			}
			if se.Reason != tt.reason {
				t.Errorf("Reason = %q, want %q", se.Reason, tt.reason)
			}
			if se.Pattern != tt.pattern {
				t.Errorf("Pattern = %q, want %q", se.Pattern, tt.pattern)
			}
			if !strings.Contains(se.Error(), tt.reason) {
				t.Errorf("Error() = %q, want it to mention %q", se.Error(), tt.reason)
			}
		})
	}
}

func TestMustCompilePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("MustCompile did not panic on an invalid pattern")
		}
	}()
	MustCompile("name[a-z")
}

func TestPatternAccessors(t *testing.T) {
	p, err := CompileBase("/home/ann/", "!build/")
	if err != nil {
		t.Fatalf("CompileBase: %v", err)
	}
	if got, want := p.String(), "!build/"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
	if got, want := p.Base(), "/home/ann"; got != want {
		t.Errorf("Base() = %q, want %q", got, want)
	}
	if !p.Negated() {
		t.Error("Negated() = false, want true")
	}
	if !p.DirOnly() {
		t.Error("DirOnly() = false, want true")
	}
}

func TestNormalizeBase(t *testing.T) {
	tests := map[string]string{
		"":            "",
		"/":           "",
		"home/ann":    "/home/ann",
		"/home/ann/":  "/home/ann",
		"/home//ann":  "/home/ann",
		"/home/./ann": "/home/ann",
	}
	for in, want := range tests {
		if got := normalizeBase(in); got != want {
			t.Errorf("normalizeBase(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizePath(t *testing.T) {
	tests := map[string]string{
		"":           "/",
		"/":          "/",
		"a/b":        "/a/b",
		"/a/b/":      "/a/b",
		"/a//b":      "/a/b",
		"/a/../b":    "/b",
		"/a/b/c/../": "/a/b",
	}
	for in, want := range tests {
		if got := normalizePath(in); got != want {
			t.Errorf("normalizePath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestEmptyBaseMatchesEverywhere(t *testing.T) {
	p := MustCompile("*.log")
	for _, base := range []string{"", "/"} {
		q, err := CompileBase(base, "*.log")
		if err != nil {
			t.Fatalf("CompileBase(%q): %v", base, err)
		}
		if q.Base() != p.Base() {
			t.Errorf("CompileBase(%q).Base() = %q, want %q", base, q.Base(), p.Base())
		}
	}
}
