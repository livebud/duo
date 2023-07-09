package dom_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/livebud/duo/internal/dom"
	"github.com/livebud/duo/internal/parser"
	"github.com/matthewmueller/diff"
)

func equal(t *testing.T, name, input, expected string) {
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
		actual, err := dom.Generate(doc)
		if err != nil {
			actual = err.Error()
		}
		if actual == expected {
			return
		}
		var b bytes.Buffer
		b.WriteString("\n\x1b[4mInput\x1b[0m:\n")
		b.WriteString(input)
		b.WriteString("\n\x1b[4mExpected\x1b[0m:\n")
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

func equalFile(t *testing.T, name string) {
	t.Helper()
	input, err := os.ReadFile(filepath.Join("..", "testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	expected, err := os.ReadFile(filepath.Join("..", "testdata", replaceExt(name, ".js.txt")))
	if err != nil {
		t.Fatal(err)
	}
	equal(t, name, string(input), string(expected))
}

func replaceExt(path, ext string) string {
	return path[:len(path)-len(filepath.Ext(path))] + ext
}

func TestFile(t *testing.T) {
	equalFile(t, "01-greeting.html")
	equalFile(t, "02-attribute.html")
}
