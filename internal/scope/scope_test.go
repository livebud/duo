package scope_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lithammer/dedent"
	"github.com/livebud/duo/internal/scope"
	"github.com/matthewmueller/diff"
)

func equalScope(t *testing.T, scope *scope.Scope, expected string) {
	actual := strings.TrimSpace(scope.String())
	expected = dedent.Dedent(expected)
	expected = strings.TrimSpace(expected)
	if actual == expected {
		return
	}
	var b bytes.Buffer
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
}

func TestNew(t *testing.T) {
	s := scope.New()
	s.Use("foo")
	s.Use("bar")
	s.Use("baz")
	equalScope(t, s, `
		"foo"
		"bar"
		"baz"
	`)
}

func TestScope(t *testing.T) {
	s := scope.New()
	s.Use("foo")
	s.Use("bar")
	s.Use("baz")
	s = s.New()
	s.Use("ok")
	s.Use("cool")
	s.Use("foo")
	equalScope(t, s, `
		"foo"
		"bar"
		"baz"

		"ok"
		"cool"
	`)
}

// func TestLookup(t *testing.T) {
// 	is := is.New(t)
// 	s := scope.New()
// 	s.Use("foo")
// 	s.Use("bar")
// 	s.Use("baz")
// 	s = s.New()
// 	s.Use("ok")
// 	s.Use("cool")
// 	foo := s.Use("foo")
// 	foo.IsMutable = true
// 	foo, ok := s.Lookup("foo")
// 	is.True(ok)
// 	is.True(foo.IsMutable)
// 	cool, ok := s.Lookup("cool")
// 	is.True(ok)
// 	is.True(!cool.IsMutable)
// 	bar, ok := s.Lookup("bar")
// 	is.True(ok)
// 	is.True(!bar.IsMutable)
// 	notfound, ok := s.Lookup("notfound")
// 	is.True(!ok)
// 	is.True(notfound == nil)
// }
