package globbing

import (
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"
)

// ErrNoPattern reports that a line carries no matchable rule: it is blank, a
// comment, or a bare negation such as "!".
var ErrNoPattern = errors.New("globbing: no pattern")

// A SyntaxError reports a malformed pattern.
type SyntaxError struct {
	// Pattern is the offending pattern, as supplied by the caller.
	Pattern string
	// Offset is the byte offset within Pattern at which the problem was found.
	Offset int
	// Reason describes what is wrong.
	Reason string
}

func (e *SyntaxError) Error() string { return "globbing: " + e.message() }

// message is Error without the package prefix, so that a wrapping error can
// supply its own without repeating it.
func (e *SyntaxError) message() string {
	return fmt.Sprintf("%s at offset %d in pattern %q", e.Reason, e.Offset, e.Pattern)
}

// A Pattern is a single compiled gitignore-style rule.
//
// The zero value is not usable. Build one with [Compile], [CompileBase] or
// [MustCompile]. A Pattern is safe for concurrent use once compiled.
type Pattern struct {
	raw     string
	base    string
	negated bool
	dirOnly bool

	file *regexp.Regexp
	dir  *regexp.Regexp
}

// Compile parses a gitignore-style pattern that applies from the filesystem
// root. It is shorthand for CompileBase("", pattern).
func Compile(pattern string) (*Pattern, error) {
	return CompileBase("", pattern)
}

// CompileBase parses a gitignore-style pattern that applies below base, the
// directory the rule was declared in — the equivalent of the directory holding
// a .gitignore file. An empty base means the filesystem root.
//
// It returns [ErrNoPattern] for blank lines, comments and bare negations, and a
// [*SyntaxError] for malformed patterns.
func CompileBase(base, pattern string) (*Pattern, error) {
	p := &Pattern{raw: pattern, base: normalizeBase(base)}

	body := trimTrailingSpaces(pattern)
	if body == "" || strings.HasPrefix(body, "#") {
		return nil, ErrNoPattern
	}
	if strings.HasPrefix(body, "!") {
		p.negated = true
		body = body[1:]
	}
	if endsWithUnescaped(body, '/') {
		p.dirOnly = true
		body = body[:len(body)-1]
	}
	if body == "" {
		return nil, ErrNoPattern
	}

	anchored := strings.Contains(body, "/")
	body = strings.TrimPrefix(body, "/")

	expr, err := translate(body)
	if err != nil {
		var se *SyntaxError
		if errors.As(err, &se) {
			se.Pattern = pattern
		}
		return nil, err
	}

	var head string
	if anchored {
		head = "^/" + expr
	} else {
		head = "^(?:/[^/]+)*/" + expr
	}

	// A directory rule also covers everything below the directory it names, so
	// the same expression serves both roles; only a file path has to sit
	// strictly below a directory-only rule.
	p.dir = regexp.MustCompile(head + `(?:/.*)?$`)
	if p.dirOnly {
		p.file = regexp.MustCompile(head + `/.*$`)
	} else {
		p.file = p.dir
	}
	return p, nil
}

// MustCompile is like [Compile] but panics if the pattern cannot be compiled.
// It is intended for patterns fixed at build time.
func MustCompile(pattern string) *Pattern {
	p, err := Compile(pattern)
	if err != nil {
		panic(err)
	}
	return p
}

// Match reports whether path, denoting a file, is selected by the pattern.
// Negation is not applied here: a negated pattern still reports a match, and
// [Pattern.Negated] tells the caller how to interpret it. Use [Filter] to
// evaluate a set of rules with their negations.
func (p *Pattern) Match(path string) bool {
	return p.match(path, false)
}

// MatchDir is like [Pattern.Match] but treats path as a directory, which lets
// directory-only rules such as "build/" match the directory itself and not only
// its contents.
func (p *Pattern) MatchDir(path string) bool {
	return p.match(path, true)
}

func (p *Pattern) match(name string, isDir bool) bool {
	rel, ok := p.relative(name)
	if !ok {
		return false
	}
	if isDir {
		return p.dir.MatchString(rel)
	}
	return p.file.MatchString(rel)
}

// relative strips the pattern's base from name, rejecting paths that lie
// outside it. The comparison is made on whole path segments, so a base of
// "/home/ann" does not swallow "/home/annex".
func (p *Pattern) relative(name string) (string, bool) {
	name = normalizePath(name)
	if p.base == "" {
		return name, true
	}
	if !strings.HasPrefix(name, p.base) {
		return "", false
	}
	rest := name[len(p.base):]
	if rest == "" || rest[0] != '/' {
		return "", false
	}
	return rest, true
}

// Negated reports whether the pattern re-includes what an earlier rule
// excluded, that is whether it was written with a leading "!".
func (p *Pattern) Negated() bool { return p.negated }

// DirOnly reports whether the pattern only applies to directories, that is
// whether it was written with a trailing "/".
func (p *Pattern) DirOnly() bool { return p.dirOnly }

// Base returns the directory the pattern is relative to, normalized. It is
// empty when the pattern applies from the filesystem root.
func (p *Pattern) Base() string { return p.base }

