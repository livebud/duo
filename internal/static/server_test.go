package static_test

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"testing"

	"github.com/livebud/duo/internal/static"
	"github.com/matryer/is"
	"github.com/matthewmueller/diff"
	"github.com/matthewmueller/virt"
)

func equal(t testing.TB, handler http.Handler, req *http.Request, expect string) {
	t.Helper()
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	res := rec.Result()
	defer res.Body.Close()
	actual, err := httputil.DumpResponse(res, true)
	if err != nil {
		t.Fatal(err)
	}
	diff.TestHTTP(t, string(actual), expect)
}

func contains(t testing.TB, handler http.Handler, req *http.Request, contains ...string) {
	t.Helper()
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	res := rec.Result()
	defer res.Body.Close()
	actual, err := httputil.DumpResponse(res, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, contain := range contains {
		if !strings.Contains(string(actual), contain) {
			t.Fatalf("expected %s to contain %q", string(actual), contain)
		}
	}
}

func TestHelloWorld(t *testing.T) {
	is := is.New(t)
	dir := t.TempDir()
	fsys := virt.Map{
		"index.svelte": `<h1>hello, world!</h1>`,
	}
	is.NoErr(virt.Sync(fsys, dir))
	handler := static.Dir(dir)

	req := httptest.NewRequest("GET", "/", nil)
	equal(t, handler, req, `
		HTTP/1.1 200 OK
		Connection: close
		Content-Type: text/html; charset=utf-8

		<html><head></head><body><main id="svelte"><h1>hello, world!</h1></main><script type="module" src="/index.svelte.js"></script><script id="props" type="text/template">{}</script></body></html>
	`)

	req = httptest.NewRequest("GET", "/index.svelte.js", nil)
	contains(t, handler, req,
		`HTTP/1.1 200 OK`,
		`Content-Type: application/javascript`,
		`// http-url:https://esm.run/svelte@next`,
		`return h2("h1", {}, ["hello, world!"]);`,
		`target: document.getElementById("svelte"),`,
		`document.getElementById("props")?.textContent || "{}"`,
	)
}
