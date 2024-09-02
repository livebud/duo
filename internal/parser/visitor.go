package parser

import (
	"fmt"
	"strings"

	"github.com/livebud/duo/internal/js"
	"github.com/livebud/duo/internal/scope"
)

func walk(sc *scope.Scope, node js.INode) error {
	v := &visitor{sc: sc}
	js.Walk(v, node)
	return v.err
}

type visitor struct {
	sc  *scope.Scope
	err error
}

func (v *visitor) Enter(node js.INode) js.IVisitor {
	if err := v.enter(node); err != nil {
		v.err = err
		return nil
	}
	return v
}

func (v *visitor) Exit(node js.INode) {
	if err := v.exit(node); err != nil {
		v.err = err
		return
	}
}

func (v *visitor) enter(n js.INode) error {
	switch n := n.(type) {
	case *js.ImportStmt:
		return v.enterImportStmt(n)
	case *js.ExportStmt:
		return v.enterExportStmt(n)
	case *js.VarDecl:
		return v.enterVarDecl(n)
	case *js.FuncDecl:
		return v.enterFuncDecl(n)
	case *js.Var:
		return v.enterVar(n)
	default:
		return nil
	}
}

func (v *visitor) exit(n js.INode) error {
	switch n := n.(type) {
	case *js.ExportStmt:
		return v.exitExportStmt(n)
	case *js.VarDecl:
		return v.exitVarDecl(n)
	case *js.FuncDecl:
		return v.exitFuncDecl(n)
	default:
		return nil
	}
}

func (v *visitor) enterImportStmt(node *js.ImportStmt) error {
	importPath := strings.Trim(string(node.Module), `"'`)
	if node.Default != nil {
		sym := v.sc.Use(string(node.Default))
		sym.Import = &scope.Import{
			Path:    importPath,
			Default: true,
		}
	}
	if len(node.List) > 0 {
		return fmt.Errorf("parser: walk imported aliases not implemented yet")
	}
	return nil
}

func (v *visitor) enterExportStmt(_ *js.ExportStmt) error {
	v.sc.IsExported = true
	return nil
}

func (v *visitor) exitExportStmt(_ *js.ExportStmt) error {
	v.sc.IsExported = false
	return nil
}

func (v *visitor) enterVarDecl(node *js.VarDecl) error {
	v.sc.IsDeclaration = true
	if node.TokenType == js.VarToken || node.TokenType == js.LetToken {
		v.sc.IsMutable = true
	}
	return nil
}

func (v *visitor) exitVarDecl(node *js.VarDecl) error {
	v.sc.IsDeclaration = false
	if node.TokenType == js.VarToken || node.TokenType == js.LetToken {
		v.sc.IsMutable = false
	}
	return nil
}

func (v *visitor) enterVar(node *js.Var) error {
	name := string(node.Data)
	v.sc.Use(name)
	return nil
}

func (v *visitor) enterFuncDecl(_ *js.FuncDecl) error {
	v.sc.IsDeclaration = true
	v.sc = v.sc.New()
	return nil
}

func (v *visitor) exitFuncDecl(_ *js.FuncDecl) error {
	v.sc.IsDeclaration = false
	v.sc = v.sc.Parent()
	return nil
}
