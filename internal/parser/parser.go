package parser

import (
	"errors"
	"fmt"

	"github.com/livebud/duo/internal/ast"
	"github.com/livebud/duo/internal/event"
	"github.com/livebud/duo/internal/lexer"
	"github.com/livebud/duo/internal/scope"
	"github.com/livebud/duo/internal/token"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/js"
)

func Parse(input string) (*ast.Document, error) {
	l := lexer.New(input)
	p := New(l)
	return p.Parse()
}

func Print(input string) string {
	doc, err := Parse(input)
	if err != nil {
		return err.Error()
	}
	return doc.String()
}

func New(l *lexer.Lexer) *Parser {
	return &Parser{l, scope.New()}
}

type Parser struct {
	l  *lexer.Lexer
	sc *scope.Scope
}

func (p *Parser) Parse() (*ast.Document, error) {
	return p.parseDocument()
}

func (p *Parser) unexpected(prefix string) error {
	return fmt.Errorf("parser: %s unexpected token %s (%d:%d)", prefix, p.l.Token.String(), p.l.Token.Line, p.l.Token.Start)
}

func (p *Parser) parseDocument() (*ast.Document, error) {
	doc := &ast.Document{
		Scope: p.sc,
	}
	for p.l.Next() {
		child, err := p.parseFragment()
		if err != nil {
			return nil, err
		}
		doc.Children = append(doc.Children, child)
	}
	return doc, nil
}

func (p *Parser) parseFragment() (ast.Fragment, error) {
	switch p.l.Token.Type {
	case token.Text:
		return p.parseText()
	case token.LessThan:
		return p.parseTag()
	case token.LeftBrace:
		return p.parseBlockMustache()
	default:
		return nil, p.unexpected("fragment")
	}
}

func (p *Parser) parseText() (*ast.Text, error) {
	return &ast.Text{
		Value: p.l.Token.Text,
	}, nil
}

func (p *Parser) parseTag() (ast.Fragment, error) {
	p.l.Next()
	switch p.l.Token.Type {
	case token.Identifier:
		return p.parseElement()
	case token.Script:
		return p.parseScript()
	default:
		return nil, p.unexpected("tag")
	}
}

func (p *Parser) parseElement() (*ast.Element, error) {
	node := new(ast.Element)
	node.Name = p.l.Token.Text

	// Opening tag
openTag:
	for p.l.Next() {
		switch p.l.Token.Type {
		case token.SlashGreaterThan:
			node.SelfClosing = true
			return node, nil
		case token.GreaterThan:
			break openTag
		case token.Identifier:
			attr, err := p.parseAttribute()
			if err != nil {
				return nil, err
			}
			node.Attributes = append(node.Attributes, attr)
			// End of tag
			if p.l.Token.Type == token.GreaterThan {
				break openTag
			} else if p.l.Token.Type == token.SlashGreaterThan {
				node.SelfClosing = true
				return node, nil
			}
		case token.LeftBrace:
			attr, err := p.parseAttributeShorthand()
			if err != nil {
				return nil, err
			}
			node.Attributes = append(node.Attributes, attr)
			// End of tag
			if p.l.Token.Type == token.GreaterThan {
				break openTag
			} else if p.l.Token.Type == token.SlashGreaterThan {
				node.SelfClosing = true
				return node, nil
			}
		default:
			return nil, p.unexpected("element")
		}
	}

	// Add any children
	for p.l.Next() && p.l.Token.Type != token.LessThanSlash {
		child, err := p.parseFragment()
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, child)

	}

	if p.l.Token.Type != token.LessThanSlash {
		return nil, p.unexpected("element")
	}
	p.l.Next()
	if p.l.Token.Type != token.Identifier {
		return nil, p.unexpected("element")
	} else if p.l.Token.Text != node.Name {
		return nil, fmt.Errorf("expected closing tag %s, got %s", node.Name, p.l.Token.Text)
	}
	p.l.Next()
	if p.l.Token.Type != token.GreaterThan {
		return nil, p.unexpected("element")
	}
	return node, nil
}

func (p *Parser) parseAttribute() (ast.Attribute, error) {
	key := p.l.Token.Text
	return p.parseField(key)
}

func (p *Parser) parseField(key string) (*ast.Field, error) {
	field := new(ast.Field)
	field.Key = key
	field.EventHandler = event.Is(key)
	for p.l.Next() {
		switch p.l.Token.Type {
		case token.Equal:
			if field.EventHandler {
				value, err := p.parseEventValue()
				if err != nil {
					return nil, err
				}
				field.Values = append(field.Values, value)
				return field, nil
			}
			values, err := p.parseAttributeValues()
			if err != nil {
				return nil, err
			}
			field.Values = values
			return field, nil
		default:
			return nil, p.unexpected("field")
		}
	}
	return field, nil
}

