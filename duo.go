package duo

import (
	"io"
	"io/fs"
	"os"

	"github.com/livebud/duo/internal/ast"
	"github.com/livebud/duo/internal/evaluator"
	"github.com/livebud/duo/internal/parser"
)

func ParseFile(path string) (*Template, error) {
	return parse(path, func(path string) ([]byte, error) {
		return os.ReadFile(path)
	})
}

func ParseFS(fsys fs.FS, path string) (*Template, error) {
	return parse(path, func(path string) ([]byte, error) {
		return fs.ReadFile(fsys, path)
	})
}

func Parse(path string, code []byte) (*Template, error) {
	doc, err := parser.Parse(string(code))
	if err != nil {
		return nil, err
	}
	evaluator := evaluator.New(doc)
	return &Template{doc, evaluator}, nil
}

func parse(path string, readFile func(path string) ([]byte, error)) (*Template, error) {
	code, err := readFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(path, code)
}

type Template struct {
	doc       *ast.Document
	evaluator *evaluator.Evaluator
}

func (t *Template) Render(w io.Writer, props interface{}) error {
	return t.evaluator.Evaluate(w, props)
}
