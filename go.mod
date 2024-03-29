module github.com/livebud/duo

go 1.22.0

require (
	github.com/evanw/esbuild v0.20.1
	github.com/hexops/valast v1.4.4
	github.com/ije/esbuild-internal v0.20.0
	github.com/lithammer/dedent v1.1.0
	github.com/livebud/cli v0.0.1
	github.com/livebud/watcher v0.0.2
	github.com/matryer/is v1.4.1
	github.com/matthewmueller/diff v0.0.2
	github.com/matthewmueller/virt v0.0.1
	github.com/tdewolff/parse/v2 v2.6.6
	golang.org/x/sync v0.6.0
)

require (
	github.com/bep/debounce v1.2.1 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/sabhiram/go-gitignore v0.0.0-20210923224102-525f6e181f06 // indirect
	github.com/sergi/go-diff v1.3.1 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	golang.org/x/mod v0.16.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/tools v0.19.0 // indirect
	mvdan.cc/gofumpt v0.6.0 // indirect
)

replace github.com/ije/esbuild-internal => ../../matthewmueller/esbuild-internal
