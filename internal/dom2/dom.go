package dom

import (
	"github.com/livebud/duo/internal/ast"
	"github.com/livebud/duo/internal/js"
	"github.com/livebud/duo/internal/parser"
)

func Generate(path string, code []byte) (string, error) {
	generator := &Generator{}
	return generator.Generate(path, code)
}

type Generator struct {
}

func (g *Generator) Generate(path string, code []byte) (string, error) {
	doc, err := parser.Parse(path, string(code))
	if err != nil {
		return "", err
	}
	return Print(doc)
}

func Print(doc *ast.Document) (string, error) {
	program, err := Transform(doc)
	if err != nil {
		return "", err
	}
	code, err := js.Format(program)
	if err != nil {
		return "", err
	}
	return code, nil
}

func Transform(doc *ast.Document) (*js.AST, error) {
	var stmts []js.IStmt
	return &js.AST{
		BlockStmt: js.BlockStmt{
			List: stmts,
		},
	}, nil
}
