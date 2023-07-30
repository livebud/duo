package resolver

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

type Resolve struct {
	From string
	Path string
}

type File struct {
	Path string
	Code []byte
}

type Interface interface {
	Resolve(r *Resolve) (*File, error)
}

func New(dir string) *Resolver {
	return &Resolver{dir}
}

type Resolver struct {
	dir string
}

var _ Interface = (*Resolver)(nil)

func (r *Resolver) Resolve(res *Resolve) (*File, error) {
	dir := r.dir
	if res.From != "" {
		dir = filepath.Dir(res.From)
	}
	relPath := filepath.Join(dir, res.Path)
	code, err := os.ReadFile(relPath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", relPath, err)
	}
	return &File{
		Path: relPath,
		Code: code,
	}, nil
}

type Embedded map[string][]byte

var _ Interface = (*Embedded)(nil)

func (e Embedded) Resolve(res *Resolve) (*File, error) {
	dir := "."
	if res.From != "" {
		dir = path.Dir(res.From)
	}
	relPath := path.Join(dir, res.Path)
	code, ok := e[relPath]
	if !ok {
		return nil, fmt.Errorf("%s: %w", relPath, fs.ErrNotExist)
	}
	return &File{
		Path: res.Path,
		Code: code,
	}, nil
}
