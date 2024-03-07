package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/livebud/cli"
	"github.com/livebud/duo"
	"github.com/livebud/duo/internal/cli/graceful"
	"github.com/livebud/duo/internal/cli/hot"
	"github.com/livebud/duo/internal/cli/pubsub"
	"github.com/livebud/duo/internal/resolver"
	"github.com/livebud/watcher"
	"github.com/matthewmueller/virt"
	"golang.org/x/sync/errgroup"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cli := cli.New("duo", "duo templating language")

	{ // serve [flags] [dir]
		cmd := new(Serve)
		cli := cli.Command("serve", "serve a directory")
		cli.Flag("listen", "address to listen on").String(&cmd.Listen).Default(":3000")
		cli.Flag("live", "enable live reloading").Bool(&cmd.Live).Default(true)
		cli.Flag("open", "open browser").Bool(&cmd.Browser).Default(true)
		cli.Arg("dir").String(&cmd.Dir).Default(".")
		cli.Run(cmd.Run)
	}

	return cli.Parse(context.Background(), os.Args[1:]...)
}

type Serve struct {
	Listen  string
	Live    bool
	Dir     string
	Browser bool
}

func (s *Serve) Run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	ps := pubsub.New()
	host, portStr, err := net.SplitHostPort(s.Listen)
	if err != nil {
		return err
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return err
	}
	// Find the next available port
	ln, port, err := findNextPort(host, port)
	if err != nil {
		return err
	}
	defer ln.Close()
	url := formatAddr(host, port)
	fmt.Println("Listening on", url)
	eg.Go(s.serve(ctx, ln, ps))
	if s.Live {
		eg.Go(s.watch(ctx, ps))
	}
	if s.Browser {
		if err := exec.CommandContext(ctx, "open", url).Run(); err != nil {
			return err
		}
	}
	return eg.Wait()
}

func (s *Serve) serve(ctx context.Context, ln net.Listener, ps pubsub.Subscriber) func() error {
	return func() error {
		fs := &fileServer{s.Dir}
		return graceful.Serve(ctx, ln, s.handler(hot.New(ps), fs))
	}
}

// Minimal favicon
var favicon = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d, 0x49,
	0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x08, 0x02,
	0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xde, 0x00, 0x00, 0x00, 0x01, 0x73, 0x52,
	0x47, 0x42, 0x00, 0xae, 0xce, 0x1c, 0xe9, 0x00, 0x00, 0x00, 0x04, 0x67, 0x41,
	0x4d, 0x41, 0x00, 0x00, 0xb1, 0x8f, 0x0b, 0xfc, 0x61, 0x05, 0x00, 0x00, 0x00,
	0x09, 0x70, 0x48, 0x59, 0x73, 0x00, 0x00, 0x0e, 0xc4, 0x00, 0x00, 0x0e, 0xc4,
	0x01, 0x95, 0x2b, 0x0e, 0x1b, 0x00, 0x00, 0x00, 0x0c, 0x49, 0x44, 0x41, 0x54,
	0x08, 0xd7, 0x63, 0xf8,
}

func (s *Serve) handler(live http.Handler, fs http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/favicon.ico":
			w.Header().Set("Content-Type", "image/png")
			w.Write(favicon)
		case "/.live":
			live.ServeHTTP(w, r)
		default:
			w.Header().Set("Cache-Control", "no-store")
			fs.ServeHTTP(w, r)
		}
	})
}

const liveReloadScript = `
<script>
var es = new EventSource('/.live');
es.onmessage = function(e) { window.location.reload(); }
</script>
`

const htmlPage = `<!doctype html>
<html>
<head>
	<meta charset="utf-8">
</head>
<body>
	<main>
		%s
	</main>
</body>
</html>
`

func isView(path string) bool {
	ext := filepath.Ext(path)
	return ext == "" || ext == ".html" || ext == ".svelte" || ext == ".duo"
}

func openFile(paths ...string) (f *os.File, err error) {
	for _, path := range paths {
		f, err = os.Open(path)
		if nil == err {
			return f, nil
		}
	}
	return nil, err
}

func fileExists(paths ...string) (path string, err error) {
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		} else if !errors.Is(err, fs.ErrNotExist) {
			return "", err
		}
	}
	return "", fs.ErrNotExist
}

func removeExt(path string) string {
	return path[:len(path)-len(filepath.Ext(path))]
}

func (s *Serve) serveError(fi fs.FileInfo, err error) (fs.File, error) {
	pre := fmt.Sprintf(`<pre>%s</pre>`, err.Error())
	html := []byte(fmt.Sprintf(htmlPage, pre+"\n"+liveReloadScript))
	// Create a buffered file
	bf := &virt.File{
		Path:    "error",
		Data:    html,
		Mode:    fi.Mode(),
		ModTime: fi.ModTime(),
	}
	return virt.Open(bf), nil
}

