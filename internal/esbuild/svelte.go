package esbuild

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/livebud/duo/internal/dom"
)

const namespace = "svelte"
const filter = `\.svelte$`

// Svelte plugin for compiling Svelte components
func Svelte(fsys fs.FS) api.Plugin {
	return api.Plugin{
		Name: "svelte",
		Setup: func(epb api.PluginBuild) {
			epb.OnResolve(api.OnResolveOptions{Filter: filter}, func(args api.OnResolveArgs) (result api.OnResolveResult, err error) {
				result.Path = args.Path
				result.Namespace = namespace
				return result, nil
			})
			epb.OnLoad(api.OnLoadOptions{Filter: filter, Namespace: namespace}, func(args api.OnLoadArgs) (api.OnLoadResult, error) {
				code, err := fs.ReadFile(fsys, filepath.Clean(args.Path))
				if err != nil {
					return api.OnLoadResult{}, fmt.Errorf("reading file: %w", err)
				}
				clientCode, err := dom.Generate(args.Path, code)
				if err != nil {
					return api.OnLoadResult{}, fmt.Errorf("generating client: %w", err)
				}
				return api.OnLoadResult{
					Contents:   &clientCode,
					Loader:     api.LoaderJS,
					ResolveDir: filepath.Dir(args.Path),
				}, nil
			})
		},
	}
}
