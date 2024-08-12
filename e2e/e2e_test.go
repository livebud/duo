package e2e_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/livebud/duo/internal/static"
	"github.com/livebud/duo/internal/wd"
	"github.com/matryer/is"
)

func serve(t testing.TB, handler http.Handler) (*wd.Browser, func()) {
	t.Helper()
	server := httptest.NewServer(handler)
	browser, err := wd.Dial(server.URL)
	if err != nil {
		t.Fatalf("unable to open webdriver url: %v", err)
	}
	return browser, func() {
		browser.Close()
		server.Close()
	}
}

func Test01Counter(t *testing.T) {
	is := is.New(t)
	browser, close := serve(t, static.Dir("testdata/01-counter"))
	defer close()
	res, err := browser.Get("/")
	is.NoErr(err)
	is.Equal(res.Status(), 200)
}