func (p *Parser) parseAttributeValues() (values []ast.Value, err error) {
	for p.l.Next() {
		switch p.l.Token.Type {
		case token.SlashGreaterThan, token.GreaterThan:
			return values, nil
		case token.Quote:
			return p.parseAttributeStringValues()
		case token.Text:
			values = append(values, &ast.Text{
				Value: p.l.Token.Text,
			})
		case token.LeftBrace:
			mustache, err := p.parseMustache()
			if err != nil {
				return nil, err
			}
			values = append(values, mustache)
		default:
			return nil, p.unexpected("attribute value")
		}
	}
	return values, nil
}

func (p *Parser) parseAttributeStringValues() (values []ast.Value, err error) {
	for p.l.Next() {
		switch p.l.Token.Type {
		case token.Quote:
			return values, nil
		case token.Text:
			values = append(values, &ast.Text{
				Value: p.l.Token.Text,
			})
		case token.LeftBrace:
			mustache, err := p.parseMustache()
			if err != nil {
				return nil, err
			}
			values = append(values, mustache)
		default:
			return nil, p.unexpected("attribute value")
		}
	}
	return nil, p.unexpected("attribute value")
}

func (p *Parser) parseEventValue() (value ast.Value, err error) {
	for p.l.Next() {
		switch p.l.Token.Type {
		case token.LeftBrace:
			return p.parseMustache()
		default:
			return nil, p.unexpected("event value")
		}
	}
	return nil, p.unexpected("event value")
}

var errDoneBlock = errors.New("done block")

func (p *Parser) parseBlockMustache() (ast.Fragment, error) {
	for p.l.Next() {
		switch p.l.Token.Type {
		case token.Expr:
			node := new(ast.Mustache)
			expr, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			node.Expr = expr
			p.l.Next()
			if p.l.Token.Type != token.RightBrace {
				return nil, p.unexpected("mustache")
			}
			return node, nil
		case token.If:
			p.l.Next()
			return p.parseIfBlock(true)
		default:
			return nil, p.unexpected("mustache")
		}
	}
	return nil, p.unexpected("mustache")
}

func (p *Parser) parseMustache() (*ast.Mustache, error) {
	node := new(ast.Mustache)
	p.l.Next()
	if p.l.Token.Type != token.Expr {
		return nil, p.unexpected("mustache")
	}
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	node.Expr = expr
	p.l.Next()
	if p.l.Token.Type != token.RightBrace {
		return nil, p.unexpected("mustache")
	}
	return node, nil
}

func (p *Parser) parseIfBlock(parseEnd bool) (*ast.IfBlock, error) {
	node := new(ast.IfBlock)
	if p.l.Token.Type != token.Expr {
		return nil, p.unexpected("if block")
	}
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	node.Cond = expr
	p.l.Next()
	if p.l.Token.Type != token.RightBrace {
		return nil, p.unexpected("if block")
	}
	for p.l.Next() {
		switch p.l.Token.Type {
		case token.LeftBrace:
			p.l.Next()
			switch p.l.Token.Type {
			case token.Expr:
				mustache := new(ast.Mustache)
				expr, err := p.parseExpression()
				if err != nil {
					return nil, err
				}
				mustache.Expr = expr
				p.l.Next()
				if p.l.Token.Type != token.RightBrace {
					return nil, p.unexpected("if block")
				}
				node.Then = append(node.Then, mustache)
			case token.End:
				if parseEnd {
					if err := p.parseEnd(); err != nil {
						return nil, err
					}
				}
				return node, nil
			case token.If:
				p.l.Next()
				ifBlock, err := p.parseIfBlock(true)
				if err != nil {
					return nil, err
				}
				node.Then = append(node.Then, ifBlock)
			case token.ElseIf:
				p.l.Next()
				ifBlock, err := p.parseIfBlock(false)
				if err != nil {
					return nil, err
				}
				node.Else = append(node.Else, ifBlock)
			case token.Else:
				p.l.Next()
				if p.l.Token.Type != token.RightBrace {
					return nil, p.unexpected("if block")
				}
				p.l.Next()
				fragment, err := p.parseFragment()
				if err != nil {
					if err != errDoneBlock {
						return nil, err
					}
					return node, nil
				}
				node.Else = append(node.Else, fragment)
			default:
				return nil, p.unexpected("if block")
			}
		// This is meant to be the right brace of an {end}
		// TODO: simplify
		case token.RightBrace:
			p.l.Next()
			return node, nil
		default:
			fragment, err := p.parseFragment()
			if err != nil {
				if err != errDoneBlock {
					return nil, err
				}
				return node, nil
			}
			node.Then = append(node.Then, fragment)
		}
	}
	return nil, fmt.Errorf("unclosed if block")
}

