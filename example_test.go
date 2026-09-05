package globbing_test

import (
	"fmt"
	"strings"

	globbing "github.com/sntns/globbing-go"
)

func ExampleCompile() {
	p, err := globbing.Compile("*.log")
	if err != nil {
		panic(err)
	}
	fmt.Println(p.Match("/var/app.log"))
	fmt.Println(p.Match("/var/app.txt"))
	// Output:
	// true
	// false
}

func ExampleCompileBase() {
	// The rule was declared in /project/vendor, so it only applies below it.
	p, err := globbing.CompileBase("/project/vendor", "*.log")
	if err != nil {
		panic(err)
	}
	fmt.Println(p.Match("/project/vendor/build.log"))
	fmt.Println(p.Match("/project/build.log"))
	// Output:
	// true
	// false
}

func ExampleFilter_Parse() {
	gitignore := `
# build output
build/
*.o
!keep.o
`
	f := globbing.NewFilter()
	if err := f.Parse("", strings.NewReader(gitignore)); err != nil {
		panic(err)
	}

	for _, path := range []string{"/main.o", "/keep.o", "/main.c", "/build/main.c"} {
		fmt.Printf("%-14s %v\n", path, f.Match(path))
	}
	// Output:
	// /main.o        true
	// /keep.o        false
	// /main.c        false
	// /build/main.c  true
}

// A path below an excluded directory can never be re-included, exactly as in
// git. To keep part of a directory, leave the directory itself included.
func ExampleFilter_Match_reInclusion() {
	ineffective := globbing.NewFilter()
	if err := ineffective.Parse("", strings.NewReader("build/\n!build/keep.o\n")); err != nil {
		panic(err)
	}
	fmt.Println(ineffective.Match("/build/keep.o"))

	effective := globbing.NewFilter()
	if err := effective.Parse("", strings.NewReader("build/*\n!build/keep.o\n")); err != nil {
		panic(err)
	}
	fmt.Println(effective.Match("/build/keep.o"))
	// Output:
	// true
	// false
}

func ExampleFilter_MatchDir() {
	f := globbing.NewFilter(globbing.MustCompile("build/"))
	fmt.Println(f.MatchDir("/build"))
	fmt.Println(f.Match("/build"))
	// Output:
	// true
	// false
}

func ExampleSyntaxError() {
	_, err := globbing.Compile("name[a-z")
	fmt.Println(err)
	// Output:
	// globbing: unterminated bracket expression at offset 4 in pattern "name[a-z"
}
