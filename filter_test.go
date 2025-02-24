package globbing

import (
	"fmt"
	"testing"
)

var testPattern = []struct {
	test     string
	prefix   string
	pattern  string
	match    []string
	notMatch []string
}{
	{
		test:    "no prefix",
		pattern: "name",
		match: []string{
			"/name.log",
			"/name/file.txt",
			"/lib/name.log",
		},
	},
	{
		test:    "prefix",
		prefix:  "/home/sntns",
		pattern: "name",
		match: []string{
			"/home/sntns/name.log",
			"/home/sntns/name/file.txt",
			"/home/sntns/lib/name.log",
		},
		notMatch: []string{
			"/name.log",
			"/name/file.txt",
			"/lib/name.log",
			"/home/not_sntns/name.log",
			"/home/not_sntns/name/file.txt",
			"/home/not_sntns/lib/name.log",
		},
	},
	{
		test:    "no prefix",
		pattern: "name/",
		match: []string{
			"/name/file.txt",
			"/name/log/name.log",
		},
		notMatch: []string{
			"/name.log",
		},
	},
	{
		test:    "prefix",
		prefix:  "/home/sntns",
		pattern: "name/",
		match: []string{
			"/home/sntns/name/file.txt",
			"/home/sntns/name/log/name.log",
		},
		notMatch: []string{
			"/home/sntns/name.log",
			"/name/file.txt",
			"/name/log/name.log",
			"/home/not_sntns/name/file.txt",
			"/home/not_sntns/name/log/name.log",
		},
	},
	{
		test:    "no prefix",
		pattern: "name.log",
		match: []string{
			"/name.log",
			"/lib/name.log",
		},
		notMatch: []string{
			"/name.log.txt",
			"/name/file.txt",
		},
	},
	{
		test:    "prefix",
		prefix:  "/home/sntns",
		pattern: "name.log",
		match: []string{
			"/home/sntns/name.log",
			"/home/sntns/lib/name.log",
		},
		notMatch: []string{
			"/home/sntns/name.log.txt",
			"/home/sntns/name/file.txt",
			"/name.log",
			"/lib/name.log",
			"/home/not_sntns/name.log",
			"/home/not_sntns/lib/name.log",
		},
	},
	{
		test:    "no prefix",
		pattern: "/name.file",
		match: []string{
			"/name.file",
		},
		notMatch: []string{
			"/lib/name.file",
		},
	},
	{
		test:    "prefix",
		prefix:  "/home/sntns",
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
		test:    "no prefix",
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
		test:    "prefix",
		prefix:  "/home/sntns",
		pattern: "lib/name.file",
		match: []string{
			"/home/sntns/lib/name.file",
		},
		notMatch: []string{
			"/home/sntns/name.file",
			"/home/sntns/test/lib/name.file",
			"/lib/name.file",
			"/home/not_sntns/lib/name.file",
		},
	},
	{
		test:    "no prefix",
		pattern: "**/lib/name.file",
		match: []string{
			"/lib/name.file",
			"/test/lib/name.file",
		},
		notMatch: []string{
			"/name.file",
		},
	},
	{
		test:    "prefix",
		prefix:  "/home/sntns",
		pattern: "**/lib/name.file",
		match: []string{
			"/home/sntns/lib/name.file",
			"/home/sntns/test/lib/name.file",
		},
		notMatch: []string{
			"/home/sntns/name.file",
			"/name.file",
			"/lib/name.file",
			"/test/lib/name.file",
			"/home/not_sntns/lib/name.file",
			"/home/not_sntns/test/lib/name.file",
		},
	},
	{
		test:    "no prefix",
		pattern: "**/name",
		match: []string{
			"/name/log.file",
			"/lib/name/log.file",
			"/name/lib/log.file",
		},
	},
	{
		test:    "prefix",
		prefix:  "/home/sntns",
		pattern: "**/name",
		match: []string{
			"/home/sntns/name/log.file",
			"/home/sntns/lib/name/log.file",
			"/home/sntns/name/lib/log.file",
		},
		notMatch: []string{
			"/name/log.file",
			"/lib/name/log.file",
			"/name/lib/log.file",
			"/home/not_sntns/name/log.file",
			"/home/not_sntns/lib/name/log.file",
			"/home/not_sntns/name/lib/log.file",
		},
	},
	{
		test:    "no prefix",
		pattern: "/lib/**/name",
		match: []string{
			"/lib/name/log.file",
			"/lib/test/name/log.file",
			"/lib/test/ver1/name/log.file",
		},
		notMatch: []string{
			"/name/log.file",
		},
	},
	{
		test:    "prefix",
		prefix:  "/home/sntns",
		pattern: "/lib/**/name",
		match: []string{
			"/home/sntns/lib/name/log.file",
			"/home/sntns/lib/test/name/log.file",
			"/home/sntns/lib/test/ver1/name/log.file",
		},
		notMatch: []string{
			"/home/sntns/name/log.file",
			"/lib/name/log.file",
			"/lib/test/name/log.file",
			"/lib/test/ver1/name/log.file",
			"/home/not_sntns/lib/name/log.file",
			"/home/not_sntns/lib/test/name/log.file",
			"/home/not_sntns/lib/test/ver1/name/log.file",
		},
	},
	{
		test:    "no prefix",
		pattern: "*.file",
		match: []string{
			"/name.file",
			"/lib/name.file",
		},
	},
	{
		test:    "prefix",
		prefix:  "/home/sntns",
		pattern: "*.file",
		match: []string{
			"/home/sntns/name.file",
			"/home/sntns/lib/name.file",
		},
		notMatch: []string{
			"/name.file",
			"/lib/name.file",
			"/home/not_sntns/name.file",
			"/home/not_sntns/lib/name.file",
		},
	},
	{
		test:    "no prefix",
		pattern: "*name/",
		match: []string{
			"/lastname/log.file",
			"/firstname/log.file",
		},
	},
	{
		test:    "prefix",
		prefix:  "/home/sntns",
		pattern: "*name/",
		match: []string{
			"/home/sntns/lastname/log.file",
			"/home/sntns/firstname/log.file",
		},
		notMatch: []string{
			"/lastname/log.file",
			"/firstname/log.file",
			"/home/not_sntns/lastname/log.file",
			"/home/not_sntns/firstname/log.file",
		},
	},
	{
		test:    "no prefix",
		pattern: "name?.file",
		match: []string{
			"/name.file",
			"/names.file",
			"/name1.file",
			"/name2.file",
		},
		notMatch: []string{
			"/names1.file",
		},
	},
	{
		test:    "prefix",
		prefix:  "/home/sntns",
		pattern: "name?.file",
		match: []string{
			"/home/sntns/name.file",
			"/home/sntns/names.file",
			"/home/sntns/name1.file",
			"/home/sntns/name2.file",
		},
		notMatch: []string{
			"/home/sntns/names1.file",
			"/name.file",
			"/names.file",
			"/name1.file",
			"/name2.file",
			"/home/not_sntns/name.file",
			"/home/not_sntns/names.file",
			"/home/not_sntns/name1.file",
			"/home/not_sntns/name2.file",
		},
	},
	{
		test:    "no prefix",
		pattern: "name[a-z].file",
		match: []string{
			"/names.file",
			"/nameb.file",
		},
		notMatch: []string{
			"/name1.file",
		},
	},
	{
		test:    "prefix",
		prefix:  "/home/sntns",
		pattern: "name[a-z].file",
		match: []string{
			"/home/sntns/names.file",
			"/home/sntns/nameb.file",
		},
		notMatch: []string{
			"/home/sntns/name1.file",
			"/names.file",
			"/nameb.file",
			"/home/not_sntns/names.file",
			"/home/not_sntns/nameb.file",
		},
	},
	{
		test:    "no prefix",
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
		test:    "prefix",
		prefix:  "/home/sntns",
		pattern: "name[abc].file",
		match: []string{
			"/home/sntns/namea.file",
			"/home/sntns/nameb.file",
		},
		notMatch: []string{
			"/home/sntns/names.file",
			"/namea.file",
			"/nameb.file",
			"/home/not_sntns/namea.file",
			"/home/not_sntns/nameb.file",
		},
	},
	{
		test:    "no prefix",
		pattern: "name[!abc].file",
		match: []string{
			"/names.file",
			"/namex.file",
		},
		notMatch: []string{
			"/namesb.file",
		},
	},
	{
		test:    "prefix",
		prefix:  "/home/sntns",
		pattern: "name[!abc].file",
		match: []string{
			"/home/sntns/names.file",
			"/home/sntns/namex.file",
		},
		notMatch: []string{
			"/home/sntns/namesb.file",
			"/names.file",
			"/namex.file",
			"/home/not_sntns/names.file",
			"/home/not_sntns/namex.file",
		},
	},
	{
		test:    "no prefix",
		pattern: "*.file",
		match: []string{
			"/name.file",
			"/lib/name.file",
		},
	},
	{
		test:    "prefix",
		prefix:  "/home/sntns",
		pattern: "*.file",
		match: []string{
			"/home/sntns/name.file",
			"/home/sntns/lib/name.file",
		},
		notMatch: []string{
			"/name.file",
			"/lib/name.file",
			"/home/not_sntns/name.file",
			"/home/not_sntns/lib/name.file",
		},
	},
}

