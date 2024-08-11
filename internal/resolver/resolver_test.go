package resolver_test

import (
	"testing"
	"testing/fstest"

	"github.com/livebud/duo/internal/resolver"
	"github.com/matryer/is"
)

func TestResolveFromRoot(t *testing.T) {
	is := is.New(t)
	fsys := fstest.MapFS{
		"users/index.duo": &fstest.MapFile{
			Data: []byte(`<h1>Index</h1>`),
		},
		"users/about.duo": &fstest.MapFile{
			Data: []byte(`<h1>About</h1>`),
		},
	}
	res := resolver.New(fsys)
	file, err := res.Resolve(&resolver.Resolve{
		Path: "./users/index.duo",
	})
	is.NoErr(err)
	is.Equal(file.Path, "users/index.duo")
	is.Equal(string(file.Code), `<h1>Index</h1>`)
}

func TestResolveFromFile(t *testing.T) {
	is := is.New(t)
	fsys := fstest.MapFS{
		"users/index.duo": &fstest.MapFile{
			Data: []byte(`<h1>Index</h1>`),
		},
		"users/about.duo": &fstest.MapFile{
			Data: []byte(`<h1>About</h1>`),
		},
	}
	res := resolver.New(fsys)
	file, err := res.Resolve(&resolver.Resolve{
		From: "users/index.duo",
		Path: "./about.duo",
	})
	is.NoErr(err)
	is.Equal(file.Path, "users/about.duo")
	is.Equal(string(file.Code), `<h1>About</h1>`)
}
