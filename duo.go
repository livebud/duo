package duo

import (
	"bytes"
	"io/fs"
	"net/http"

	"github.com/livebud/duo/internal/resolver"
	"github.com/livebud/duo/internal/ssr"
)

func New(fsys fs.FS) *View {
	return &View{ssr.New(resolver.New(fsys))}
}

type View struct {
	ssr *ssr.Renderer
}

func (d *View) Render(w http.ResponseWriter, path string, v interface{}) {
	html := new(bytes.Buffer)
	if err := d.ssr.Render(html, path, v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(html.Bytes())
}
