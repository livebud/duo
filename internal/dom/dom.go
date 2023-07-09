package dom

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/livebud/duo/internal/ast"
	"github.com/tdewolff/parse/v2/js"
)

func Generate(doc *ast.Document) (string, error) {
	program, err := Transform(doc)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(program.JS()) + "\n", nil
}

func Transform(doc *ast.Document) (*js.AST, error) {
	s := &script{}
	// Create a new for the script scope
	scope := &js.Scope{}
	hVar, ok := scope.Declare(js.ArgumentDecl, []byte("__h__"))
	if !ok {
		return nil, fmt.Errorf("transform: unable to declare __h__")
	}
	proxyVar, ok := scope.Declare(js.ArgumentDecl, []byte("__proxy__"))
	if !ok {
		return nil, fmt.Errorf("transform: unable to declare __proxy__")
	}
	// Transform the document
	if err := s.transformDocument(scope, doc); err != nil {
		return nil, err
	}
	// Create the program with imports, then default export with render function
	var stmts []js.IStmt
	stmts = append(stmts, s.imports...)
	if len(s.vnodes) > 0 {
		// Create the render function
		// TODO: handle multiple top-level vnodes
		returnStmt, err := toRenderFunction(scope, s.vnodes[0])
		if err != nil {
			return nil, err
		}
		s.stmts = append(s.stmts, returnStmt)
	}
	if len(s.stmts) > 0 {
		// Create the default export
		defaultExport, err := toDefaultExport(scope, hVar, proxyVar, s.stmts...)
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, defaultExport)
	}
	return &js.AST{
		BlockStmt: js.BlockStmt{
			List: stmts,
		},
	}, nil
}

type script struct {
	imports []js.IStmt
	stmts   []js.IStmt
	vnodes  []js.IExpr
}

func (s *script) transformDocument(scope *js.Scope, doc *ast.Document) error {
	for _, child := range doc.Children {
		switch n := child.(type) {
		case *ast.Script:
			if err := s.transformScript(scope, n); err != nil {
				return err
			}
		case *ast.Element:
			vnode, err := generateFragment(scope, n)
			if err != nil {
				return err
			}
			s.vnodes = append(s.vnodes, vnode)
		// Ignore any other types of nodes
		default:
			continue
		}
	}
	return nil
}

func generateFragment(scope *js.Scope, node ast.Fragment) (js.IExpr, error) {
	switch n := node.(type) {
	case *ast.Text:
		return generateText(scope, n)
	case *ast.Element:
		return generateElement(scope, n)
	case *ast.Mustache:
		return generateMustache(scope, n)
	default:
		return nil, fmt.Errorf("unable to generate fragment %T", node)
	}
}

func generateText(scope *js.Scope, node *ast.Text) (*js.LiteralExpr, error) {
	// TODO handle escaping & different types of text
	return &js.LiteralExpr{
		Data:      []byte(strconv.Quote(node.Value)),
		TokenType: js.StringToken,
	}, nil
}

func generateElement(scope *js.Scope, node *ast.Element) (*js.CallExpr, error) {
	// Create the element
	element := createElement(scope, node.Name)

	// Create the attributes
	var attributes []js.Property
	for _, attr := range node.Attributes {
		attribute, err := generateAttribute(scope, attr)
		if err != nil {
			return nil, err
		}
		attributes = append(attributes, attribute)
	}
	element.Args.List = append(element.Args.List, js.Arg{
		Value: &js.ObjectExpr{
			List: attributes,
		},
	})

	// Create the children
	var children []js.Element
	for _, child := range node.Children {
		child, err := generateFragment(scope, child)
		if err != nil {
			return nil, err
		}
		children = append(children, js.Element{
			Value: child,
		})
	}
	element.Args.List = append(element.Args.List, js.Arg{
		Value: &js.ArrayExpr{
			List: children,
		},
	})

	return element, nil
}