// String returns the pattern as it was supplied to Compile.
func (p *Pattern) String() string { return p.raw }

// translate converts a gitignore pattern body — negation, trailing slash and
// leading slash already removed — into a regular expression fragment.
func translate(body string) (string, error) {
	var b strings.Builder
	b.Grow(len(body) * 2)

	runes := []rune(body)
	segmentStart := true
	for i := 0; i < len(runes); {
		switch c := runes[i]; c {
		case '\\':
			if i+1 >= len(runes) {
				return "", &SyntaxError{Offset: i, Reason: "trailing backslash"}
			}
			b.WriteString(regexp.QuoteMeta(string(runes[i+1])))
			i += 2
			segmentStart = false

		case '/':
			b.WriteByte('/')
			i++
			segmentStart = true

		case '*':
			j := i
			for j < len(runes) && runes[j] == '*' {
				j++
			}
			wholeSegment := segmentStart && (j == len(runes) || runes[j] == '/')
			switch {
			case j-i >= 2 && wholeSegment && j == len(runes):
				b.WriteString(".*")
				i = j
			case j-i >= 2 && wholeSegment:
				b.WriteString("(?:[^/]+/)*")
				i = j + 1
				continue
			default:
				b.WriteString("[^/]*")
				i = j
			}
			segmentStart = false

		case '?':
			b.WriteString("[^/]")
			i++
			segmentStart = false

		case '[':
			class, next, err := translateClass(runes, i)
			if err != nil {
				return "", err
			}
			b.WriteString(class)
			i = next
			segmentStart = false

		default:
			b.WriteString(regexp.QuoteMeta(string(c)))
			i++
			segmentStart = false
		}
	}
	return b.String(), nil
}

// translateClass converts the bracket expression starting at runes[start] and
// returns the fragment along with the index just past the closing bracket.
func translateClass(runes []rune, start int) (string, int, error) {
	var b strings.Builder
	b.WriteByte('[')

	i := start + 1
	if i < len(runes) && (runes[i] == '!' || runes[i] == '^') {
		// A negated class must not be allowed to swallow the separator, or
		// "a[!b]c" would reach across directory boundaries.
		b.WriteString("^/")
		i++
	}
	if i < len(runes) && runes[i] == ']' {
		b.WriteString(`\]`)
		i++
	}

	for i < len(runes) {
		switch c := runes[i]; {
		case c == ']':
			b.WriteByte(']')
			return b.String(), i + 1, nil

		case c == '\\':
			if i+1 >= len(runes) {
				return "", 0, &SyntaxError{Offset: i, Reason: "trailing backslash"}
			}
			b.WriteString(quoteInClass(runes[i+1]))
			i += 2

		case c == '[' && i+1 < len(runes) && runes[i+1] == ':':
			end := indexRunes(runes, i+2, ":]")
			if end < 0 {
				return "", 0, &SyntaxError{Offset: i, Reason: "unterminated character class name"}
			}
			b.WriteString(string(runes[i : end+2]))
			i = end + 2

		case c == '-':
			b.WriteByte('-')
			i++

		default:
			b.WriteString(quoteInClass(c))
			i++
		}
	}
	return "", 0, &SyntaxError{Offset: start, Reason: "unterminated bracket expression"}
}

// quoteInClass renders c as a literal inside a regexp bracket expression.
// Backslash-escaping a letter or digit would turn it into a class shorthand
// such as \d, so those are emitted bare.
func quoteInClass(c rune) string {
	if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
		return string(c)
	}
	if c < 0x80 {
		return `\` + string(c)
	}
	return regexp.QuoteMeta(string(c))
}

func indexRunes(runes []rune, from int, sep string) int {
	target := []rune(sep)
	for i := from; i+len(target) <= len(runes); i++ {
		found := true
		for j, r := range target {
			if runes[i+j] != r {
				found = false
				break
			}
		}
		if found {
			return i
		}
	}
	return -1
}

// trimTrailingSpaces drops the trailing spaces git ignores, keeping any that
// were escaped with a backslash.
func trimTrailingSpaces(s string) string {
	i := len(s)
	for i > 0 && s[i-1] == ' ' && !isEscaped(s, i-1) {
		i--
	}
	return s[:i]
}

func endsWithUnescaped(s string, c byte) bool {
	return len(s) > 0 && s[len(s)-1] == c && !isEscaped(s, len(s)-1)
}

func isEscaped(s string, i int) bool {
	n := 0
	for j := i - 1; j >= 0 && s[j] == '\\'; j-- {
		n++
	}
	return n%2 == 1
}

// normalizePath puts a path in the slash-separated, rooted, cleaned form the
// matchers expect.
func normalizePath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	return path.Clean(p)
}

// normalizeBase is normalizePath for a base directory, mapping both the empty
// string and the root to "" so that relative can skip the prefix check.
func normalizeBase(b string) string {
	if b == "" {
		return ""
	}
	b = normalizePath(b)
	if b == "/" {
		return ""
	}
	return b
}
