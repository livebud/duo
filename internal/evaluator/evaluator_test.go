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
			actual = strings.ReplaceAll(actual, "  ", "")
			actual = strings.ReplaceAll(actual, "\t", "")
			actual = strings.ReplaceAll(actual, "\n", "")
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
	equal(t, "attributes", `<hr target={target} />`, Map{"target": "_blank"}, `<hr target="_blank"/>`)
	equal(t, "attributes", `<hr target={target} />`, Map{}, `<hr/>`)
	equal(t, "attributes", `<hr name="{target}-{name}" />`, Map{"target": "_blank", "name": "anki"}, `<hr name="_blank-anki"/>`)
	equal(t, "attributes", `<hr name="{target}-{name}" />`, Map{"target": "_blank"}, `<hr name="_blank-"/>`)
	equal(t, "attributes", `<hr name="{target}-{name}" />`, Map{"name": "anki"}, `<hr name="-anki"/>`)
	equal(t, "attributes", `<hr target="{target}-{name}" />`, Map{}, `<hr target="-"/>`)
	equal(t, "attributes", `<hr {name} />`, Map{"name": "hello"}, `<hr name="hello"/>`)
	equal(t, "attributes", `<hr {name} />`, Map{}, `<hr/>`)
	// TODO: Should this be `<h1 name></h1>`?
	equal(t, "attributes", `<h1 name=""></h1>`, Map{}, `<h1></h1>`)
}

func TestFile(t *testing.T) {
	equalFile(t, "01-greeting.html", Map{}, `<h1></h1>`)
	equalFile(t, "01-greeting.html", Map{"greeting": "hi"}, `<h1>hi</h1>`)
	equalFile(t, "02-attribute.html", Map{}, `<div><hr/><hr/><hr/><hr name="-"/></div>`)
	equalFile(t, "02-attribute.html", Map{"name": "anki"}, `<div><hr name="anki"/><hr name="anki"/><hr name="anki"/><hr name="-anki"/></div>`)
	equalFile(t, "02-attribute.html", Map{"target": "window", "name": "anki"}, `<div><hr name="anki"/><hr name="anki"/><hr name="anki"/><hr name="window-anki"/></div>`)
}
