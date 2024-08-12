package static

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/livebud/duo/internal/esbuild"
	"github.com/livebud/duo/internal/resolver"
	"github.com/livebud/duo/internal/ssr"
)

func Serve(ln net.Listener, dir string) error {
	return http.Serve(ln, Dir(dir))
}

func Dir(dir string) http.Handler {
	fsys := os.DirFS(dir)
	resolver := resolver.New(fsys)
	return &Server{
		Dir:      dir,
		Fsys:     fsys,
		Resolver: resolver,
		SSR:      ssr.New(resolver),
		Cache:    false,
		Minify:   false,
		Layouts:  &layoutResolver{fsys},
		Frames:   &frameResolver{fsys},
		Errors:   &errorResolver{fsys},
		ClientPath: func(filePath string) string {
			return fmt.Sprintf("/%s.js", filePath)
		},
	}
}

type Server struct {
	Dir        string
	Fsys       fs.FS
	Resolver   resolver.Interface
	SSR        *ssr.Renderer
	Cache      bool
	Minify     bool
	Layouts    resolver.Interface
	Frames     resolver.Interface
	Errors     resolver.Interface
	ClientPath func(filePath string) string
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	urlPath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	ext := path.Ext(urlPath)
	switch ext {
	case ".js":
		s.serveJS(w, urlPath)
	case "":
		s.serveHTML(w, r, urlPath)
	default:
		s.serveAsset(w, r, urlPath)
	}
}

const defaultLayout = `<html><head></head><body><main id="svelte">{children}</main>{script}</body></html>`
const defaultError = `<html><head></head><body><main id="svelte"><h1>Internal Server Error</h1></main>{script}</body></html>`

func (s *Server) serveHTML(w http.ResponseWriter, r *http.Request, urlPath string) {
	filePath := path.Join(urlPath, "index.svelte")
	page, err := s.FindPage(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	props := map[string]interface{}{}
	query := r.URL.Query()
	for key := range query {
		props[key] = query.Get(key)
	}
	if err := s.Render(w, page, props); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) FindPage(filePath string) (*Page, error) {
	content, err := s.Resolver.Resolve(&resolver.Resolve{Path: filePath})
	if err != nil {
		return nil, err
	}
	page := &Page{
		Content: content,
	}
	if s.Layouts != nil {
		layout, err := s.Layouts.Resolve(&resolver.Resolve{Path: filePath})
		if err == nil {
			page.Layout = layout
		} else if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
	}
	// If page layout is still nil, set a default layout
	if page.Layout == nil {
		page.Layout = &resolver.File{
			Path: "layout.svelte",
			Code: []byte(defaultLayout),
		}
	}
	if s.Frames != nil {
		frame, err := s.Frames.Resolve(&resolver.Resolve{Path: filePath})
		if err == nil {
			// TODO: handle nested frames
			page.Frames = append(page.Frames, frame)
		} else if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
	}
	if s.Errors != nil {
		error, err := s.Errors.Resolve(&resolver.Resolve{Path: filePath})
		if err == nil {
			page.Error = error
		} else if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
	}
	// If page error is still nil, set a default error page
	if page.Error == nil {
		page.Error = &resolver.File{
			Path: "error.svelte",
			Code: []byte(defaultError),
		}
	}
	return page, nil
}

type Page struct {
	Layout  *resolver.File
	Frames  []*resolver.File
	Error   *resolver.File
	Content *resolver.File
}

func (s *Server) Render(w http.ResponseWriter, page *Page, props map[string]interface{}) error {
	children := new(bytes.Buffer)
	jsonProps, err := json.Marshal(props)
	if err != nil {
		return err
	}
	if err := s.SSR.Evaluate(children, page.Content.Path, page.Content.Code, props); err != nil {
		return s.SSR.Evaluate(w, page.Error.Path, page.Error.Code, props)
	}
	for _, frame := range page.Frames {
		props["children"] = children.String()
		children = new(bytes.Buffer)
		if err := s.SSR.Evaluate(children, frame.Path, frame.Code, props); err != nil {
			return s.SSR.Evaluate(w, page.Error.Path, page.Error.Code, props)
		}
	}
	props["children"] = children.String()
	props["script"] = fmt.Sprintf(`<script type="module" src=%q></script><script id="props" type="text/template">%s</script>`, s.ClientPath(page.Content.Path), string(jsonProps))
	children = new(bytes.Buffer)
	if err := s.SSR.Evaluate(children, page.Layout.Path, page.Layout.Code, props); err != nil {
		return err
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := children.WriteTo(w); err != nil {
		return err
	}
	return nil
}

const entryCode = `
	import { hydrate } from "https://esm.run/svelte@next";
	import Content from "./%[1]s";
	const props = document.getElementById("props")?.textContent || "{}";
	hydrate(Content, {
		target: document.getElementById("svelte"),
		props: JSON.parse(props),
	});
	window.prerenderReady = true
`

func (s *Server) serveJS(w http.ResponseWriter, urlPath string) {
	file, err := esbuild.BuildOne(esbuild.BuildOptions{
		AbsWorkingDir: s.Dir,
		EntryPoints:   []string{urlPath},
		Format:        esbuild.FormatESModule,
		Platform:      esbuild.PlatformBrowser,
		// Add "import" condition to support svelte/internal
		// https://esbuild.github.io/api/#how-conditions-work
		Conditions: []string{"browser", "default", "import"},
		// Always bundle, use plugins to granularly mark files as external
		Bundle: true,
		Plugins: []esbuild.Plugin{
			esbuild.HTTP(http.DefaultClient),
			esbuild.Svelte(s.Fsys),
			esbuild.Virtual(urlPath, func(args api.OnLoadArgs) (api.OnLoadResult, error) {
				componentPath := strings.TrimSuffix(urlPath, ".js")
				contents := fmt.Sprintf(entryCode, componentPath)
				return api.OnLoadResult{
					Contents:   &contents,
					ResolveDir: s.Dir,
				}, nil
			}),
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/javascript")
	w.Write(file.Contents)
}

func (s *Server) serveAsset(w http.ResponseWriter, r *http.Request, urlPath string) {
	// ...
}