var clientCode = `
<script type="module">
	import { render as preactRender, h, hydrate } from 'https://cdn.jsdelivr.net/npm/preact@10.15.1/+esm'
	import Proxy from 'https://esm.run/internal/proxy'
	import Component from %q

	export function render(Component, target, props = {}) {
		const proxy = Proxy(props)
		const component = Component(h, proxy)
		hydrate(h(component, proxy, []), target)
		window.requestAnimationFrame(() => {
			props.subscribe(() => {
				preactRender(h(component, props, []), target)
			})
		})
	}

	render(Component, document.querySelector('main'))
</script>
`

func (s *Serve) openClient(name string) (fs.File, error) {
	f, err := openFile(
		filepath.Join(s.Dir, name),
		filepath.Join(s.Dir, removeExt(name)+".svelte"),
		filepath.Join(s.Dir, removeExt(name)+".html"),
		filepath.Join(s.Dir, removeExt(name)+".duo"),
	)
	if err != nil {
		return nil, fmt.Errorf("error opening %s: %w", name, err)
	}
	fi, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("error stating %s: %w", name, err)
	}
	// If we detect HTML, inject the live reload script
	code, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	// Close the existing file because we don't need it anymore
	if f.Close(); err != nil {
		return nil, err
	}
	if filepath.Ext(f.Name()) == ".js" {
		bf := &virt.File{
			Path:    name,
			Data:    []byte(code),
			Mode:    fi.Mode(),
			ModTime: fi.ModTime(),
		}
		return virt.Open(bf), nil
	}
	jsCode, err := duo.Generate(name, code)
	if err != nil {
		return nil, fmt.Errorf("error generating %s: %w", name, err)
	}
	bf := &virt.File{
		Path:    name,
		Data:    []byte(jsCode),
		Mode:    fi.Mode(),
		ModTime: fi.ModTime(),
	}
	return virt.Open(bf), nil
}

type fileServer struct {
	Dir string
}

func (f *fileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	urlPath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	ext := filepath.Ext(urlPath)
	if ext == ".js" {
		f.serveClient(w, r, urlPath)
	} else if isView(urlPath) {
		f.serveView(w, r, urlPath)
	} else {
		f.serveStatic(w, r, urlPath)
	}
	// file, err := f.fs.Open(filePath)
	// fmt.Println("got url", urlPath)
	// file, err := f.fs.Open(urlPath)
	// if err != nil {
	// 	if errors.Is(err, fs.ErrNotExist) {
	// 		http.Error(w, err.Error(), http.StatusNotFound)
	// 		return
	// 	}
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// }
	// defer file.Close()
	// fi, err := file.Stat()
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	// if fi.IsDir() {
	// 	file.Close()
	// 	file, err = f.fs.Open(path.Join(urlPath, "index.html"))
	// 	if err != nil {
	// 		if errors.Is(err, fs.ErrNotExist) {
	// 			http.Error(w, err.Error(), http.StatusNotFound)
	// 			return
	// 		}
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 		return
	// 	}
	// }
	// fmt.Println("got file", file)
	// if
	// fmt.Println("serving", urlPath, path.Clean(urlPath))
	// if !strings.HasPrefix(upath, "/") {
	// 	upath = "/" + upath
	// 	r.URL.Path = upath
	// }
	// code, err := f.fs.Open(path.Clean(upath))
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusNotFound)
	// 	return
	// }
	// // serveFile(w, r, f.root, path.Clean(upath), true)
	// http.FileServer(f.fs).ServeHTTP(w, r)
}

