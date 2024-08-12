package parser

import (
	"fmt"
	"strings"

	"github.com/livebud/duo/internal/ast"
	"github.com/livebud/duo/internal/event"
	"github.com/livebud/duo/internal/js"
	"github.com/livebud/duo/internal/lexer"
	"github.com/livebud/duo/internal/scope"
	"github.com/livebud/duo/internal/token"
	"github.com/matthewmueller/css"
)

func Parse(path, input string) (*ast.Document, error) {
	l := lexer.New(input)
	p := New(path, l)
	return p.Parse()
}

func Print(path, input string) string {
	doc, err := Parse(path, input)
	if err != nil {
		return err.Error()
	}
	return doc.String()
}

func New(path string, l *lexer.Lexer) *Parser {
	return &Parser{path, l, scope.New()}
}

type Parser struct {
	path string
	l    *lexer.Lexer
	sc   *scope.Scope
}

func (p *Parser) Parse() (*ast.Document, error) {
	return p.parseDocument()
}

func (p *Parser) errorf(format string, args ...interface{}) error {
	return fmt.Errorf("parser: %s: %s", p.path, fmt.Sprintf(format, args...))
}

// TODO: this needs to be updated to better handle peaked tokens
func (p *Parser) unexpected(prefix string) error {
	token := p.l.Latest()
	return p.errorf("%s unexpected token %s (%d:%d)", prefix, token.String(), token.Line, token.Start)
}

func (p *Parser) parseDocument() (*ast.Document, error) {
	doc := &ast.Document{
		Scope: p.sc,
	}
	for !p.Accept(token.EOF) {
		child, err := p.parseFragment()
		if err != nil {
			return nil, err
		}
		doc.Children = append(doc.Children, child)
	}
	return doc, nil
}

func (p *Parser) parseFragment() (ast.Fragment, error) {
	switch {
	case p.Accept(token.Text):
		return p.parseText()
	case p.Accept(token.LessThan):
		return p.parseTag()
	case p.Accept(token.Comment):
		return p.parseComment()
	case p.Accept(token.LeftBrace):
		switch {
		case p.Accept(token.Hash):
			return p.parseBlock()
		default:
			return p.parseMustache()
		}
	default:
		return nil, p.unexpected("fragment")
	}
}

func (p *Parser) parseText() (*ast.Text, error) {
	return &ast.Text{
		Value: p.Text(),
	}, nil
}

func (p *Parser) parseTag() (ast.Fragment, error) {
	switch {
	// case p.Accept(token.Doctype):
	// 	return p.parseDoctype()
	case p.Accept(token.Identifier):
		return p.parseElement()
	case p.Accept(token.PascalIdentifier):
		return p.parseComponent()
	case p.Accept(token.Style):
		return p.parseStyle()
	case p.Accept(token.Script):
		return p.parseScript()
	case p.Accept(token.Slot):
		return p.parseSlot()
	default:
		return nil, p.unexpected("tag")
	}
}

func (p *Parser) parseElement() (*ast.Element, error) {
	node := &ast.Element{
		Name: p.Text(),
	}

	// Handle attributes
	for !p.Check(token.GreaterThan) && !p.Check(token.SlashGreaterThan) {
		attr, err := p.parseAttribute()
		if err != nil {
			return nil, err
		}
		node.Attributes = append(node.Attributes, attr)
	}
	if p.Accept(token.SlashGreaterThan) {
		node.SelfClosing = true
		return node, nil
	}
	if err := p.Expect(token.GreaterThan); err != nil {
		return nil, err
	}

	for !p.Accept(token.LessThanSlash) {
		child, err := p.parseFragment()
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, child)
	}

	// Closing tag
	if err := p.Expect(token.Identifier); err != nil {
		return nil, err
	} else if p.Text() != node.Name {
		return nil, p.errorf("expected closing tag %s, got %s", node.Name, p.Text())
	}
	if err := p.Expect(token.GreaterThan); err != nil {
		return nil, err
	}

	return node, nil
}

