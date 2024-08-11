package esbuild

import (
	"github.com/evanw/esbuild/pkg/api"
)

// Virtual creates a virtual file that can be imported as an entry. The
// loader function is called when the file is imported and is like any other
// ESBuild loader.
//
// Note: This plugin doesn't handle resolving relative paths to virtual files.
// See the test for an example of this limitation.
func Virtual(filter string, loader func(api.OnLoadArgs) (api.OnLoadResult, error)) api.Plugin {
	return api.Plugin{
		Name: "virtual",
		Setup: func(epb api.PluginBuild) {
			epb.OnResolve(api.OnResolveOptions{Filter: filter}, func(args api.OnResolveArgs) (result api.OnResolveResult, err error) {
				result.Path = args.Path
				result.Namespace = filter
				return result, nil
			})
			epb.OnLoad(api.OnLoadOptions{Filter: filter, Namespace: filter}, loader)
		},
	}
}
