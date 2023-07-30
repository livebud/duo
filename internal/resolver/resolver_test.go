package resolver_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/livebud/duo/internal/resolver"
	"github.com/matryer/is"
)

func TestResolveFromRoot(t *testing.T) {
	is := is.New(t)
	dir := t.TempDir()
	is.NoErr(os.MkdirAll(filepath.Join(dir, "users"), 0755))
	is.NoErr(os.WriteFile(filepath.Join(dir, "users", "index.duo"), []byte(`<h1>Index</h1>`), 0644))
	is.NoErr(os.WriteFile(filepath.Join(dir, "users", "about.duo"), []byte(`<h1>About</h1>`), 0644))
	res := resolver.New(dir)
	file, err := res.Resolve(&resolver.Resolve{
		Path: "./users/index.duo",
	})
	is.NoErr(err)
	is.Equal(file.Path, filepath.Join(dir, "users", "index.duo"))
	is.Equal(string(file.Code), `<h1>Index</h1>`)
}

func TestResolveFromFile(t *testing.T) {
	is := is.New(t)
	dir := t.TempDir()
	is.NoErr(os.MkdirAll(filepath.Join(dir, "users"), 0755))
	is.NoErr(os.WriteFile(filepath.Join(dir, "users", "index.duo"), []byte(`<h1>Index</h1>`), 0644))
	is.NoErr(os.WriteFile(filepath.Join(dir, "users", "about.duo"), []byte(`<h1>About</h1>`), 0644))
	res := resolver.New(dir)
	file, err := res.Resolve(&resolver.Resolve{
		From: filepath.Join(dir, "users", "index.duo"),
		Path: "./about.duo",
	})
	is.NoErr(err)
	is.Equal(file.Path, filepath.Join(dir, "users", "about.duo"))
	is.Equal(string(file.Code), `<h1>About</h1>`)
}
