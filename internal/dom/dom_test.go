package dom_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/livebud/duo/internal/dom"
	"github.com/livebud/duo/internal/parser"
	"github.com/matthewmueller/diff"
)

func equal(t *testing.T, path, input, expected string) {
	t.Helper()
	if path == "" {
		path = input
	}
	t.Run(path, func(t *testing.T) {
		t.Helper()
		doc, err := parser.Parse(path, input)
		if err != nil {
			t.Fatal(err)
		}
		actual, err := dom.Print(doc)
		if err != nil {
			actual = err.Error()
		}
		diff.TestString(t, actual, expected)
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
	equalFile(t, "03-counter.html")
}
