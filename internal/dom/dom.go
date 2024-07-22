package dom

import (
	"fmt"
	"strconv"

	"github.com/livebud/duo/internal/ast"
	"github.com/livebud/duo/internal/js"
	"github.com/livebud/duo/internal/scope"
)

func Generate(doc *ast.Document) (string, error) {
	program, err := Transform(doc)
	if err != nil {
		return "", err
	}
	code, err := js.PrettyPrint(program.JS())
	if err != nil {
		return "", err
	}
	return code, nil
}

func Transform(doc *ast.Document) (*js.AST, error) {
	s := &script{
		scope: doc.Scope,
	}
	// New scope modification
	scope := scope.New()
	// Transform the document
	if err := s.transformDocument(scope, doc); err != nil {
		return nil, err
	}
	// Create the program with imports, then default export with render function
	var stmts []js.IStmt
	stmts = append(stmts, s.imports...)
	if s.render != nil {
		s.stmts = append(s.stmts, s.render)
	}
	if len(s.stmts) > 0 {
		// Create the default export
		defaultExport, err := toDefaultExport(scope, s.stmts...)
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
	scope    *scope.Scope
	imports  []js.IStmt
	stmts    []js.IStmt
	render   js.IStmt
	inScript bool // Set while traversing the script
}

// TODO: I think we can can clean this up a bunch, basically look for the script
// first and transform that. Then introduce a new scope with props and transform
// the fragments.
func (s *script) transformDocument(scope *scope.Scope, doc *ast.Document) error {
	hName := doc.Scope.FindFree("h")
	proxyName := doc.Scope.FindFree("proxy")
	if _, err := scope.Declare("h", hName); err != nil {
		return err
	}
	if _, err := scope.Declare("proxy", proxyName); err != nil {
		return err
	}
	for _, child := range doc.Children {
		switch n := child.(type) {
		case *ast.Script:
			s.inScript = true
			if err := s.transformScript(scope, n); err != nil {
				s.inScript = false
				return err
			}
			s.inScript = false
		case *ast.Element:
			childScope := scope.New()
			propsName := childScope.FindFree("props")
			if _, err := childScope.Declare("props", propsName); err != nil {
				return err
			}
			vnode, err := s.generateFragment(childScope, n)
			if err != nil {
				return err
			}
			s.render, err = s.toRenderFunction(childScope, vnode)
			if err != nil {
				return err
			}
		// Ignore any other types of nodes
		default:
			continue
		}
	}
	return nil
}

func (s *script) generateFragment(scope *scope.Scope, node ast.Fragment) (js.IExpr, error) {
	switch n := node.(type) {
	case *ast.Text:
		return s.generateText(scope, n)
	case *ast.Element:
		return s.generateElement(scope, n)
	case *ast.Mustache:
		return s.generateMustache(scope, n)
	default:
		return nil, fmt.Errorf("unable to generate fragment %T", node)
	}
}

func (s *script) generateText(_ *scope.Scope, node *ast.Text) (*js.LiteralExpr, error) {
	// TODO handle escaping & different types of text
	return &js.LiteralExpr{
		Data:      []byte(strconv.Quote(node.Value)),
		TokenType: js.StringToken,
	}, nil
}

func (s *script) generateElement(scope *scope.Scope, node *ast.Element) (*js.CallExpr, error) {
	// Create the element
	element, err := createElement(scope, node.Name)
	if err != nil {
		return nil, err
	}
	// Create the attributes
	var attributes []js.Property
	for _, attr := range node.Attributes {
		attribute, err := s.generateAttribute(scope, attr)
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
		child, err := s.generateFragment(scope, child)
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

func (s *script) generateMustache(scope *scope.Scope, node *ast.Mustache) (js.IExpr, error) {
	return s.generateExpr(scope, node.Expr)
}

func (s *script) generateExpr(scope *scope.Scope, node js.IExpr) (js.IExpr, error) {
	switch n := node.(type) {
	case *js.LiteralExpr:
		return generateLiteralExpr(scope, n)
	case *js.Var:
		return s.rewriteVar(scope, n)
	case *js.CondExpr:
		return s.generateCondExpr(scope, n)
	case *js.BinaryExpr:
		return s.generateBinaryExpr(scope, n)
	default:
		return nil, fmt.Errorf("unable to generate expression for %T", n)
	}
}

func generateLiteralExpr(_ *scope.Scope, node *js.LiteralExpr) (js.IExpr, error) {
	return node, nil
}

func (s *script) generateCondExpr(scope *scope.Scope, node *js.CondExpr) (js.IExpr, error) {
	cond, err := s.generateExpr(scope, node.Cond)
	if err != nil {
		return nil, err
	}
	x, err := s.generateExpr(scope, node.X)
	if err != nil {
		return nil, err
	}
	y, err := s.generateExpr(scope, node.Y)
	if err != nil {
		return nil, err
	}
	return &js.CondExpr{
		Cond: cond,
		X:    x,
		Y:    y,
	}, nil
}

func (s *script) generateBinaryExpr(scope *scope.Scope, node *js.BinaryExpr) (js.IExpr, error) {
	x, err := s.generateExpr(scope, node.X)
	if err != nil {
		return nil, err
	}
	y, err := s.generateExpr(scope, node.Y)
	if err != nil {
		return nil, err
	}
	return &js.BinaryExpr{
		X:  x,
		Op: node.Op,
		Y:  y,
	}, nil
}

func (s *script) generateAttribute(scope *scope.Scope, node ast.Attribute) (js.Property, error) {
	switch n := node.(type) {
	case *ast.Field:
		return s.generateField(scope, n)
	case *ast.AttributeShorthand:
		return s.generateAttributeShorthand(scope, n)
	default:
		return js.Property{}, fmt.Errorf("unable to generate attribute %T", node)
	}
}

func (s *script) generateField(scope *scope.Scope, node *ast.Field) (js.Property, error) {
	values, err := s.generateValues(scope, node.Values)
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

func (s *script) generateAttributeShorthand(scope *scope.Scope, node *ast.AttributeShorthand) (js.Property, error) {
	value, err := s.rewriteVar(scope, &js.Var{
		Data: []byte(node.Key),
	})
	if err != nil {
		return js.Property{}, err
	}
	return js.Property{
		Name: &js.PropertyName{
			Literal: toIdentifier([]byte(node.Key)),
		},
		Value: value,
	}, nil
}

func (s *script) generateValues(scope *scope.Scope, values []ast.Value) ([]js.IExpr, error) {
	var exprs []js.IExpr
	for _, value := range values {
		expr, err := s.generateValue(scope, value)
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, expr)
	}
	return exprs, nil
}

func (s *script) generateValue(scope *scope.Scope, value ast.Value) (js.IExpr, error) {
	switch n := value.(type) {
	case *ast.Text:
		return s.generateText(scope, n)
	case *ast.Mustache:
		return s.generateMustache(scope, n)
	default:
		return nil, fmt.Errorf("unable to generate value %T", value)
	}
}

// Create `h('h1', { ... }, [ ... ])`
func createElement(scope *scope.Scope, name string) (*js.CallExpr, error) {
	h, err := scope.LookupByID("h")
	if !err {
		return nil, fmt.Errorf("transform: unable to lookup h in scope")
	}
	return &js.CallExpr{
		X: h.ToVar(),
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
	}, nil
}

func concat(values []js.IExpr) js.IExpr {
	if len(values) == 0 {
		return &js.LiteralExpr{
			Data:      []byte{'"', '"'},
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

func (s *script) transformScript(scope *scope.Scope, node *ast.Script) error {
	return s.transformTopLevelBlockStmt(scope, &node.Program.BlockStmt)
}

func (s *script) transformTopLevelBlockStmt(scope *scope.Scope, node *js.BlockStmt) error {
	for _, stmt := range node.List {
		if err := s.transformTopLevelStmt(scope, stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s *script) transformTopLevelStmt(scope *scope.Scope, node js.IStmt) error {
	switch stmt := node.(type) {
	case *js.ImportStmt:
		s.imports = append(s.imports, stmt)
		return nil
	case *js.ExportStmt:
		return s.transformExportStmt(scope, stmt)
	case *js.VarDecl:
		return s.transformExportVarDecl(scope, stmt)
	default:
		if err := s.transformStmt(scope, stmt); err != nil {
			return err
		}
		s.stmts = append(s.stmts, stmt)
		return nil
	}
}

func (s *script) transformBlockStmt(scope *scope.Scope, node *js.BlockStmt) error {
	for _, stmt := range node.List {
		if err := s.transformStmt(scope, stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s *script) transformStmt(scope *scope.Scope, node js.IStmt) error {
	switch stmt := node.(type) {
	case *js.ExprStmt:
		return s.transformExprStmt(scope, stmt)
	case *js.FuncDecl:
		return s.transformFuncDecl(scope, stmt)
	default:
		return fmt.Errorf("unknown statement %T", stmt)
	}
}

func (s *script) transformExportStmt(scope *scope.Scope, node *js.ExportStmt) error {
	switch decl := node.Decl.(type) {
	case *js.VarDecl:
		return s.transformExportVarDecl(scope, decl)
	default:
		return fmt.Errorf("unknown declaration %T", node.Decl)
	}
}

func (s *script) transformExportVarDecl(scope *scope.Scope, node *js.VarDecl) error {
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

func (s *script) transformExportBindingElement(scope *scope.Scope, node *js.BindingElement) error {
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
	dotExpr, err := s.rewriteVar(scope, letVar)
	if err != nil {
		return err
	}
	s.stmts = append(s.stmts, exprStmt(
		assignExpr(
			dotExpr,
			orExpr(dotExpr, node.Default),
		),
	))
	return nil
}

func (s *script) transformExprStmt(scope *scope.Scope, node *js.ExprStmt) error {
	return s.transformExpr(scope, node.Value)
}

func (s *script) transformFuncDecl(scope *scope.Scope, node *js.FuncDecl) error {
	// TODO: this should be js.FunctionDecl, but that panics
	// if _, ok := scope.Declare(js.NoDecl, node.Name.Data); !ok {
	// 	return fmt.Errorf("transform: unable to declare %s", node.Name.Data)
	// }
	return s.transformBlockStmt(scope, &node.Body)
}

func (s *script) transformCallExpr(scope *scope.Scope, node *js.CallExpr) error {
	if variable, ok := node.X.(*js.Var); ok {
		x, err := s.rewriteVar(scope, variable)
		if err != nil {
			return err
		}
		node.X = x
	}
	for _, arg := range node.Args.List {
		if err := s.transformArg(scope, arg); err != nil {
			return err
		}
	}
	return nil
}

func (s *script) transformArg(scope *scope.Scope, arg js.Arg) error {
	return s.transformExpr(scope, arg.Value)
}

func (s *script) transformExpr(scope *scope.Scope, node js.IExpr) error {
	switch expr := node.(type) {
	case *js.Var:
		return transformVar(scope, expr)
	case *js.CallExpr:
		return s.transformCallExpr(scope, expr)
	case *js.ArrowFunc:
		return s.transformArrowFunc(scope, expr)
	case *js.BinaryExpr:
		return s.transformBinaryExpr(scope, expr)
	case *js.LiteralExpr:
		return s.transformLiteralExpr(scope, expr)
	default:
		return fmt.Errorf("unknown expression %T", expr)
	}
}

func transformVar(_ *scope.Scope, _ *js.Var) error {
	return nil
}

func (s *script) transformArrowFunc(scope *scope.Scope, node *js.ArrowFunc) error {
	if err := s.transformBlockStmt(scope, &node.Body); err != nil {
		return err
	}
	return nil
}

func (s *script) transformLiteralExpr(_ *scope.Scope, _ *js.LiteralExpr) error {
	return nil
}

func (s *script) transformBinaryExpr(scope *scope.Scope, node *js.BinaryExpr) error {
	if variable, ok := node.X.(*js.Var); ok {
		x, err := s.rewriteVar(scope, variable)
		if err != nil {
			return err
		}
		node.X = x
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

func (s *script) rewriteVar(scope *scope.Scope, v *js.Var) (js.IExpr, error) {
	// TODO: Handle global scoping better
	if globals[string(v.Data)] {
		return v, nil
	}
	// Find the symbol in the scope
	sym, ok := s.scope.LookupByName(string(v.Data))
	if !ok {
		return nil, fmt.Errorf("transform: unable to find symbol %s", v.Data)
	}
	// Mutable variables in the script are re-written as proxy properties
	if s.inScript && sym.IsMutable() {
		proxy, err := scope.LookupByID("proxy")
		if !err {
			return nil, fmt.Errorf("transform: unable to find proxy in scope")
		}
		return &js.DotExpr{
			X: proxy.ToVar(),
			Y: toIdentifier(v.Data),
		}, nil
	}
	// Mutable or undeclared variables in the template are re-written as props
	// properties
	if !s.inScript && (sym.IsMutable() || !sym.IsDeclared()) {
		props, err := scope.LookupByID("props")
		if !err {
			return nil, fmt.Errorf("transform: unable to find props in scope to rewrite %q", v.Data)
		}
		return &js.DotExpr{
			X: props.ToVar(),
			Y: toIdentifier(v.Data),
		}, nil
	}
	return v, nil
}

func toDefaultExport(scope *scope.Scope, body ...js.IStmt) (*js.ExportStmt, error) {
	h, ok := scope.LookupByID("h")
	if !ok {
		return nil, fmt.Errorf("transform: unable to lookup h in scope")
	}
	proxy, ok := scope.LookupByID("proxy")
	if !ok {
		return nil, fmt.Errorf("transform: unable to find proxy in scope")
	}
	return &js.ExportStmt{
		Default: true,
		Decl: &js.FuncDecl{
			Params: js.Params{
				List: []js.BindingElement{
					{
						Binding: h.ToVar(),
					},
					{
						Binding: proxy.ToVar(),
					},
				},
			},
			Body: js.BlockStmt{
				Scope: js.Scope{
					Parent: &js.Scope{},
				},
				List: body,
			},
		},
	}, nil
}

func (s *script) toRenderFunction(scope *scope.Scope, expr js.IExpr) (js.IStmt, error) {
	props, ok := scope.LookupByID("props")
	if !ok {
		return nil, fmt.Errorf("transform: unable to find props in scope")
	}
	return &js.ReturnStmt{
		Value: &js.ArrowFunc{
			Params: js.Params{
				List: []js.BindingElement{
					{
						Binding: props.ToVar(),
					},
				},
			},
			Body: js.BlockStmt{
				Scope: js.Scope{
					Parent: &js.Scope{},
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
