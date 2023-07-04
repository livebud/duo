package parser_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
		actual := parser.Print(input)
		actual = strings.ReplaceAll(actual, "\n", "")
		actual = strings.ReplaceAll(actual, "\t", "")
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

func equalFile(t *testing.T, name, expected string) {
	t.Helper()
	input, err := os.ReadFile(filepath.Join("..", "testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	equal(t, name, string(input), expected)
}

func Test(t *testing.T) {
	equal(t, "simple", "<h1>hi</h1>", `<h1>hi</h1>`)
	equal(t, "expr", "<h1>{greeting}</h1>", `<h1>{greeting}</h1>`)
	equal(t, "expr", "<hr/>", `<hr />`)
	equal(t, "expr", "<hr   />", `<hr />`)
}

func TestFile(t *testing.T) {
	equalFile(t, "01-greeting.svelte", `<script>export let greeting = "hello"; setInterval(() => { greeting += "o"; }, 500); </script><h1>{greeting}</h1>`)
}