func TestPattern(t *testing.T) {
	for _, tt := range testPattern {
		var test string
		if tt.prefix == "" {
			test = fmt.Sprintf("no prefix(%s)", tt.pattern)
		} else {
			test = fmt.Sprintf("prefix(%s)", tt.pattern)
		}
		t.Run(test, func(t *testing.T) {
			filter := filter{prefix: tt.prefix, pattern: tt.pattern}
			for _, name := range tt.match {
				if got := filter.match(name); !got {
					t.Errorf("match(%s, %s) = %v, want %v using %s", tt.pattern, name, got, true, filter.regexp)
				}
			}
			for _, name := range tt.notMatch {
				if got := filter.match(name); got {
					t.Errorf("match(%s, %s) = %v, want %v using %s", tt.pattern, name, got, false, filter.regexp)
				}
			}
		})
	}
}

var testFilter = []struct {
	test     string
	patterns map[string][]string
	match    []string
	notMatch []string
}{
	{
		test: "multi patterns, no prefix",
		patterns: map[string][]string{
			"": {
				"name/",
				"sntns/",
			},
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
		test: "multi patterns, prefix",
		patterns: map[string][]string{
			"": {
				"name/",
				"sntns/",
			},
			"/home": {
				"sntns/",
			},
		},
		match: []string{
			"/name/file.txt",
			"/name/log/name.log",
			"/sntns/file.txt",
			"/sntns/log/name.log",
			"/home/sntns/file.txt",
			"/home/sntns/log/name.log",
			"/home/name/file.txt",
			"/home/name/log/name.log",
		},
		notMatch: []string{
			"/name.log",
			"/sntns.log",
			"/home/sntns.log",
		},
	},
	{
		test: "multi patterns, prefix",
		patterns: map[string][]string{
			"": {
				"/name/",
				"/sntns/",
			},
			"/home": {
				"/sntns/",
			},
		},
		match: []string{
			"/name/file.txt",
			"/name/log/name.log",
			"/sntns/file.txt",
			"/sntns/log/name.log",
			"/home/sntns/file.txt",
			"/home/sntns/log/name.log",
		},
		notMatch: []string{
			"/home/name/file.txt",
			"/home/name/log/name.log",
			"/name.log",
			"/sntns.log",
			"/home/sntns.log",
		},
	},
	{
		test: "multi patterns, prefix, exception",
		patterns: map[string][]string{
			"": {
				"/name/",
				"/sntns/",
			},
			"/home": {
				"/sntns/",
				"!/sntns/log/",
			},
		},
		match: []string{
			"/name/file.txt",
			"/name/log/name.log",
			"/sntns/file.txt",
			"/sntns/log/name.log",
			"/home/sntns/file.txt",
		},
		notMatch: []string{
			"/home/name/file.txt",
			"/home/name/log/name.log",
			"/name.log",
			"/sntns.log",
			"/home/sntns.log",
			"/home/sntns/log/name.log",
		},
	},
}

func TestFilter(t *testing.T) {
	for _, tt := range testFilter {
		t.Run(tt.test, func(t *testing.T) {
			filter := NewFilter()
			for prefix, v := range tt.patterns {
				for _, pattern := range v {
					filter.AddPattern(prefix, pattern)
				}

			}
			for _, name := range tt.match {
				if got := filter.Match(name); !got {
					t.Errorf("Match(%s) = %v, want %v", name, got, true)
				}
			}
			for _, name := range tt.notMatch {
				if got := filter.Match(name); got {
					t.Errorf("Match( %s) = %v, want %v", name, got, false)
				}
			}
		})
	}
}