func generateMustache(scope *js.Scope, node *ast.Mustache) (js.IExpr, error) {
	switch n := node.Expr.(type) {
	case *js.Var:
		expr := rewriteVar(scope, "__props__", n)
		return expr, nil
	default:
		return nil, fmt.Errorf("unable to generate mustache for %T", n)
	}
}

func generateAttribute(scope *js.Scope, node ast.Attribute) (js.Property, error) {
	switch n := node.(type) {
	case *ast.Field:
		return generateField(scope, n)
	case *ast.AttributeShorthand:
		return generateAttributeShorthand(scope, n)
	default:
		return js.Property{}, fmt.Errorf("unable to generate attribute %T", node)
	}
}

func generateField(scope *js.Scope, node *ast.Field) (js.Property, error) {
	values, err := generateValues(scope, node.Values)
	if err != nil {
		return js.Property{}, err
	}
	return js.Property{
		Name: &js.PropertyName{
			Literal: toIdentifier([]byte(node.Key)),
		},
		Value: concat(values),
	}, nil
}

func generateAttributeShorthand(scope *js.Scope, node *ast.AttributeShorthand) (js.Property, error) {
	return js.Property{
		Name: &js.PropertyName{
			Literal: toIdentifier([]byte(node.Key)),
		},
		Value: rewriteVar(scope, "__props__", &js.Var{
			Data: []byte(node.Key),
		}),
	}, nil
}

func generateValues(scope *js.Scope, values []ast.Value) ([]js.IExpr, error) {
	var exprs []js.IExpr
	for _, value := range values {
		expr, err := generateValue(scope, value)
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, expr)
	}
	return exprs, nil
}

func generateValue(scope *js.Scope, value ast.Value) (js.IExpr, error) {
	switch n := value.(type) {
	case *ast.Text:
		return generateText(scope, n)
	case *ast.Mustache:
		return generateMustache(scope, n)
	default:
		return nil, fmt.Errorf("unable to generate value %T", value)
	}
}

// Create `h('h1', { ... }, [ ... ])`
func createElement(scope *js.Scope, name string) *js.CallExpr {
	return &js.CallExpr{
		X: scope.Use([]byte("__h__")),
		Args: js.Args{
			List: []js.Arg{
				{
					Value: js.LiteralExpr{
						Data:      []byte(strconv.Quote(name)),
						TokenType: js.StringToken,
					},
				},
			},
		},
	}
}

func concat(values []js.IExpr) js.IExpr {
	if len(values) == 0 {
		return &js.LiteralExpr{
			Data:      []byte(""),
			TokenType: js.StringToken,
		}
	} else if len(values) == 1 {
		return values[0]
	}
	left := concat(values[:len(values)-1])
	right := values[len(values)-1]
	return &js.BinaryExpr{
		X:  left,
		Op: js.AddToken,
		Y:  right,
	}
}

func (s *script) transformScript(scope *js.Scope, node *ast.Script) error {
	return s.transformTopLevelBlockStmt(scope, &node.Program.BlockStmt)
}

