// Package globbing matches filesystem paths against gitignore-style patterns.
//
// The syntax and the evaluation rules follow gitignore(5). A conformance suite
// checks the package against git itself, so a divergence shows up as a test
// failure rather than as a surprise.
//
// # Pattern syntax
//
//	#          a leading hash introduces a comment
//	!          a leading bang negates the rule, re-including a path
//	/          a leading or inner slash anchors the rule to its base directory
//	/          a trailing slash restricts the rule to directories
//	*          matches any run of characters, never crossing a separator
//	?          matches exactly one character, never a separator
//	[abc]      matches one of the listed characters
//	[a-z]      matches one character in the range
//	[!abc]     matches one character outside the set, never a separator
//	[[:digit:]] matches one character of the POSIX class
//	**/x       matches x at any depth
//	x/**       matches everything below x
//	a/**/b     matches b below a, at any depth
//	\          escapes the next character, including #, ! and a trailing space
//
// A pattern that contains no slash — other than a trailing one — matches an
// entry of that name at any depth. A pattern that contains a slash is anchored
// to the directory it was declared in.
//
// Unescaped trailing spaces are dropped. Leading whitespace is significant, so
// a line starting with spaces is a pattern and not a comment, as in git.
//
// # Matching a single rule
//
//	p, err := globbing.Compile("*.log")
//	if err != nil {
//		return err
//	}
//	p.Match("/var/app.log") // true
//
// # Matching a set of rules
//
// A [Filter] evaluates rules the way git does: the last rule that matches a
// path decides, and a path below an excluded directory can never be
// re-included.
//
//	f := globbing.NewFilter()
//	if err := f.Parse("", strings.NewReader("*.log\n!keep.log\n")); err != nil {
//		return err
//	}
//	f.Match("/app.log")  // true
//	f.Match("/keep.log") // false
//
// Rules read from a .gitignore file in a subdirectory apply only below that
// directory. Pass the directory as the base:
//
//	f.Parse("/project/vendor", vendorGitignore)
//
// # Paths
//
// Paths are slash-separated and interpreted as absolute; a path without a
// leading slash gets one. They are cleaned before matching, so "a//b",
// "a/./b" and "/a/b" are equivalent. On Windows, convert paths with
// [path/filepath.ToSlash] before passing them in.
//
// Whether a path denotes a directory cannot be inferred from the path itself,
// so callers say so explicitly: [Filter.Match] treats the path as a file and
// [Filter.MatchDir] treats it as a directory. The distinction only matters for
// directory-only rules such as "build/", which match the directory itself as
// well as everything below it.
package globbing
