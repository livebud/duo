package parser

import (
	"fmt"

	"github.com/livebud/duo/internal/ast"
	"github.com/livebud/duo/internal/lexer"
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
	return &Parser{l: l}
}

type Parser struct {
	l *lexer.Lexer
}

func (p *Parser) Parse() (*ast.Document, error) {
	return p.parseDocument()
}

func (p *Parser) unexpected() error {
	return fmt.Errorf("unexpected token %s", p.l.Token.String())
}

func (p *Parser) parseDocument() (*ast.Document, error) {
	doc := new(ast.Document)
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
		return p.parseMustache()
	default:
		return nil, p.unexpected()
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
		return nil, p.unexpected()
	}
}

func (p *Parser) parseElement() (*ast.Element, error) {
	node := new(ast.Element)
	node.Name = p.l.Token.Text

	// Opening tag
openTag:
	for p.l.Next() {
		switch p.l.Token.Type {
		// Skip over whitespace
		case token.Space:
			continue
		case token.SlashGreaterThan:
			node.SelfClosing = true
			return node, nil
		case token.GreaterThan:
			break openTag
		// TODO: handle attributes
		default:
			return nil, p.unexpected()
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

	// Closing tag
	if p.l.Token.Type != token.LessThanSlash {
		return nil, p.unexpected()
	}
	p.l.Next()
	for p.l.Token.Type == token.Space {
		p.l.Next()
	}
	if p.l.Token.Type != token.Identifier {
		return nil, p.unexpected()
	} else if p.l.Token.Text != node.Name {
		return nil, fmt.Errorf("expected closing tag %s, got %s", node.Name, p.l.Token.Text)
	}
	p.l.Next()
	for p.l.Token.Type == token.Space {
		p.l.Next()
	}
	if p.l.Token.Type != token.GreaterThan {
		return nil, p.unexpected()
	}
	return node, nil
}

func (p *Parser) parseMustache() (*ast.Mustache, error) {
	node := new(ast.Mustache)
	p.l.Next()
	if p.l.Token.Type != token.Expr {
		return nil, p.unexpected()
	}
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	node.Expr = expr
	p.l.Next()
	if p.l.Token.Type != token.RightBrace {
		return nil, p.unexpected()
	}
	return node, nil
}

var options = js.Options{}

func (p *Parser) parseExpression() (js.IExpr, error) {
	ast, err := js.Parse(parse.NewInputString(p.l.Token.Text), options)
	if err != nil {
		return nil, err
	}
	stmts := ast.BlockStmt.List
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

func (p *Parser) parseScript() (*ast.Script, error) {
	node := new(ast.Script)
	node.Name = p.l.Token.Text

	// Opening tag
openTag:
	for p.l.Next() {
		switch p.l.Token.Type {
		// Skip over whitespace
		case token.Space:
			continue
		case token.SlashGreaterThan:
			return node, nil
		case token.GreaterThan:
			break openTag
		// TODO: handle attributes
		default:
			return nil, p.unexpected()
		}
	}

	// Parse the program
	p.l.Next()
	program, err := js.Parse(parse.NewInputString(p.l.Token.Text), options)
	if err != nil {
		return nil, err
	}
	node.Program = program
	p.l.Next()

	// Closing tag
	if p.l.Token.Type != token.LessThanSlash {
		return nil, p.unexpected()
	}
	p.l.Next()
	for p.l.Token.Type == token.Space {
		p.l.Next()
	}
	if p.l.Token.Type != token.Script {
		return nil, p.unexpected()
	}
	if p.l.Token.Text != node.Name {
		return nil, fmt.Errorf("expected closing tag %s, got %s", node.Name, p.l.Token.Text)
	}
	p.l.Next()
	for p.l.Token.Type == token.Space {
		p.l.Next()
	}
	if p.l.Token.Type != token.GreaterThan {
		return nil, p.unexpected()
	}
	return node, nil
}
