package js_test

import (
	"os"
	"testing"
	"text/template"

	"github.com/livebud/duo/internal/js"
	"github.com/matthewmueller/diff"
)

func equal(t *testing.T, path, input, expected string) {
	t.Helper()
	if path == "" {
		path = input
	}
	t.Run(path, func(t *testing.T) {
		t.Helper()
		ast, err := js.ParseScript(input)
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

func TestSample(t *testing.T) {
	equal(t, "", `const a = 'hello';`, `const a = "hello";`)
	equal(t, "", `const a: string = 'hello';`, `const a = "hello";`)
	equal(t, "", `export let props: Props = []`, `export let props = [];`)
	equal(t, "", `import Sub from './04-sub.html';`, `import Sub from "./04-sub.html";`)
	equal(t, "", `import type Sub from './04-sub.html';`, ``)
	equal(t, "", `import { Sub } from './04-sub.html';`, `import { Sub } from "./04-sub.html";`)
}
