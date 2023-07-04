package evaluator_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/livebud/duo/internal/evaluator"
	"github.com/livebud/duo/internal/parser"
	"github.com/matthewmueller/diff"
)

func equal(t *testing.T, name, input string, props interface{}, expected string) {
	t.Helper()
	if name == "" {
		name = input
	}
	t.Run(name, func(t *testing.T) {
		t.Helper()
		doc, err := parser.Parse(input)
		if err != nil {
			t.Fatal(err)
		}
		evaluator := evaluator.New(doc)
		str := new(strings.Builder)
		actual := ""
		if err := evaluator.Evaluate(str, props); err != nil {
			actual = err.Error()
		} else {
			// TODO: remove this, this should happen earlier
			actual = strings.TrimSpace(str.String())
		}
		if actual == expected {
			return
		}
		var b bytes.Buffer
		b.WriteString("\n\x1b[4mInput\x1b[0m:\n")
		b.WriteString(input)
		b.WriteString("\n\n")
		b.WriteString("\x1b[4mExpected\x1b[0m:\n")
		b.WriteString(expected)
		b.WriteString("\n\n")
		b.WriteString("\x1b[4mActual\x1b[0m: \n")
		b.WriteString(actual)
		b.WriteString("\n\n")
		b.WriteString("\x1b[4mDifference\x1b[0m: \n")
		b.WriteString(diff.String(expected, actual))
		b.WriteString("\n")
		t.Fatal(b.String())
	})
}

func equalFile(t *testing.T, name string, props interface{}, expected string) {
	t.Helper()
	input, err := os.ReadFile(filepath.Join("..", "testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	equal(t, name, string(input), props, expected)
}

type Map = map[string]interface{}

func TestSimple(t *testing.T) {
	equal(t, "simple", `<h1>hi</h1>`, nil, `<h1>hi</h1>`)
	equal(t, "greeting", `<h1>{greeting}</h1>`, nil, `<h1></h1>`)
	equal(t, "greeting", `<h1>{greeting}</h1>`, Map{"greeting": "hi"}, `<h1>hi</h1>`)
	equal(t, "greeting", `<h1>{greeting}</h1>`, Map{}, `<h1></h1>`)
	equal(t, "planet", `<h1>hello {planet}!</h1>`, Map{"planet": "mars"}, `<h1>hello mars!</h1>`)
	equal(t, "planet", `<h1>hello {planet}!</h1>`, Map{}, `<h1>hello !</h1>`)
	equal(t, "greeting_planet", `<h1>{greeting} {planet}!</h1>`, Map{"greeting": "hola", "planet": "Earth"}, `<h1>hola Earth!</h1>`)
}

func TestFile(t *testing.T) {
	equalFile(t, "01-greeting.svelte", Map{}, `<h1></h1>`)
	equalFile(t, "01-greeting.svelte", Map{"greeting": "hi"}, `<h1>hi</h1>`)
}
