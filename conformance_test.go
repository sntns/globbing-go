package globbing_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	globbing "github.com/sntns/globbing-go"
)

// The cases below are checked twice: once against this package and once against
// git itself, so that any divergence from gitignore(5) shows up as a test
// failure rather than as a surprise for a user.
var conformanceCases = []struct {
	name  string
	rules []string
}{
	{"bare name", []string{"name"}},
	{"directory only", []string{"name/"}},
	{"literal file", []string{"name.log"}},
	{"anchored file", []string{"/name.file"}},
	{"anchored by inner slash", []string{"lib/name.file"}},
	{"anchored directory by inner slash", []string{"lib/name"}},
	{"leading doublestar", []string{"**/lib/name.file"}},
	{"leading doublestar directory", []string{"**/name"}},
	{"inner doublestar", []string{"/lib/**/name"}},
	{"trailing doublestar", []string{"/lib/**"}},
	{"surrounding doublestars", []string{"**/foo/**"}},
	{"star", []string{"*.file"}},
	{"star before a directory", []string{"*name/"}},
	{"question mark", []string{"name?.file"}},
	{"character range", []string{"name[a-z].file"}},
	{"character set", []string{"name[abc].file"}},
	{"negated character set", []string{"name[!abc].file"}},
	{"leading character range", []string{"[a-c]ame"}},
	{"consecutive stars inside a segment", []string{"a**b"}},
	{"everything", []string{"**"}},
	{"extension below a directory", []string{"doc/**/*.md"}},
	{"object files", []string{"*.o"}},
	{"single directory", []string{"build"}},
	{"star inside an anchored directory", []string{"src/*.c"}},
	{"escaped star", []string{`name\*.file`}},
	{"trailing spaces", []string{"name.log   "}},
	{"comment and blank lines", []string{"", "# comment", "*.o"}},
	{"negation", []string{"*.o", "!keep.o"}},
	{"negation overridden by a later rule", []string{"*.o", "!keep.o", "/lib/keep.o"}},
	{"negation cannot re-include below an excluded directory", []string{"build/", "!build/keep.o"}},
	{"several rules", []string{"*.o", "build/", "doc/**/*.md", "!doc/api/index.md"}},
	{"re-inclusion keeping every parent", []string{"/*", "!/doc", "/doc/*", "!/doc/api"}},
	{"negated then re-excluded", []string{"!keep.o", "*.o"}},
}

var conformancePaths = []string{
	"name", "name.log", "name.file", "names.file", "name1.file", "namea.file",
	"namex.file", "names1.file", "name*.file", "bame", "came", "dame", "nome",
	"ab", "aXb", "aXYb", "keep.o", "main.o", "build", "src/main.c", "src/main.o",
	"src/sub/main.c", "lib/name", "lib/name.log", "lib/name.file", "lib/keep.o",
	"lib/test/name/log.file", "lib/test/ver1/name/log.file", "test/lib/name.file",
	"a/b/c/lib/name.file", "name/file.txt", "name/lib/log.file", "lastname/log.file",
	"firstname/log.file", "doc/a.md", "doc/x/y/a.md", "doc/api/index.md",
	"build/out.o", "build/keep.o", "foo/bar/x", "x/foo/y/z",
}

func TestConformanceWithGit(t *testing.T) {
	git, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git is not installed; skipping the gitignore conformance suite")
	}

	dir := t.TempDir()
	run(t, git, dir, "init", "-q")
	dirs := materialize(t, dir, conformancePaths)

	for _, tc := range conformanceCases {
		t.Run(tc.name, func(t *testing.T) {
			content := strings.Join(tc.rules, "\n") + "\n"
			if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}
			want := gitIgnored(t, git, dir, conformancePaths)

			f := globbing.NewFilter()
			if err := f.Parse("", strings.NewReader(content)); err != nil {
				t.Fatalf("Parse(%q): %v", content, err)
			}

			for _, p := range conformancePaths {
				var got bool
				if dirs[p] {
					got = f.MatchDir(p)
				} else {
					got = f.Match(p)
				}
				if got != want[p] {
					t.Errorf("rules %q: Match(%q) = %v, git check-ignore says %v", tc.rules, p, got, want[p])
				}
			}
		})
	}
}

// TestConformanceNestedGitignore checks that a base directory behaves like a
// .gitignore file placed in that directory.
func TestConformanceNestedGitignore(t *testing.T) {
	git, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git is not installed; skipping the gitignore conformance suite")
	}

	dir := t.TempDir()
	run(t, git, dir, "init", "-q")
	paths := []string{
		"main.o", "sub/main.o", "sub/keep.o", "sub/deep/main.o",
		"sub/build/x", "build/x", "other/main.o",
	}
	dirs := materialize(t, dir, paths)

	root := "*.o\n"
	nested := "!keep.o\n/build/\n"
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(root), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sub", ".gitignore"), []byte(nested), 0o600); err != nil {
		t.Fatal(err)
	}
	want := gitIgnored(t, git, dir, paths)

	f := globbing.NewFilter()
	if err := f.Parse("", strings.NewReader(root)); err != nil {
		t.Fatal(err)
	}
	if err := f.Parse("/sub", strings.NewReader(nested)); err != nil {
		t.Fatal(err)
	}

	for _, p := range paths {
		var got bool
		if dirs[p] {
			got = f.MatchDir(p)
		} else {
			got = f.Match(p)
		}
		if got != want[p] {
			t.Errorf("Match(%q) = %v, git check-ignore says %v", p, got, want[p])
		}
	}
}

func materialize(t *testing.T, root string, paths []string) map[string]bool {
	t.Helper()
	dirs := map[string]bool{}
	for _, p := range paths {
		for d := filepath.Dir(p); d != "."; d = filepath.Dir(d) {
			dirs[d] = true
		}
	}
	for _, p := range paths {
		full := filepath.Join(root, filepath.FromSlash(p))
		if dirs[p] {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for d := range dirs {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(d)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return dirs
}

func gitIgnored(t *testing.T, git, dir string, paths []string) map[string]bool {
	t.Helper()
	args := append([]string{"-C", dir, "check-ignore", "--no-index", "--"}, paths...)
	cmd := exec.Command(git, args...)
	cmd.Env = isolatedEnv()
	out, err := cmd.Output()
	// check-ignore exits 1 when nothing matched, which is not an error here.
	if err != nil && cmd.ProcessState.ExitCode() != 1 {
		t.Fatalf("git check-ignore: %v", err)
	}
	ignored := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			ignored[filepath.ToSlash(line)] = true
		}
	}
	return ignored
}

func run(t *testing.T, git, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command(git, append([]string{"-C", dir}, args...)...)
	cmd.Env = isolatedEnv()
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

// isolatedEnv keeps the user's global and system git configuration — a
// core.excludesFile in particular — out of the comparison.
func isolatedEnv() []string {
	return append(os.Environ(),
		"GIT_CONFIG_GLOBAL="+os.DevNull,
		"GIT_CONFIG_SYSTEM="+os.DevNull,
		"GIT_CONFIG_NOSYSTEM=1",
	)
}
