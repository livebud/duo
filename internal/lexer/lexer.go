package lexer

import (
	"strings"

	"github.com/livebud/duo/internal/token"
)

type stateFn func(l *Lexer) token.Token

func New(input string) *Lexer {
	l := &Lexer{input: input, statesFns: []stateFn{initialState}}
	l.step()
	return l
}

func Lex(input string) []token.Token {
	l := New(input)
	var tokens []token.Token
	for {
		tok := l.Next()
		tokens = append(tokens, tok)
		if tok.Type == token.EndOfInput {
			break
		}
	}
	return tokens
}

// Print the input as tokens
func Print(input string) string {
	tokens := Lex(input)
	stoken := make([]string, len(tokens))
	for i, token := range tokens {
		stoken[i] = token.String()
	}
	return strings.Join(stoken, " ")
}

type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
	statesFns    []stateFn
}

func (l *Lexer) step() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

func (l *Lexer) Next() token.Token {
	return l.statesFns[len(l.statesFns)-1](l)
}

func (l *Lexer) pushState(state func(l *Lexer) token.Token) {
	l.statesFns = append(l.statesFns, state)
}

func (l *Lexer) popState() {
	l.statesFns = l.statesFns[:len(l.statesFns)-1]
}

func isHeaderNumber(ch byte) bool {
	return '1' <= ch && ch <= '7'
}

func (l *Lexer) readIdentifier() (token.Type, string) {
	position := l.position
	typ := token.Identifier
	i := 0
loop:
	for l.ch != 0 {
		switch {
		case i == 0 && isUpper(l.ch):
			typ = token.UpperIdentifier
		case i > 0 && l.ch == '-':
			typ = token.DashIdentifier
		case i > 0 && l.ch == '.':
			typ = token.DotIdentifier
		case i > 0 && l.ch == ':':
			typ = token.ColonIdentifier
		case i == 1 && isHeaderNumber(l.ch):
		case isLetter(l.ch):
			// continue on
		default:
			break loop
		}
		i++
		l.step()
	}
	return typ, l.input[position:l.position]
}

func (l *Lexer) peek() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) readUntil(ends ...byte) string {
	position := l.position
loop:
	for l.ch != 0 {
		for _, end := range ends {
			if l.ch == end {
				break loop
			}
		}
		l.step()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readLessThan() token.Type {
	switch ch := l.peek(); ch {
	default:
		return token.LessThan
	case '/':
		l.step()
		return token.LessThanSlash
	case '!':
		l.step()
		// Continue on
	}
	if ch := l.peek(); ch != '-' {
		return token.LessThanExclamation
	}
	l.step()
	if ch := l.peek(); ch != '-' {
		return token.Text
	}
	l.step()
	return token.OpenComment
}

func isUpper(ch byte) bool {
	return 'A' <= ch && ch <= 'Z'
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z'
}

func initialState(l *Lexer) token.Token {
	var tok token.Token
	switch l.ch {
	case '<':
		switch tok.Type = l.readLessThan(); tok.Type {
		case token.LessThan, token.LessThanExclamation, token.LessThanSlash:
			l.pushState(tagState)
		case token.OpenComment:
			l.pushState(commentState)
		case token.Text:
			// Continue on
		default:
			tok.Type = token.Unexpected
		}
	case '{':
		tok.Type = token.OpenCurly
		l.pushState(expressionState)
	case ' ', '\t', '\n':
		l.readSpace()
		tok.Type = token.Space
		return tok
	case 0:
		tok.Type = token.EndOfInput
	default:
		tok.Type = token.Text
		tok.Literal = l.readUntil('<', '{')
		return tok
	}
	l.step()
	return tok
}

func isSpace(ch byte) bool {
	switch ch {
	case ' ', '\t', '\n':
		return true
	default:
		return false
	}
}

func (l *Lexer) readSpace() {
	for isSpace(l.ch) {
		l.step()
	}
}

func tagState(l *Lexer) token.Token {
	var tok token.Token
	switch l.ch {
	case '>':
		tok.Type = token.GreaterThan
		l.popState()
	case '=':
		tok.Type = token.Equal

		if l.peek() != '{' {
			l.pushState(attributeValueState)
		}
	case ' ', '\t', '\n':
		l.readSpace()
		tok.Type = token.Space
		return tok
	case '{':
		tok.Type = token.OpenCurly
		l.pushState(expressionState)
	default:
		if isLetter(l.ch) {
			tok.Type, tok.Literal = l.readIdentifier()
			return tok
		} else {
			tok.Type = token.Unexpected
			l.popState()
		}
	}
	l.step()
	return tok
}

func (l *Lexer) readCloseComment() token.Type {
	l.step()
	if l.ch != '-' {
		return token.Text
	}
	l.step()
	if l.ch != '>' {
		return token.Text
	}
	l.step()
	return token.CloseComment
}

func commentState(l *Lexer) token.Token {
	var tok token.Token
	switch l.ch {
	case '-':
		switch tok.Type = l.readCloseComment(); tok.Type {
		case token.CloseComment:
			l.popState()
			return tok
		case token.Text:
			// Continue on
		default:
			tok.Type = token.Unexpected
		}
	default:
		tok.Type = token.Text
		tok.Literal = l.readUntil('-')
		return tok
	}
	l.step()
	return tok
}

func expressionState(l *Lexer) token.Token {
	var tok token.Token
	switch l.ch {
	case '}':
		l.popState()
		tok.Type = token.CloseCurly
	default:
		tok.Type = token.Text
		tok.Literal = l.readUntil('}')
		return tok
	}
	l.step()
	return tok
}

func isEndOfAttributeValue(ch byte) bool {
	switch ch {
	case ' ', '\t', '\n', '>', 0:
		return true
	default:
		return false
	}
}

func (l *Lexer) readAttributeValue() string {
	position := l.position
	for !isEndOfAttributeValue(l.ch) {
		l.step()
	}
	return l.input[position:l.position]
}

func attributeValueState(l *Lexer) token.Token {
	var tok token.Token
	switch l.ch {
	// case '{':
	// 	tok.Type = token.OpenCurly
	// 	l.pushState(expressionState)
	case '"':
		tok.Type = token.Quote
		l.popState()
		l.pushState(doubleQuoteState)
	case '\'':
		tok.Type = token.Quote
		l.popState()
		l.pushState(singleQuoteState)
	default:
		tok.Type = token.Text
		tok.Literal = l.readAttributeValue()
		l.popState()
		return tok
	}
	l.step()
	return tok
}

func doubleQuoteState(l *Lexer) token.Token {
	var tok token.Token
	switch {
	case l.ch == '"':
		tok.Type = token.Quote
		l.popState()
	case l.ch == '{':
		tok.Type = token.OpenCurly
		l.pushState(expressionState)
	default:
		tok.Type = token.Text
		tok.Literal = l.readUntil('"', '{')
		return tok
	}
	l.step()
	return tok
}

func singleQuoteState(l *Lexer) token.Token {
	var tok token.Token
	switch {
	case l.ch == '\'':
		tok.Type = token.Quote
		l.popState()
	case l.ch == '{':
		tok.Type = token.OpenCurly
		l.pushState(expressionState)
	default:
		tok.Type = token.Text
		tok.Literal = l.readUntil('\'', '{')
		return tok
	}
	l.step()
	return tok
}
