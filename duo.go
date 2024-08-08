package duo

import (
	"io"
	"net/http"

	"github.com/livebud/duo/internal/dom"
	"github.com/livebud/duo/internal/parser"
	"github.com/livebud/duo/internal/resolver"
	"github.com/livebud/duo/internal/ssr"
)

func Generate(path string, code []byte) (string, error) {
	doc, err := parser.Parse(path, string(code))
	if err != nil {
		return "", err
	}
	return dom.Generate(doc)
}

type Input struct {
	// Dir is the root directory of the templates. Defaults to cwd.
	Dir string
	// Cache enables caching of parsed templates. Defaults to false.
	Cache bool
	// Minify enables minification of the output. Defaults to false.
	Minify bool
	// Live attaches a live-reload script to the rendered HTML. Defaults to "".
	Live string
}

func New(in Input) *Template {
	return &Template{
		evaluator: &ssr.Renderer{
			Resolver: resolver.New(in.Dir),
		},
	}
}

// func Development() *Template {

// }

// func Production() *Template {

// }

type Template struct {
	evaluator *ssr.Renderer
}

func (t *Template) Render(w io.Writer, path string, props map[string]any) error {
	return t.evaluator.Render(w, path, props)
}

func (t *Template) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// ...
}
