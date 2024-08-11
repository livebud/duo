package esbuild_test

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/livebud/duo/internal/esbuild"
	"github.com/matryer/is"
)

func TestHTTP(t *testing.T) {
	is := is.New(t)
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "index.js"), []byte(`
		import { uid } from 'https://esm.run/uid'
		export function createElement() { return uid() }
	`), 0644)
	is.NoErr(err)
	file, err := esbuild.BuildOne(esbuild.BuildOptions{
		AbsWorkingDir: dir,
		EntryPoints:   []string{"./index.js"},
		Plugins: []esbuild.Plugin{
			esbuild.HTTP(http.DefaultClient),
		},
		Format:   esbuild.FormatESModule,
		Platform: esbuild.PlatformBrowser,
		// Add "import" condition to support svelte/internal
		// https://esbuild.github.io/api/#how-conditions-work
		Conditions: []string{"browser", "default", "import"},
		// Always bundle, use plugins to granularly mark files as external
		Bundle: true,
	})
	is.NoErr(err)
	code := string(file.Contents)
	is.True(strings.Contains(code, "function createElement() {"))
	is.True(strings.Contains(code, "(t + 256).toString(16).substring(1)"))
	is.True(strings.Contains(code, "Math.random()"))
}

func TestHTTPDepOfDep(t *testing.T) {
	is := is.New(t)
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "index.js"), []byte(`
 		import { uid } from 'https://esm.run/uid/secure'
 		export function createElement() { return uid() }
	`), 0644)
	is.NoErr(err)
	file, err := esbuild.BuildOne(esbuild.BuildOptions{
		AbsWorkingDir: dir,
		EntryPoints:   []string{"./index.js"},
		Plugins: []esbuild.Plugin{
			esbuild.HTTP(http.DefaultClient),
		},
		Format:   esbuild.FormatESModule,
		Platform: esbuild.PlatformBrowser,
		// Add "import" condition to support svelte/internal
		// https://esbuild.github.io/api/#how-conditions-work
		Conditions: []string{"browser", "default", "import"},
		// Always bundle, use plugins to granularly mark files as external
		Bundle: true,
	})
	is.NoErr(err)
	code := string(file.Contents)
	is.True(strings.Contains(code, "function createElement() {"))
	is.True(strings.Contains(code, "(n + 256).toString(16).substring(1)"))
	is.True(strings.Contains(code, "crypto.getRandomValues"))
}
