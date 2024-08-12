package e2e_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

const testdata = "../testdata"

func Test(t *testing.T) {
	is := is.New(t)
	dir, err := filepath.Abs(testdata)
	is.NoErr(err)
	des, err := os.ReadDir(dir)
	is.NoErr(err)
	for _, de := range des {
		if !de.IsDir() || strings.HasPrefix(de.Name(), "_") {
			continue
		}
		t.Run(de.Name(), func(t *testing.T) {
			is := is.New(t)
			e2e, err := os.ReadFile(filepath.Join(dir, de.Name(), "e2e.txt"))
			if err != nil {
				if os.IsNotExist(err) {
					t.Skip("e2e.txt not found")
					return
				}
				t.Fatal(err)
			}
			input, err := os.ReadFile(filepath.Join(dir, de.Name(), "input.svelte"))
			is.NoErr(err)
			_, _ = e2e, input
		})
	}
}
