package js

import (
	"errors"
	"fmt"
	"strings"

	esbuild "github.com/evanw/esbuild/pkg/api"
	"github.com/ije/esbuild-internal/ast"
	"github.com/ije/esbuild-internal/config"
	"github.com/ije/esbuild-internal/js_parser"
	"github.com/ije/esbuild-internal/js_printer"
	"github.com/ije/esbuild-internal/logger"
	"github.com/ije/esbuild-internal/renamer"
	"github.com/ije/esbuild-internal/test"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/js"
)

type (
	IExpr          = js.IExpr
	BlockStmt      = js.BlockStmt
	IStmt          = js.IStmt
	ExprStmt       = js.ExprStmt
	ExportStmt     = js.ExportStmt
	ImportStmt     = js.ImportStmt
	FuncDecl       = js.FuncDecl
	VarDecl        = js.VarDecl
	BindingElement = js.BindingElement
	IBinding       = js.IBinding
	BinaryExpr     = js.BinaryExpr
	Var            = js.Var
	CondExpr       = js.CondExpr
	ArrowFunc      = js.ArrowFunc
	UnaryExpr      = js.UnaryExpr
	GroupExpr      = js.GroupExpr
	CallExpr       = js.CallExpr
	ReturnStmt     = js.ReturnStmt
	LiteralExpr    = js.LiteralExpr
	ArrayExpr      = js.ArrayExpr
	AST            = js.AST
	ObjectExpr     = js.ObjectExpr
	Property       = js.Property
	Arg            = js.Arg
	Element        = js.Element
	PropertyName   = js.PropertyName
	Args           = js.Args
	DotExpr        = js.DotExpr
	Params         = js.Params
	Scope          = js.Scope
)

var (
	VarToken        = js.VarToken
	LetToken        = js.LetToken
	StringToken     = js.StringToken
	AddToken        = js.AddToken
	EqToken         = js.EqToken
	OrToken         = js.OrToken
	IdentifierToken = js.IdentifierToken
)

func parseToString(script string) (string, error) {
	log := logger.NewDeferLog(logger.DeferLogNoVerboseOrDebug, nil)
	tree, ok := js_parser.Parse(log, test.SourceForTest(script), js_parser.OptionsFromConfig(&config.Options{
		TS: config.TSOptions{
			Parse: true,
			Config: config.TSConfig{
				VerbatimModuleSyntax: config.True,
			},
		},
	}))
	msgs := log.Done()
	text := ""
	for _, msg := range msgs {
		text += msg.String(logger.OutputOptions{}, logger.TerminalInfo{})
	}
	if !ok {
		return "", errors.New(text)
	}
	// 	// for _, imp := range tree.ImportRecords {
	// 	// 	fmt.Println(imp.Kind, imp.Path)
	// 	// }
	// 	// fmt.Println(tree.ImportRecords)
	// 	for _, part := range tree.Parts {
	// 		for _, stmt := range part.Stmts {
	// 			fmt.Println(valast.String(stmt.Data))
	// 		}
	// 	}
	// 	// for i, symbol := range tree.Symbols {
	// 	// 	if symbol.UseCountEstimate == 0 {
	// 	// 		symbol.UseCountEstimate += 10
	// 	// 	}
	// 	// 	tree.Symbols[i] = symbol
	// 	// 	// fmt.Println(symbol.OriginalName, symbol.UseCountEstimate)
	// 	// 	// if symbol.ImportItemStatus == ast.ImportItemMissing {
	// 	// 	// 	fmt.Println("ImportItemMissing", symbol.OriginalName)
	// 	// 	// }
	// 	// }
	// 	// fmt.Println(tree.Symbols)
	symbols := ast.NewSymbolMap(1)
	symbols.SymbolsForSource[0] = tree.Symbols
	r := renamer.NewNoOpRenamer(symbols)
	js := js_printer.Print(tree, symbols, r, js_printer.Options{
		OutputFormat: config.FormatPreserve,
		ASCIIOnly:    true,
	}).JS
	return string(js), nil
	// 	// return ast, nil
	// }
}

// Parse a script

func ParseScript(script string) (*js.AST, error) {
	code, err := parseToString(script)
	if err != nil {
		return nil, err
	}
	// Re-parse the code using the tdewolff/parser
	ast, err := js.Parse(parse.NewInputString(code), js.Options{})
	if err != nil {
		return nil, err
	}
	return ast, nil
}

// Parse a script
func ParseScript2(script string) (*js.AST, error) {
	// We use esbuild to parse the JavaScript because it supports Typescript
	// obviously, this is a bit of a hack, but it works for now
	result := esbuild.Transform(script, esbuild.TransformOptions{
		Sourcefile: "script.js",
		Loader:     esbuild.LoaderTS,
		// Format:            esbuild.FormatESModule,
		// TreeShaking:       esbuild.TreeShakingFalse,
		// Platform:          esbuild.PlatformBrowser,
		// Target:            esbuild.ESNext,
		TsconfigRaw: `{ "compilerOptions": { "verbatimModuleSyntax": true } }`,
		// KeepNames:         true,
		// MangleProps:       "",
		// MinifyIdentifiers: false,
	})
	if len(result.Errors) > 0 {
		var errs []error
		for _, err := range result.Errors {
			errs = append(errs, errors.New(err.Text))
		}
		return nil, errors.Join(errs...)
	}
	// Re-parse the code using the tdewolff/parser
	ast, err := js.Parse(parse.NewInputBytes(result.Code), js.Options{})
	if err != nil {
		return nil, err
	}
	return ast, nil
}

// ParseExpr parses a JavaScript expression
func ParseExpr(contents string) (js.IExpr, error) {
	ast, err := js.Parse(parse.NewInputString(contents), js.Options{})
	if err != nil {
		return nil, err
	}
	blockStmt := ast.BlockStmt
	stmts := blockStmt.List
	if len(stmts) != 1 {
		return nil, fmt.Errorf("expected one statement, got %d", len(stmts))
	}
	stmt := stmts[0]
	es, ok := stmt.(*js.ExprStmt)
	if !ok {
		return nil, fmt.Errorf("expected expression statement, got %T", stmt)
	}
	return es.Value, nil
}

// Print a JavaScript AST
func Print(ast js.INode) string {
	return strings.TrimSpace(ast.JS())
}
