package js_test

import (
	"os"
	"testing"
	"text/template"

	"github.com/livebud/duo/internal/js"
	"github.com/matthewmueller/diff"
)

func equalJS(t *testing.T, path, input, expected string) {
	t.Helper()
	if path == "" {
		path = input
	}
	t.Run(path, func(t *testing.T) {
		t.Helper()
		ast, err := js.Parse(input)
		if err != nil {
			diff.TestString(t, err.Error(), expected)
			return
		}
		actual := js.Print(ast)
		diff.TestString(t, actual, expected)
	})
	tpl := template.Must(template.New("test").Parse(`{{.}}`))
	tpl.Execute(os.Stdout, "Hello, World!")
}

func TestSampleJS(t *testing.T) {
	equalJS(t, "", `const a = 'hello';`, `const a = "hello";`)
	equalJS(t, "", `const a: string = 'hello';`, `const a = "hello";`)
	equalJS(t, "", `export let props: Props = []`, `export let props = [];`)
	equalJS(t, "", `import Sub from './04-sub.html';`, `import Sub from "./04-sub.html";`)
	equalJS(t, "", `import type Sub from './04-sub.html';`, ``)
	equalJS(t, "", `import { Sub } from './04-sub.html';`, `import { Sub } from "./04-sub.html";`)
}

func TestParse(t *testing.T) {
	ast, err := js.Parse("export let props: Props = []")
	if err != nil {
		t.Fatal(err)
	}
	actual := js.Print(ast)
	ast2, err := js.ParseTS("export let props: Props = []")
	if err != nil {
		t.Fatal(err)
	}
	actual2 := js.Print(ast2)
	expect := `export let props = [];`
	if actual != expect {
		t.Fatalf("expected %s, got %s", expect, actual)
	}
	if actual != actual2 {
		t.Fatalf("expected %s, got %s", actual, actual2)
	}
}

func BenchmarkParseJS(b *testing.B) {
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		js.Parse("export let props: Props = []")
	}
}

func BenchmarkParseTS(b *testing.B) {
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		js.ParseTS("export let props: Props = []")
	}
}
