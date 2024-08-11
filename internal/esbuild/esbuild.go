package esbuild

import (
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

type (
	BuildOptions = api.BuildOptions
	OnLoadResult = api.OnLoadResult
	OnLoadArgs   = api.OnLoadArgs
	EntryPoint   = api.EntryPoint
	File         = api.OutputFile
	Plugin       = api.Plugin
)

var (
	JSXAutomatic    = api.JSXAutomatic
	FormatIIFE      = api.FormatIIFE
	FormatESModule  = api.FormatESModule
	PlatformNeutral = api.PlatformNeutral
	PlatformBrowser = api.PlatformBrowser
	LoaderTSX       = api.LoaderTSX
)

func BuildOne(options BuildOptions) (File, error) {
	result := api.Build(options)
	if len(result.Errors) > 0 {
		return File{}, &Error{result.Errors}
	}
	return result.OutputFiles[0], nil
}

func Build(options BuildOptions) ([]File, error) {
	result := api.Build(options)
	if len(result.Errors) > 0 {
		return nil, &Error{result.Errors}
	}
	return result.OutputFiles, nil
}

type Error struct {
	messages []api.Message
}

func (e *Error) Error() string {
	errors := api.FormatMessages(e.messages, api.FormatMessagesOptions{
		Color: true,
	})
	return strings.Join(errors, "\n")
}