func (p *Parser) parseEnd() error {
	p.l.Next()
	if p.l.Token.Type != token.RightBrace {
		return p.unexpected("end")
	}
	return nil
}

func (p *Parser) parseAttributeShorthand() (ast.Attribute, error) {
	node := new(ast.AttributeShorthand)
	p.l.Next()
	if p.l.Token.Type != token.Expr {
		return nil, p.unexpected("attribute shorthand")
	}
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	ident, ok := expr.(*js.Var)
	if !ok {
		return nil, fmt.Errorf("expected and identifier, got %T", expr)
	}
	name := string(ident.Data)
	node.Key = name
	node.EventHandler = event.Is(name)
	p.l.Next()
	if p.l.Token.Type != token.RightBrace {
		return nil, p.unexpected("attribute shorthand")
	}
	return node, nil
}

var options = js.Options{}

func (p *Parser) parseExpression() (js.IExpr, error) {
	ast, err := js.Parse(parse.NewInputString(p.l.Token.Text), options)
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
	// Walk the program to update scope
	if err := p.walkBlockStatement(p.sc, blockStmt); err != nil {
		return nil, err
	}
	return es.Value, nil
}

func (p *Parser) parseScript() (*ast.Script, error) {
	node := new(ast.Script)
	node.Name = p.l.Token.Text

	// Opening tag
openTag:
	for p.l.Next() {
		switch p.l.Token.Type {
		case token.SlashGreaterThan:
			return node, nil
		case token.GreaterThan:
			break openTag
		// TODO: handle attributes
		default:
			return nil, p.unexpected("script")
		}
	}

	// Parse the program
	p.l.Next()
	program, err := js.Parse(parse.NewInputString(p.l.Token.Text), options)
	if err != nil {
		return nil, err
	}
	// Walk the program to update scope
	if err := p.walkBlockStatement(p.sc, program.BlockStmt); err != nil {
		return nil, err
	}
	node.Program = program
	p.l.Next()

	// Closing tag
	if p.l.Token.Type != token.LessThanSlash {
		return nil, p.unexpected("script")
	}
	p.l.Next()
	if p.l.Token.Type != token.Script {
		return nil, p.unexpected("script")
	}
	if p.l.Token.Text != node.Name {
		return nil, fmt.Errorf("expected closing tag %s, got %s", node.Name, p.l.Token.Text)
	}
	p.l.Next()
	if p.l.Token.Type != token.GreaterThan {
		return nil, p.unexpected("script")
	}
	return node, nil
}

