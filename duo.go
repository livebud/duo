package duo

import (
	"io"

	"github.com/livebud/duo/internal/dom"
	"github.com/livebud/duo/internal/evaluator"
	"github.com/livebud/duo/internal/parser"
	"github.com/livebud/duo/internal/resolver"
)

// func ParseFile(path string) (*Template, error) {
// 	return parse(path, func(path string) ([]byte, error) {
// 		return os.ReadFile(path)
// 	})
// }

// func ParseFS(fsys fs.FS, path string) (*Template, error) {
// 	return parse(path, func(path string) ([]byte, error) {
// 		return fs.ReadFile(fsys, path)
// 	})
// }

// func Parse(path string, code []byte) (*Template, error) {
// 	resolver := resolver.Embedded{path: code}
// 	doc, err := parser.Parse(path, string(code))
// 	if err != nil {
// 		return nil, err
// 	}
// 	evaluator := evaluator.New(resolver)
// 	return &Template{doc, evaluator}, nil
// }

func Generate(path string, code []byte) (string, error) {
	doc, err := parser.Parse(path, string(code))
	if err != nil {
		return "", err
	}
	return dom.Generate(doc)
}

// func parse(path string, readFile func(path string) ([]byte, error)) (*Template, error) {
// 	code, err := readFile(path)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return Parse(path, code)
// }

func New(resolver resolver.Interface) *Template {
	return &Template{evaluator.New(resolver)}
}

type Template struct {
	evaluator *evaluator.Evaluator
}

func (t *Template) Render(w io.Writer, path string, props interface{}) error {
	return t.evaluator.Evaluate(w, path, props)
}