func (s *script) transformTopLevelBlockStmt(scope *js.Scope, node *js.BlockStmt) error {
	node.Scope.Parent = scope
	for _, stmt := range node.List {
		if err := s.transformTopLevelStmt(&node.Scope, stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s *script) transformTopLevelStmt(scope *js.Scope, node js.IStmt) error {
	switch stmt := node.(type) {
	case *js.ImportStmt:
		s.imports = append(s.imports, stmt)
		return nil
	case *js.ExportStmt:
		return s.transformExportStmt(scope, stmt)
	default:
		if err := transformStmt(scope, stmt); err != nil {
			return err
		}
		s.stmts = append(s.stmts, stmt)
		return nil
	}
}

func transformBlockStmt(scope *js.Scope, node *js.BlockStmt) error {
	node.Scope.Parent = scope
	for _, stmt := range node.List {
		if err := transformStmt(&node.Scope, stmt); err != nil {
			return err
		}
	}
	return nil
}

func transformStmt(scope *js.Scope, node js.IStmt) error {
	switch stmt := node.(type) {
	case *js.ExprStmt:
		return transformExprStmt(scope, stmt)
	default:
		return fmt.Errorf("unknown statement %T", stmt)
	}
}

func (s *script) transformExportStmt(scope *js.Scope, node *js.ExportStmt) error {
	switch decl := node.Decl.(type) {
	case *js.VarDecl:
		return s.transformExportVarDecl(scope, decl)
	default:
		return fmt.Errorf("unknown declaration %T", node.Decl)
	}
}

func (s *script) transformExportVarDecl(scope *js.Scope, node *js.VarDecl) error {
	// Ignore any exports that are not let or var
	if node.TokenType != js.LetToken && node.TokenType != js.VarToken {
		return nil
	}
	for _, decl := range node.List {
		if err := s.transformExportBindingElement(scope, &decl); err != nil {
			return err
		}
	}
	return nil
}

func (s *script) transformExportBindingElement(scope *js.Scope, node *js.BindingElement) error {
	// Spreads and empty lets are ignored (already passed in above)
	if node.Binding == nil || node.Default == nil {
		return nil
	}
	letVar, ok := node.Binding.(*js.Var)
	if !ok {
		// Ignore destructured exports
		return nil
	}
	// Rewrite `export let x = ...` to `__proxy__.x = __proxy__.x || ...`
	dotExpr := rewriteVar(scope, "__proxy__", letVar)
	s.stmts = append(s.stmts, exprStmt(
		assignExpr(
			dotExpr,
			orExpr(dotExpr, node.Default),
		),
	))
	return nil
}

func transformExprStmt(scope *js.Scope, node *js.ExprStmt) error {
	return transformExpr(scope, node.Value)
}

func transformCallExpr(scope *js.Scope, node *js.CallExpr) error {
	if x, ok := node.X.(*js.Var); ok {
		node.X = rewriteVar(scope, "__proxy__", x)
	}
	for _, arg := range node.Args.List {
		if err := transformArg(scope, arg); err != nil {
			return err
		}
	}
	return nil
}

func transformArg(scope *js.Scope, arg js.Arg) error {
	return transformExpr(scope, arg.Value)
}

func transformExpr(scope *js.Scope, node js.IExpr) error {
	switch expr := node.(type) {
	case *js.Var:
		return transformVar(scope, expr)
	case *js.CallExpr:
		return transformCallExpr(scope, expr)
	case *js.ArrowFunc:
		return transformArrowFunc(scope, expr)
	case *js.BinaryExpr:
		return transformBinaryExpr(scope, expr)
	case *js.LiteralExpr:
		return transformLiteralExpr(scope, expr)
	default:
		return fmt.Errorf("unknown expression %T", expr)
	}
}

func transformVar(scope *js.Scope, node *js.Var) error {
	return nil
}

func transformArrowFunc(scope *js.Scope, node *js.ArrowFunc) error {
	if err := transformBlockStmt(scope, &node.Body); err != nil {
		return err
	}
	return nil
}

func transformLiteralExpr(scope *js.Scope, node *js.LiteralExpr) error {
	return nil
}

func transformBinaryExpr(scope *js.Scope, node *js.BinaryExpr) error {
	if x, ok := node.X.(*js.Var); ok {
		node.X = rewriteVar(scope, "__proxy__", x)
	}
	return nil
}

var globals = map[string]bool{
	"window":                    true,
	"document":                  true,
	"console":                   true,
	"localStorage":              true,
	"sessionStorage":            true,
	"navigator":                 true,
	"location":                  true,
	"XMLHttpRequest":            true,
	"setTimeout":                true,
	"clearTimeout":              true,
	"setInterval":               true,
	"clearInterval":             true,
	"requestAnimationFrame":     true,
	"cancelAnimationFrame":      true,
	"fetch":                     true,
	"atob":                      true,
	"btoa":                      true,
	"FormData":                  true,
	"URL":                       true,
	"URLSearchParams":           true,
	"Headers":                   true,
	"AbortController":           true,
	"Event":                     true,
	"CustomEvent":               true,
	"MouseEvent":                true,
	"KeyboardEvent":             true,
	"FocusEvent":                true,
	"TouchEvent":                true,
	"IntersectionObserver":      true,
	"IntersectionObserverEntry": true,
	"MutationObserver":          true,
	"MutationRecord":            true,
	"ResizeObserver":            true,
	"ResizeObserverEntry":       true,
	"Promise":                   true,
	"Symbol":                    true,
	"Map":                       true,
	"Set":                       true,
	"WeakMap":                   true,
	"WeakSet":                   true,
	"Intl":                      true,
	"Intl.DateTimeFormat":       true,
	"Intl.NumberFormat":         true,
	"Object":                    true,
	"Array":                     true,
	"Function":                  true,
	"Boolean":                   true,
	"Error":                     true,
	"EvalError":                 true,
	"RangeError":                true,
	"ReferenceError":            true,
	"SyntaxError":               true,
	"TypeError":                 true,
	"URIError":                  true,
	"Number":                    true,
	"BigInt":                    true,
	"Math":                      true,
	"Date":                      true,
	"RegExp":                    true,
	"String":                    true,
	"JSON":                      true,
	"parseFloat":                true,
	"parseInt":                  true,
	"isNaN":                     true,
	"isFinite":                  true,
	"decodeURI":                 true,
	"decodeURIComponent":        true,
	"encodeURI":                 true,
	"encodeURIComponent":        true,
	"eval":                      true,
	"NaN":                       true,
	"Infinity":                  true,
}

func rewriteVar(scope *js.Scope, name string, v *js.Var) js.IExpr {
	// TODO: handle scoping better
	if globals[string(v.Data)] {
		return v
	}
	return &js.DotExpr{
		X: scope.Use([]byte(name)),
		Y: toIdentifier(v.Data),
	}
}

func toDefaultExport(scope *js.Scope, hVar, proxyVar *js.Var, body ...js.IStmt) (*js.ExportStmt, error) {
	return &js.ExportStmt{
		Default: true,
		Decl: &js.FuncDecl{
			Params: js.Params{
				List: []js.BindingElement{
					{
						Binding: hVar,
					},
					{
						Binding: proxyVar,
					},
				},
			},
			Body: js.BlockStmt{
				Scope: js.Scope{
					Parent: scope,
				},
				List: body,
			},
		},
	}, nil
}

func toRenderFunction(scope *js.Scope, expr js.IExpr) (js.IStmt, error) {
	scope = &js.Scope{
		Parent: scope,
	}
	propsVar, ok := scope.Declare(js.ArgumentDecl, []byte("__props__"))
	if !ok {
		return nil, fmt.Errorf("transform: unable to declare __props__")
	}
	return &js.ReturnStmt{
		Value: &js.ArrowFunc{
			Params: js.Params{
				List: []js.BindingElement{
					{
						Binding: propsVar,
					},
				},
			},
			Body: js.BlockStmt{
				Scope: js.Scope{
					Parent: scope,
				},
				List: []js.IStmt{
					&js.ReturnStmt{
						Value: expr,
					},
				},
			},
		},
	}, nil
}

func exprStmt(expr js.IExpr) *js.ExprStmt {
	return &js.ExprStmt{
		Value: expr,
	}
}

func assignExpr(x js.IExpr, y js.IExpr) *js.BinaryExpr {
	return &js.BinaryExpr{
		X:  x,
		Op: js.EqToken,
		Y:  y,
	}
}

func orExpr(x js.IExpr, y js.IExpr) *js.BinaryExpr {
	return &js.BinaryExpr{
		X:  x,
		Op: js.OrToken,
		Y:  y,
	}
}

func toIdentifier(name []byte) js.LiteralExpr {
	return js.LiteralExpr{
		Data:      name,
		TokenType: js.IdentifierToken,
	}
}