func (p *Parser) walkBlockStatement(sc *scope.Scope, node js.BlockStmt) error {
	for _, stmt := range node.List {
		if err := p.walkStmt(sc, stmt); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) walkStmt(sc *scope.Scope, node js.IStmt) error {
	switch stmt := node.(type) {
	case *js.VarDecl:
		return p.walkVarDecl(sc, stmt)
	case *js.ExportStmt:
		return p.walkExportStmt(sc, stmt)
	case *js.ExprStmt:
		return p.walkExprStmt(sc, stmt)
	case *js.FuncDecl:
		return p.walkFuncDecl(sc, stmt)
	case *js.ReturnStmt:
		return p.walkReturnStmt(sc, stmt)
	default:
		return fmt.Errorf("parser: unexpected statement %T", stmt)
	}
}

func (p *Parser) walkExprStmt(sc *scope.Scope, node *js.ExprStmt) error {
	return p.walkExpr(sc, node.Value)
}

func (p *Parser) walkExportStmt(sc *scope.Scope, node *js.ExportStmt) error {
	sc.IsExported = true
	defer func() {
		sc.IsExported = false
	}()
	if err := p.walkExpr(sc, node.Decl); err != nil {
		return err
	}
	if len(node.List) > 0 {
		return fmt.Errorf("parser: walk exported aliases not implemented yet")
	}
	return nil
}

func (p *Parser) walkFuncDecl(sc *scope.Scope, node *js.FuncDecl) error {
	sc.IsDeclaration = true
	defer func() { sc.IsDeclaration = false }()
	if node.Name != nil {
		if err := p.walkVar(sc, node.Name); err != nil {
			return err
		}
	}
	childScope := sc.New()
	for _, param := range node.Params.List {
		if err := p.walkBindingElement(childScope, param); err != nil {
			return err
		}
	}
	if err := p.walkBlockStatement(childScope, node.Body); err != nil {
		return err
	}
	return nil
}

func (p *Parser) walkExpr(sc *scope.Scope, node js.IExpr) error {
	switch expr := node.(type) {
	case *js.Var:
		return p.walkVar(sc, expr)
	case *js.VarDecl:
		return p.walkVarDecl(sc, expr)
	case *js.CallExpr:
		return p.walkCallExpr(sc, expr)
	case *js.BinaryExpr:
		return p.walkBinaryExpr(sc, expr)
	case *js.LiteralExpr:
		return p.walkLiteralExpr(sc, expr)
	case *js.CondExpr:
		return p.walkCondExpr(sc, expr)
	case *js.ArrowFunc:
		return p.walkArrowFunc(sc, expr)
	case *js.UnaryExpr:
		return p.walkUnaryExpr(sc, expr)
	case *js.GroupExpr:
		return p.walkGroupExpr(sc, expr)
	default:
		return fmt.Errorf("parser: unexpected expression %T", expr)
	}
}

func (p *Parser) walkVarDecl(sc *scope.Scope, node *js.VarDecl) error {
	sc.IsDeclaration = true
	defer func() { sc.IsDeclaration = false }()
	if node.TokenType == js.VarToken || node.TokenType == js.LetToken {
		sc.IsMutable = true
		defer func() { sc.IsMutable = false }()
	}
	for _, binding := range node.List {
		if err := p.walkBindingElement(sc, binding); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) walkBindingElement(sc *scope.Scope, node js.BindingElement) error {
	if err := p.walkBinding(sc, node.Binding); err != nil {
		return nil
	}
	if node.Default != nil {
		if err := p.walkExpr(sc, node.Default); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) walkBinding(sc *scope.Scope, node js.IBinding) error {
	switch binding := node.(type) {
	case *js.Var:
		return p.walkVar(sc, binding)
	default:
		return fmt.Errorf("unexpected binding %T", binding)
	}
}

func (p *Parser) walkBinaryExpr(sc *scope.Scope, node *js.BinaryExpr) error {
	if err := p.walkExpr(sc, node.X); err != nil {
		return err
	}
	if err := p.walkExpr(sc, node.Y); err != nil {
		return err
	}
	return nil
}

func (p *Parser) walkVar(sc *scope.Scope, node *js.Var) error {
	name := string(node.Data)
	sc.Use(name)
	return nil
}

func (p *Parser) walkCondExpr(sc *scope.Scope, node *js.CondExpr) error {
	if err := p.walkExpr(sc, node.Cond); err != nil {
		return err
	}
	if err := p.walkExpr(sc, node.X); err != nil {
		return err
	}
	if err := p.walkExpr(sc, node.Y); err != nil {
		return err
	}
	return nil
}

func (p *Parser) walkArrowFunc(sc *scope.Scope, node *js.ArrowFunc) error {
	childScope := sc.New()
	for _, param := range node.Params.List {
		if err := p.walkBindingElement(childScope, param); err != nil {
			return err
		}
	}
	if err := p.walkBlockStatement(childScope, node.Body); err != nil {
		return err
	}
	return nil
}

func (p *Parser) walkUnaryExpr(sc *scope.Scope, node *js.UnaryExpr) error {
	if err := p.walkExpr(sc, node.X); err != nil {
		return err
	}
	return nil
}

func (p *Parser) walkGroupExpr(sc *scope.Scope, node *js.GroupExpr) error {
	if err := p.walkExpr(sc, node.X); err != nil {
		return err
	}
	return nil
}

func (p *Parser) walkCallExpr(sc *scope.Scope, node *js.CallExpr) error {
	if err := p.walkExpr(sc, node.X); err != nil {
		return err
	}
	for _, arg := range node.Args.List {
		if err := p.walkExpr(sc, arg.Value); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) walkReturnStmt(sc *scope.Scope, node *js.ReturnStmt) error {
	if err := p.walkExpr(sc, node.Value); err != nil {
		return err
	}
	return nil
}

func (p *Parser) walkLiteralExpr(sc *scope.Scope, lit *js.LiteralExpr) error {
	return nil
}