func (p *Parser) parseComponent() (*ast.Component, error) {
	node := &ast.Component{
		Name: p.Text(),
	}

	// Handle attributes
	for !p.Check(token.GreaterThan) && !p.Check(token.SlashGreaterThan) {
		attr, err := p.parseAttribute()
		if err != nil {
			return nil, err
		}
		node.Attributes = append(node.Attributes, attr)
	}
	if p.Accept(token.SlashGreaterThan) {
		node.SelfClosing = true
		return node, nil
	}
	if err := p.Expect(token.GreaterThan); err != nil {
		return nil, err
	}

	for !p.Accept(token.LessThanSlash) {
		child, err := p.parseFragment()
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, child)
	}

	// Closing tag
	if err := p.Expect(token.PascalIdentifier); err != nil {
		return nil, err
	} else if p.Text() != node.Name {
		return nil, p.errorf("expected closing tag %s, got %s", node.Name, p.Text())
	}
	if err := p.Expect(token.GreaterThan); err != nil {
		return nil, err
	}

	return node, nil
}

func (p *Parser) parseAttribute() (ast.Attribute, error) {
	switch {
	case p.Accept(token.Identifier):
		name := p.Text()
		if p.Accept(token.Colon) {
			switch name {
			case "bind":
				return p.parseBind()
			case "class":
				return p.parseClass()
			default:
				return nil, p.unexpected("colon attribute")
			}
		}
		return p.parseField()
	case p.Accept(token.LeftBrace):
		return p.parseAttributeShorthand()
	case p.Accept(token.Slot):
		return p.parseNamedSlot()
	default:
		return nil, p.unexpected("attribute")
	}
}

func (p *Parser) parseBind() (*ast.Binding, error) {
	node := &ast.Binding{}
	if err := p.Expect(token.Identifier); err != nil {
		return nil, err
	}
	node.Key = p.Text()
	if err := p.Expect(token.Equal); err != nil {
		return nil, err
	}
	if err := p.Expect(token.LeftBrace); err != nil {
		return nil, err
	}
	value, err := p.parseMustache()
	if err != nil {
		return nil, err
	}
	node.Value = value
	return node, nil
}

func (p *Parser) parseClass() (*ast.Class, error) {
	node := &ast.Class{}
	if err := p.Expect(token.Identifier); err != nil {
		return nil, err
	}
	node.Name = p.Text()
	if err := p.Expect(token.Equal); err != nil {
		return nil, err
	}
	if err := p.Expect(token.LeftBrace); err != nil {
		return nil, err
	}
	value, err := p.parseMustache()
	if err != nil {
		return nil, err
	}
	node.Value = value
	return node, nil
}

func (p *Parser) parseField() (*ast.Field, error) {
	field := &ast.Field{
		Key:          p.Text(),
		EventHandler: event.Is(p.Text()),
	}
	if err := p.Expect(token.Equal); err != nil {
		return nil, err
	}
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
}