func (f *fileServer) serveView(w http.ResponseWriter, r *http.Request, name string) {
	stat, err := os.Stat(filepath.Join(f.Dir, name))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if stat.IsDir() {
		name = path.Join(name, "index.html")
	}
	extless := filepath.Join(f.Dir, removeExt(name))
	filePath, err := fileExists(
		extless+".svelte",
		extless+".html",
		extless+".duo",
	)
	if err != nil {
		f.serveError(w, r, err)
		return
	}
	template := duo.New(resolver.New("."))
	buffer := new(bytes.Buffer)
	props := map[string]string{}
	values := r.URL.Query()
	for key := range values {
		props[key] = values.Get(key)
	}
	if err := template.Render(buffer, filePath, props); err != nil {
		f.serveError(w, r, err)
		return
	}
	html := buffer.Bytes()
	// Inject the live reload script
	if bytes.Contains(html, []byte(`<html>`)) {
		html = append(html, []byte(liveReloadScript)...)
		html = append(html, []byte(fmt.Sprintf(clientCode, "/"+removeExt(name)+".js"))...)
	} else {
		client := fmt.Sprintf(clientCode, "/"+removeExt(name)+".js")
		body := fmt.Sprintf("%s\n%s\n%s", string(html), liveReloadScript, client)
		html = []byte(fmt.Sprintf(htmlPage, body))
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(html)
}

func (f *fileServer) serveClient(w http.ResponseWriter, r *http.Request, name string) {
	extless := filepath.Join(f.Dir, removeExt(name))
	filePath, err := fileExists(
		extless+".js",
		extless+".svelte",
		extless+".html",
		extless+".duo",
	)
	if err != nil {
		f.serveClientError(w, r, err)
		return
	}
	code, err := os.ReadFile(filePath)
	if err != nil {
		f.serveClientError(w, r, err)
		return
	}
	if filepath.Ext(filePath) == ".js" {
		w.Header().Set("Content-Type", "text/javascript")
		w.Write(code)
		return
	}
	jsCode, err := duo.Generate(name, code)
	if err != nil {
		f.serveClientError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "text/javascript")
	w.Write([]byte(jsCode))
}

func (f *fileServer) serveStatic(w http.ResponseWriter, r *http.Request, name string) {
	http.ServeFile(w, r, filepath.Join(f.Dir, name))
}

func (f *fileServer) serveError(w http.ResponseWriter, _ *http.Request, err error) {
	pre := fmt.Sprintf(`<pre>%s</pre>`, err.Error())
	html := []byte(fmt.Sprintf(htmlPage, pre+"\n"+liveReloadScript))
	w.Header().Set("Content-Type", "text/html")
	w.Write(html)
}

func (f *fileServer) serveClientError(w http.ResponseWriter, _ *http.Request, err error) {
	script := fmt.Sprintf(`console.error(%q)`, err.Error())
	w.Header().Set("Content-Type", "text/javascript")
	w.Write([]byte(script))
}

func (s *Serve) Open(name string) (fs.File, error) {
	if filepath.Ext(name) == ".js" {
		return s.openClient(name)
	} else if !isView(name) {
		return os.Open(filepath.Join(s.Dir, name))
	}
	var f *os.File
	if filepath.Base(name) == "index.html" {
		file, err := openFile(
			filepath.Join(s.Dir, removeExt(name)+".svelte"),
			filepath.Join(s.Dir, removeExt(name)+".html"),
			filepath.Join(s.Dir, removeExt(name)+".duo"),
		)
		if err != nil {
			return nil, err
		}
		f = file
	}
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	// Skip directories
	if fi.IsDir() {
		return f, nil
	}
	defer f.Close()
	// Close the existing file because we don't need it anymore
	if err := f.Close(); err != nil {
		return nil, err
	}
	template := duo.New(resolver.New(s.Dir))
	buffer := new(bytes.Buffer)
	if err := template.Render(buffer, name, map[string]interface{}{
		"greeting": "hello",
	}); err != nil {
		return s.serveError(fi, fmt.Errorf("error rendering %s: %w", name, err))
	}
	html := buffer.Bytes()
	// Inject the live reload script
	if bytes.Contains(html, []byte(`<html>`)) {
		html = append(html, []byte(liveReloadScript)...)
		html = append(html, []byte(fmt.Sprintf(clientCode, "/"+removeExt(name)+".js"))...)
	} else {
		client := fmt.Sprintf(clientCode, "/"+removeExt(name)+".js")
		body := fmt.Sprintf("%s\n%s\n%s", string(html), liveReloadScript, client)
		html = []byte(fmt.Sprintf(htmlPage, body))
	}
	// Create a buffered file
	bf := &virt.File{
		Path:    name,
		Data:    html,
		Mode:    fi.Mode(),
		ModTime: fi.ModTime(),
	}
	return virt.Open(bf), nil
}

func (s *Serve) watch(ctx context.Context, ps pubsub.Publisher) func() error {
	return func() error {
		return watcher.Watch(ctx, s.Dir, func(events []watcher.Event) error {
			if len(events) == 0 {
				return nil
			}
			event := events[0]
			ps.Publish(string(event.Op), []byte(event.Path))
			return nil
		})
	}
}

// Find the next available port starting at 3000
func findNextPort(host string, port int) (net.Listener, int, error) {
	for i := 0; i < 100; i++ {
		ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port+i))
		if err == nil {
			return ln, port + i, nil
		}
	}
	return nil, 0, fmt.Errorf("could not find an available port")
}

func formatAddr(host string, port int) string {
	if host == "" {
		host = "0.0.0.0"
	}
	return fmt.Sprintf("http://%s:%d", host, port)
}
