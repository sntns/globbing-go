package globbing

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

type filter struct {
	prefix  string
	pattern string
	regexp  *regexp.Regexp
}

func (f filter) isFile() bool {
	return strings.Contains(f.pattern, ".")
}

func (f filter) isFolder() bool {
	return strings.Contains(f.pattern, "/")
}

func (f filter) isAbsolute() bool {
	return strings.HasPrefix(f.pattern, "/")
}

func (f filter) isException() bool {
	return strings.HasPrefix(f.pattern, "!")
}

func (f filter) toRegExp() *regexp.Regexp {
	if f.regexp == nil {

		var escaped string
		if f.isException() {
			escaped, _ = strings.CutPrefix(f.pattern, "!")
		} else {
			escaped = f.pattern
		}
		escaped = strings.ReplaceAll(escaped, ".", "\\.")
		escaped = strings.ReplaceAll(escaped, "?", ".?")
		escaped = strings.ReplaceAll(escaped, "!", "^")
		escaped = strings.ReplaceAll(escaped, "/", "\\/")
		escaped = strings.ReplaceAll(escaped, "**\\/", "(*\\/)?")
		escaped = strings.ReplaceAll(escaped, "*", ".*")

		escaped, _ = strings.CutPrefix(escaped, "\\/")

		exp := escaped
		if strings.HasPrefix(f.pattern, "**/") {
			exp = fmt.Sprintf("^%s", escaped)
		} else {

			if f.isAbsolute() || (f.isFile() && f.isFolder()) {
				exp = fmt.Sprintf("^\\/%s", escaped)
			}
		}
		if f.isFile() {
			exp = fmt.Sprintf("%s$", exp)
		}
		f.regexp = regexp.MustCompile(exp)
	}

	return f.regexp
}

func (filter filter) match(path string) bool {
	if filter.prefix != "" {
		if !strings.HasPrefix(path, filter.prefix) {
			return false
		}
		path, _ = strings.CutPrefix(path, filter.prefix)
		if !strings.HasPrefix(path, "/") {
			path = fmt.Sprintf("/%s", path)
		}
	}
	return filter.toRegExp().MatchString(path)
}

type Filter struct {
	filters []filter
}

func NewFilter(filters ...filter) Filter {
	set := Filter{}
	set.filters = append(set.filters, filters...)
	return set
}

func (set *Filter) Parse(prefix string, reader io.Reader) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		set.filters = append(set.filters, filter{prefix: prefix, pattern: line})
	}
	return scanner.Err()
}

func (set *Filter) AddPattern(prefix string, pattern string) {
	set.filters = append(set.filters, filter{prefix: prefix, pattern: pattern})
}

func (set Filter) Match(path string) bool {
	match := false
	for _, filter := range set.filters {
		if filter.match(path) {
			if filter.isException() {
				return false
			} else {
				match = true
			}
		}
	}
	return match
}
