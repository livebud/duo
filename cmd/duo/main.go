package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/livebud/cli"
	"github.com/livebud/duo"
	"github.com/livebud/duo/internal/cli/graceful"
	"github.com/livebud/duo/internal/cli/hot"
	"github.com/livebud/duo/internal/cli/pubsub"
	"github.com/livebud/duo/internal/cli/virtual"
	"github.com/livebud/duo/internal/resolver"
	"github.com/livebud/watcher"
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
		fs := http.FileServer(http.FS(s))
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
	return ext == ".html" || ext == ".svelte" || ext == ".duo"
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

func removeExt(path string) string {
	return path[:len(path)-len(filepath.Ext(path))]
}

func (s *Serve) serveError(fi fs.FileInfo, err error) (fs.File, error) {
	pre := fmt.Sprintf(`<pre>%s</pre>`, err.Error())
	html := []byte(fmt.Sprintf(htmlPage, pre+"\n"+liveReloadScript))
	// Create a buffered file
	bf := &virtual.File{
		Path:    "error",
		Data:    html,
		Mode:    fi.Mode(),
		ModTime: fi.ModTime(),
	}
	return bf.Open(), nil
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
		bf := &virtual.File{
			Path:    name,
			Data:    []byte(code),
			Mode:    fi.Mode(),
			ModTime: fi.ModTime(),
		}
		return bf.Open(), nil
	}
	jsCode, err := duo.Generate(name, code)
	if err != nil {
		return nil, fmt.Errorf("error generating %s: %w", name, err)
	}
	bf := &virtual.File{
		Path:    name,
		Data:    []byte(jsCode),
		Mode:    fi.Mode(),
		ModTime: fi.ModTime(),
	}
	return bf.Open(), nil
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
	if f.Close(); err != nil {
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
	bf := &virtual.File{
		Path:    name,
		Data:    html,
		Mode:    fi.Mode(),
		ModTime: fi.ModTime(),
	}
	return bf.Open(), nil
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
