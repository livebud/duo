package dom_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	dom "github.com/livebud/duo/internal/dom2"
	"github.com/livebud/duo/internal/js"
	"github.com/matryer/is"
	"github.com/matthewmueller/diff"
)

func equal(t testing.TB, actual, expect string) {
	t.Helper()
	actualJs, err := js.Parse(actual)
	if err != nil {
		t.Fatal(err)
	}
	expectJs, err := js.Parse(expect)
	if err != nil {
		t.Fatal(err)
	}
	actualStr, err := js.Format(actualJs)
	if err != nil {
		t.Fatal(err)
	}
	expectStr, err := js.Format(expectJs)
	if err != nil {
		t.Fatal(err)
	}
	diff.TestString(t, actualStr, expectStr)
}

const testdata = "../../testdata"

func Test(t *testing.T) {
	is := is.New(t)
	dir, err := filepath.Abs(testdata)
	is.NoErr(err)
	des, err := os.ReadDir(dir)
	is.NoErr(err)
	for _, de := range des {
		if !de.IsDir() || strings.HasPrefix(de.Name(), "_") {
			continue
		}
		t.Run(de.Name(), func(t *testing.T) {
			is := is.New(t)
			expect, err := os.ReadFile(filepath.Join(dir, de.Name(), "dom.js"))
			if err != nil {
				if os.IsNotExist(err) {
					t.Skip("dom.js not found")
					return
				}
				t.Fatal(err)
			}
			input, err := os.ReadFile(filepath.Join(dir, de.Name(), "input.svelte"))
			is.NoErr(err)
			actual, err := dom.Generate(de.Name(), input)
			is.NoErr(err)
			equal(t, actual, string(expect))
		})
	}
}