func (p *Parser) parseAttributeValues() (values []ast.Value, err error) {
	for !p.Is(token.GreaterThan, token.SlashGreaterThan, token.Identifier) {
		switch {
		case p.Accept(token.Quote):
			return p.parseAttributeStringValues()
		case p.Accept(token.Text):
			text, err := p.parseText()
			if err != nil {
				return nil, err
			}
			values = append(values, text)
		case p.Accept(token.LeftBrace):
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
	for {
		switch {
		case p.Accept(token.Quote):
			// Empty text node
			values = append(values, &ast.Text{Value: ""})
			return values, nil
		case p.Accept(token.Text):
			text, err := p.parseText()
			if err != nil {
				return nil, err
			}
			values = append(values, text)
		case p.Accept(token.LeftBrace):
			mustache, err := p.parseMustache()
			if err != nil {
				return nil, err
			}
			values = append(values, mustache)
		default:
			return nil, p.unexpected("attribute value")
		}
	}
}

func (p *Parser) parseEventValue() (value ast.Value, err error) {
	if err := p.Expect(token.LeftBrace); err != nil {
		return nil, err
	}
	return p.parseMustache()
}

func (p *Parser) parseMustache() (*ast.Mustache, error) {
	node := new(ast.Mustache)
	if err := p.Expect(token.Expr); err != nil {
		return nil, err
	}
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	node.Expr = expr
	if err := p.Expect(token.RightBrace); err != nil {
		return nil, err
	}
	return node, nil
}

func (p *Parser) parseBlock() (ast.Fragment, error) {
	switch {
	case p.Accept(token.If):
		return p.parseIfBlock()
	case p.Accept(token.Each):
		return p.parseEachBlock()
	default:
		return nil, p.unexpected("block")
	}
}

// func (p *Parser) parseCloseBlock() (ast.Fragment, error) {
// 	switch {
// 	case p.Accept(token.If):
// 		return p.parseIfBlock()
// 	case p.Accept(token.Each):
// 		return p.parseEachBlock()
// 	default:
// 		return nil, p.unexpected("close block")
// 	}
// }

// func (p *Parser) parseContinuedBlock() (ast.Fragment, error) {
// 	switch {
// 	case p.Accept(token.Else):
// 		return p.parseElseBlock()
// 	case p.Accept(token.ElseIf):
// 		return p.parseElseIfBlock()
// 	default:
// 		return nil, p.unexpected("close block")
// 	}
// }

func (p *Parser) parseIfBlock() (*ast.IfBlock, error) {
	node := new(ast.IfBlock)
	if err := p.Expect(token.Expr); err != nil {
		return nil, err
	}
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	node.Cond = expr
	if err := p.Expect(token.RightBrace); err != nil {
		return nil, err
	}
	for !p.Accept(token.LeftBrace, token.Slash) {
		switch {
		case p.Accept(token.EOF):
			return nil, p.errorf("unclosed if block")
		case p.Accept(token.LeftBrace, token.Colon, token.ElseIf):
			ifBlock, err := p.parseElseIfBlock()
			if err != nil {
				return nil, err
			}
			node.Else = append(node.Else, ifBlock)
			continue
		case p.Accept(token.LeftBrace, token.Colon, token.Else):
			fragments, err := p.parseElseBlock()
			if err != nil {
				return nil, err
			}
			node.Else = append(node.Else, fragments...)
		default:
			fragment, err := p.parseFragment()
			if err != nil {
				return nil, err
			}
			node.Then = append(node.Then, fragment)
		}
	}
	if err := p.Expect(token.If, token.RightBrace); err != nil {
		return nil, err
	}
	return node, nil
}

func (p *Parser) parseElseIfBlock() (*ast.IfBlock, error) {
	node := new(ast.IfBlock)
	if err := p.Expect(token.Expr); err != nil {
		return nil, err
	}
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	node.Cond = expr
	if err := p.Expect(token.RightBrace); err != nil {
		return nil, err
	}
	for !p.Check(token.LeftBrace, token.Slash) {
		switch {
		case p.Accept(token.EOF):
			return nil, p.errorf("unclosed if block")
		case p.Accept(token.LeftBrace, token.Colon, token.ElseIf):
			ifBlock, err := p.parseElseIfBlock()
			if err != nil {
				return nil, err
			}
			node.Else = append(node.Else, ifBlock)
		case p.Accept(token.LeftBrace, token.Colon, token.Else):
			fragments, err := p.parseElseBlock()
			if err != nil {
				return nil, err
			}
			node.Else = append(node.Else, fragments...)
		default:
			fragment, err := p.parseFragment()
			if err != nil {
				return nil, err
			}
			node.Then = append(node.Then, fragment)
		}
	}
	return node, nil
}

func (p *Parser) parseElseBlock() (fragments []ast.Fragment, err error) {
	if err := p.Expect(token.RightBrace); err != nil {
		return nil, err
	}
	for !p.Check(token.LeftBrace, token.Slash) {
		fragment, err := p.parseFragment()
		if err != nil {
			return nil, err
		}
		fragments = append(fragments, fragment)
	}
	return fragments, nil
}

func (p *Parser) parseEachBlock() (*ast.EachBlock, error) {
	node := new(ast.EachBlock)
	if err := p.Expect(token.Expr); err != nil {
		return nil, err
	}
	left, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	node.List = left

	// handle as
	if p.Accept(token.As) {
		if err := p.Expect(token.Expr); err != nil {
			return nil, err
		}
		middle, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		value, err := p.exprToVar(middle)
		if err != nil {
			return nil, err
		}
		node.Value = value

		// handle key
		if p.Accept(token.Comma) {
			if err := p.Expect(token.Expr); err != nil {
				return nil, err
			}
			right, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			key, err := p.exprToVar(right)
			if err != nil {
				return nil, err
			}
			node.Key = key
		}
	}

	// Closing brace
	if err := p.Expect(token.RightBrace); err != nil {
		return nil, err
	}

	// Parse the body
	for !p.Accept(token.LeftBrace, token.Slash) {
		switch {
		case p.Accept(token.EOF):
			return nil, p.errorf("unclosed each block")
		default:
			fragment, err := p.parseFragment()
			if err != nil {
				return nil, err
			}
			node.Body = append(node.Body, fragment)
		}
	}

	// Closing block
	if err := p.Expect(token.Each, token.RightBrace); err != nil {
		return nil, err
	}
	return node, nil
}

// Checks that the next token is one of the given types
func (p *Parser) Is(types ...token.Type) bool {
	token := p.l.Peak(1)
	for _, t := range types {
		if token.Type == t {
			return true
		}
	}
	return false
}

// Returns true if all the given tokens are next
func (p *Parser) Check(tokens ...token.Type) bool {
	for i, token := range tokens {
		if p.l.Peak(i+1).Type != token {
			return false
		}
	}
	return true
}

// Returns true if all the given tokens are next
func (p *Parser) Accept(tokens ...token.Type) bool {
	if !p.Check(tokens...) {
		return false
	}
	for i := 0; i < len(tokens); i++ {
		p.l.Next()
	}
	return true
}

func (p *Parser) Expect(tokens ...token.Type) error {
	for i, tok := range tokens {
		peaked := p.l.Peak(i + 1)
		if peaked.Type == token.Error {
			return p.errorf(peaked.Text)
		} else if peaked.Type != tok {
			return p.errorf("expected %s, got %s", tok, peaked.Type)
		}
	}
	for i := 0; i < len(tokens); i++ {
		p.l.Next()
	}
	return nil
}

// Type of the current token
func (p *Parser) Type() token.Type {
	return p.l.Token.Type
}

// Text of the current token
func (p *Parser) Text() string {
	return p.l.Token.Text
}

func (p *Parser) parseAttributeShorthand() (ast.Attribute, error) {
	if err := p.Expect(token.Expr); err != nil {
		return nil, err
	}
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	ident, err := p.exprToVar(expr)
	if err != nil {
		return nil, err
	}
	if err := p.Expect(token.RightBrace); err != nil {
		return nil, err
	}
	name := string(ident.Data)
	return &ast.AttributeShorthand{
		Key:          name,
		EventHandler: event.Is(name),
	}, nil
}

func (p *Parser) parseNamedSlot() (*ast.NamedSlot, error) {
	node := &ast.NamedSlot{}
	if err := p.Expect(token.Equal); err != nil {
		return nil, err
	}
	if err := p.Expect(token.Quote); err != nil {
		return nil, err
	}
	if err := p.Expect(token.Text); err != nil {
		return nil, err
	}
	node.Name = p.Text()
	if err := p.Expect(token.Quote); err != nil {
		return nil, err
	}
	return node, nil
}

func (p *Parser) parseExpression() (js.IExpr, error) {
	expr, err := js.ParseExpr(p.l.Token.Text)
	if err != nil {
		return nil, err
	}
	// Walk the expression to update scope
	if err := p.walkExpr(p.sc, expr); err != nil {
		return nil, err
	}
	return expr, nil
}

func (p *Parser) parseScript() (*ast.Script, error) {
	node := &ast.Script{}

	// Handle attributes
	for !p.Check(token.GreaterThan) && !p.Check(token.SlashGreaterThan) {
		attr, err := p.parseAttribute()
		if err != nil {
			return nil, err
		}
		node.Attributes = append(node.Attributes, attr)
	}
	if p.Accept(token.SlashGreaterThan) {
		node.SelfClosing = true
		return node, nil
	}
	if err := p.Expect(token.GreaterThan); err != nil {
		return nil, err
	}

	// Expect the program
	if err := p.Expect(token.Text); err != nil {
		return nil, err
	}

	jsCode := p.Text()

	// Closing tag
	if err := p.Expect(token.LessThanSlash); err != nil {
		return nil, err
	}
	if err := p.Expect(token.Script); err != nil {
		return nil, err
	}
	if err := p.Expect(token.GreaterThan); err != nil {
		return nil, err
	}
	// Parse the program
	program, err := js.ParseTS(jsCode)
	if err != nil {
		return nil, err
	}
	// Walk the program to update scope
	if err := p.walkBlockStatement(p.sc, program.BlockStmt); err != nil {
		return nil, err
	}
	node.Program = program
	return node, nil
}

func (p *Parser) parseStyle() (*ast.Style, error) {
	node := &ast.Style{}

	// Handle attributes
	for !p.Check(token.GreaterThan) && !p.Check(token.SlashGreaterThan) {
		attr, err := p.parseAttribute()
		if err != nil {
			return nil, err
		}
		node.Attributes = append(node.Attributes, attr)
	}
	if p.Accept(token.SlashGreaterThan) {
		node.SelfClosing = true
		return node, nil
	}
	if err := p.Expect(token.GreaterThan); err != nil {
		return nil, err
	}

	// Expect the program
	if err := p.Expect(token.Text); err != nil {
		return nil, err
	}

	cssCode := p.Text()

	// Closing tag
	if err := p.Expect(token.LessThanSlash); err != nil {
		return nil, err
	}
	if err := p.Expect(token.Style); err != nil {
		return nil, err
	}
	if err := p.Expect(token.GreaterThan); err != nil {
		return nil, err
	}
	// Parse the stylesheet
	stylesheet, err := css.Parse(p.path, cssCode)
	if err != nil {
		return nil, err
	}
	node.StyleSheet = stylesheet
	return node, nil
}

func (p *Parser) parseSlot() (*ast.Slot, error) {
	node := &ast.Slot{}

	if p.Accept(token.Identifier) {
		if p.Text() != "name" {
			return nil, p.errorf("expected slot name, got %s", p.Text())
		}
		if err := p.Expect(token.Equal); err != nil {
			return nil, err
		}
		if err := p.Expect(token.Quote); err != nil {
			return nil, err
		}
		if err := p.Expect(token.Text); err != nil {
			return nil, err
		}
		node.Name = p.Text()
		if err := p.Expect(token.Quote); err != nil {
			return nil, err
		}
	}
	if p.Accept(token.SlashGreaterThan) {
		node.SelfClosing = true
		return node, nil
	}
	if err := p.Expect(token.GreaterThan); err != nil {
		return nil, err
	}
	for !p.Accept(token.LessThanSlash) {
		child, err := p.parseFragment()
		if err != nil {
			return nil, err
		}
		node.Fallback = append(node.Fallback, child)
	}

	// Closing tag
	if err := p.Expect(token.Slot); err != nil {
		return nil, err
	}
	if err := p.Expect(token.GreaterThan); err != nil {
		return nil, err
	}

	return node, nil
}

func (p *Parser) parseComment() (*ast.Comment, error) {
	return &ast.Comment{
		Value: p.Text(),
	}, nil
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
	case *js.ImportStmt:
		return p.walkImportStmt(sc, stmt)
	case *js.ExprStmt:
		return p.walkExprStmt(sc, stmt)
	case *js.FuncDecl:
		return p.walkFuncDecl(sc, stmt)
	case *js.ReturnStmt:
		return p.walkReturnStmt(sc, stmt)
	default:
		return p.errorf("parser: unexpected statement %T", stmt)
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
		return p.errorf("parser: walk exported aliases not implemented yet")
	}
	return nil
}

func (p *Parser) walkImportStmt(sc *scope.Scope, node *js.ImportStmt) error {
	importPath := strings.Trim(string(node.Module), `"'`)
	if node.Default != nil {
		sym := sc.Use(string(node.Default))
		sym.Import = &scope.Import{
			Path:    importPath,
			Default: true,
		}
	}
	if len(node.List) > 0 {
		return p.errorf("parser: walk imported aliases not implemented yet")
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
	case *js.ArrayExpr:
		return p.walkArrayExpr(sc, expr)
	case *js.CondExpr:
		return p.walkCondExpr(sc, expr)
	case *js.ArrowFunc:
		return p.walkArrowFunc(sc, expr)
	case *js.UnaryExpr:
		return p.walkUnaryExpr(sc, expr)
	case *js.GroupExpr:
		return p.walkGroupExpr(sc, expr)
	case *js.DotExpr:
		return p.walkDotExpr(sc, expr)
	default:
		return p.errorf("parser: unexpected expression %T", expr)
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
		return p.errorf("unexpected binding %T", binding)
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

func (p *Parser) walkDotExpr(sc *scope.Scope, node *js.DotExpr) error {
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

func (p *Parser) walkLiteralExpr(_ *scope.Scope, _ *js.LiteralExpr) error {
	return nil
}

func (p *Parser) walkArrayExpr(_ *scope.Scope, _ *js.ArrayExpr) error {
	return nil
}

func (p *Parser) exprToVar(expr js.IExpr) (*js.Var, error) {
	ident, ok := expr.(*js.Var)
	if !ok {
		return nil, p.errorf("expected an identifier, got %T", expr)
	}
	return ident, nil
}
