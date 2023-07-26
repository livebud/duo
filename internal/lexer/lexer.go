package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/livebud/duo/internal/token"
)

type state = func(l *Lexer) token.Type

func New(input string) *Lexer {
	l := &Lexer{
		input:  input,
		states: []state{textState},
		line:   1,
	}
	l.step()
	return l
}

func Lex(input string) []token.Token {
	l := New(input)
	var tokens []token.Token
	for l.Next() {
		tokens = append(tokens, l.Token)
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
	Token token.Token // Current token
	input string      // Input string
	start int         // Index to the start of the current token
	end   int         // Index to the end of the current token
	cp    rune        // Code point being considered
	next  int         // Index to the next rune to be considered
	line  int         // Line number
	err   string      // Error message for an error token

	states []state // Stack of states
	peaked []token.Token

	inScript bool
	inStyle  bool
}

func (l *Lexer) nextToken() token.Token {
	l.start = l.end
	tokenType := l.states[len(l.states)-1](l)
	t := token.Token{
		Type:  tokenType,
		Start: l.start,
		Text:  l.input[l.start:l.end],
		Line:  l.line,
	}
	if tokenType == token.Error {
		t.Text = l.err
		l.err = ""
	}
	return t
}

func (l *Lexer) Next() bool {
	if len(l.peaked) > 0 {
		l.Token = l.peaked[0]
		l.peaked = l.peaked[1:]
	} else {
		l.Token = l.nextToken()
	}
	return l.Token.Type != token.EOF
}

func (l *Lexer) Peak(nth int) token.Token {
	if len(l.peaked) >= nth {
		return l.peaked[nth-1]
	}
	for i := len(l.peaked); i < nth; i++ {
		l.peaked = append(l.peaked, l.nextToken())
	}
	return l.peaked[nth-1]
}

// TODO: replace with an errorf that creates a nice error message
func (l *Lexer) Latest() token.Token {
	if len(l.peaked) > 0 {
		return l.peaked[len(l.peaked)-1]
	}
	return l.Token
}

// Use -1 to indicate the end of the file
const eof = -1

// Step advances the lexer to the next token
func (l *Lexer) step() {
	codePoint, width := utf8.DecodeRuneInString(l.input[l.next:])
	if width == 0 {
		codePoint = eof
	}
	l.cp = codePoint
	l.end = l.next
	l.next += width
	if l.cp == '\n' {
		l.line++
	}
}

func (l *Lexer) ignore() {
	l.start = l.end
}

func (l *Lexer) accept(cp rune, run ...rune) bool {
	// Check the current rune
	if l.cp != cp {
		return false
	}
	str := l.peak(len(run))
	if len(str) != len(run) {
		return false
	}
	for i, r := range str {
		if r != run[i] {
			return false
		}
	}
	for i := 0; i < len(run)+1; i++ {
		l.step()
	}
	return true
}

func (l *Lexer) acceptFold(cp rune, run ...rune) bool {
	// Check the current rune
	if unicode.ToLower(l.cp) != unicode.ToLower(cp) {
		return false
	}
	str := l.peak(len(run))
	if len(str) != len(run) {
		return false
	}
	for i, r := range str {
		if unicode.ToLower(r) != unicode.ToLower(run[i]) {
			return false
		}
	}
	for i := 0; i < len(run)+1; i++ {
		l.step()
	}
	return true
}

func (l *Lexer) text() string {
	return l.input[l.start:l.end]
}

func (l *Lexer) stepUntil(rs ...rune) bool {
	for {
		if l.cp == eof {
			return false
		}
		for _, r := range rs {
			if l.cp == r {
				return true
			}
		}
		l.step()
	}
}

func (l *Lexer) peak(n int) string {
	s := new(strings.Builder)
	if n == 0 {
		s.WriteRune(l.cp)
		return s.String()
	}
	next := l.next
	for i := 0; i < n; i++ {
		cp, width := utf8.DecodeRuneInString(l.input[next:])
		if width == 0 {
			cp = eof
			break
		}
		s.WriteRune(cp)
		next += width
	}
	return s.String()
}

func (l *Lexer) pushState(state state) {
	l.states = append(l.states, state)
}

func (l *Lexer) popState() {
	l.states = l.states[:len(l.states)-1]
}

func (l *Lexer) errorf(msg string, args ...interface{}) token.Type {
	l.err = fmt.Sprintf(msg, args...)
	return token.Error
}

func (l *Lexer) unexpected() token.Type {
	cp := l.cp
	l.step()
	if l.cp == eof {
		return l.errorf("unexpected end of input")
	}
	return l.errorf("unexpected token '%s'", string(cp))
}

func textState(l *Lexer) (t token.Type) {
	switch l.cp {
	case eof:
		return token.EOF
	case '<':
		l.step()
		switch {
		case l.cp == '/':
			l.step()
			l.pushState(startCloseTagState)
			return token.LessThanSlash
		case l.cp == '!':
			l.step()
			switch {
			case l.accept('-', '-'):
				for !l.accept('-', '-', '>') {
					l.step()
				}
				return token.Comment
			case l.acceptFold('d', 'o', 'c', 't', 'y', 'p', 'e'):
				l.pushState(doctypeState)
				return token.Doctype
			default:
				return l.unexpected()
			}
		default:
			l.pushState(startOpenTagState)
		}
		return token.LessThan
	case '{':
		l.step()
		l.pushState(blockExpressionState)
		return token.LeftBrace
	default:
		l.step()
		for l.cp != '<' && l.cp != '{' && l.cp != eof {
			l.step()
		}
		return token.Text
	}
}

func startOpenTagState(l *Lexer) token.Type {
	for {

		switch {
		case l.cp == eof:
			l.popState()
			return l.unexpected()
		case isAlpha(l.cp):
			l.step()
			for isAlphaNumeric(l.cp) || isDash(l.cp) {
				l.step()
			}
			l.popState()
			l.pushState(middleTagState)
			switch l.text() {
			case "script":
				l.inScript = true
				return token.Script
			case "style":
				l.inStyle = true
				return token.Style
			}
			return token.Identifier
		case isSpace(l.cp):
			l.step()
			for isSpace(l.cp) {
				l.step()
			}
			l.ignore()
			continue
		default:
			l.popState()
			return l.unexpected()
		}
	}
}

func middleTagState(l *Lexer) (t token.Type) {
	for {
		switch {
		case l.cp == eof:
			l.popState()
			return l.unexpected()
		case isAlpha(l.cp):
			l.step()
			for isAlphaNumeric(l.cp) || isDash(l.cp) {
				l.step()
			}
			return token.Identifier
		case l.cp == '>':
			l.step()
			l.popState()
			if l.inScript {
				l.pushState(scriptState)
			} else if l.inStyle {
				l.pushState(styleState)
			}
			return token.GreaterThan
		case l.cp == '/':
			l.step()
			if l.cp != '>' {
				return l.unexpected()
			}
			l.step()
			l.popState()
			if l.inScript {
				l.pushState(scriptState)
			} else if l.inStyle {
				l.pushState(styleState)
			}
			return token.SlashGreaterThan
		case l.cp == '=':
			l.step()
			l.pushState(attributeState)
			return token.Equal
		case l.cp == '{':
			l.step()
			l.pushState(expressionState)
			return token.LeftBrace
		case isSpace(l.cp):
			l.step()
			for isSpace(l.cp) {
				l.step()
			}
			l.ignore()
			continue
		default:
			l.popState()
			return l.unexpected()
		}
	}
}

func startCloseTagState(l *Lexer) token.Type {
	for {
		switch {
		case l.cp == eof:
			l.popState()
			return l.unexpected()
		case l.cp == '/':
			l.step()
			return token.Slash
		case isAlpha(l.cp):
			l.step()
			for isAlphaNumeric(l.cp) || isDash(l.cp) {
				l.step()
			}
			switch l.text() {
			case "script":
				l.inScript = false
				return token.Script
			case "style":
				l.inStyle = false
				return token.Style
			}
			return token.Identifier
		case l.cp == '>':
			l.step()
			l.popState()
			return token.GreaterThan
		case isSpace(l.cp):
			l.step()
			for isSpace(l.cp) && l.cp != eof {
				l.step()
			}
			l.ignore()
			continue
		default:
			l.popState()
			return l.unexpected()
		}
	}
}

func attributeState(l *Lexer) (t token.Type) {
	switch {
	case l.cp == eof:
		return l.unexpected()
	case l.cp == '\'':
		l.step()
		l.pushState(attributeValueState('\'', token.Text, token.Quote))
		return token.Quote
	case l.cp == '"':
		l.step()
		l.pushState(attributeValueState('"', token.Text, token.Quote))
		return token.Quote
	case l.cp == '{':
		l.step()
		l.popState()
		l.pushState(expressionState)
		return token.LeftBrace
	case isSpace(l.cp):
		return l.unexpected()
	default:
		l.step()
		for !isSpace(l.cp) && l.cp != '>' && l.cp != eof {
			l.step()
		}
		l.popState()
		return token.Text
	}
}

func attributeValueState(until rune, midToken, endToken token.Type) state {
	return func(l *Lexer) token.Type {
		switch l.cp {
		case eof:
			return l.unexpected()
		case until:
			l.step()
			l.popState()
			// Pop out of attributeState as well
			// TODO: clean this up
			l.popState()
			return endToken
		case '{':
			l.step()
			l.pushState(expressionState)
			return token.LeftBrace
		default:
			for l.cp != until && l.cp != '{' && l.cp != eof {
				l.step()
			}
			return midToken
		}
	}
}

func scriptState(l *Lexer) token.Type {
	for {
		switch l.cp {
		case eof:
			return l.unexpected()
		case '<':
			if strings.HasPrefix(l.input[l.end:], "</script>") {
				l.popState()
				return token.Text
			}
			l.step()
		default:
			l.step()
		}
	}
}

func styleState(l *Lexer) token.Type {
	for {
		switch l.cp {
		case eof:
			return l.unexpected()
		case '<':
			if strings.HasPrefix(l.input[l.end:], "</style>") {
				l.popState()
				return token.Text
			}
			l.step()
		default:
			l.step()
		}
	}
}

func blockExpressionState(l *Lexer) token.Type {
	for {
		switch l.cp {
		case eof:
			l.popState()
			return l.unexpected()
		case ' ':
			l.step()
			for l.cp == ' ' {
				l.step()
			}
			continue
			// return token.Space
		case 'i':
			l.popState()
			l.pushState(expressionState)
			if l.accept('i', 'f', ' ') {
				return token.If
			}
			return expressionState(l)
		case 'e':
			l.popState()
			l.pushState(expressionState)
			if l.accept('e', 'n', 'd') {
				for l.cp == ' ' {
					l.step()
				}
				return token.End
			} else if l.accept('e', 'l', 's', 'e') {
				for l.cp == ' ' {
					l.step()
				}
				if l.accept('i', 'f', ' ') {
					return token.ElseIf
				}
				return token.Else
			}
			return expressionState(l)
		case 'f':
			l.popState()
			if l.accept('f', 'o', 'r', ' ') {
				l.pushState(forState)
				return token.For
			}
			l.pushState(expressionState)
			return expressionState(l)
		default:
			l.popState()
			l.pushState(expressionState)
			return expressionState(l)
		}
	}
}

func expressionState(l *Lexer) token.Type {
	for {
		switch l.cp {
		case eof:
			return l.unexpected()
		case '}':
			l.step()
			l.popState()
			return token.RightBrace
		case ' ':
			l.step()
			for l.cp == ' ' {
				l.step()
			}
			// return token.Space
		default:
			// Handle inner right brace } characters
			depth := 1
			for {
				switch l.cp {
				case eof:
					return l.unexpected()
				case '{':
					depth++
					l.step()
				case '}':
					depth--
					if depth == 0 {
						return token.Expr
					}
					l.step()
				default:
					l.step()
				}
			}
		}
	}
}

func doctypeState(l *Lexer) token.Type {
	for {
		switch {
		case l.cp == eof:
			l.popState()
			return l.unexpected()
		case l.cp == '>':
			l.step()
			l.popState()
			return token.GreaterThan
		case l.cp == '/':
			if l.accept('/', '>') {
				l.popState()
				return token.SlashGreaterThan
			}
			return l.unexpected()
		case isSpace(l.cp):
			l.step()
			for isSpace(l.cp) {
				l.step()
			}
			l.ignore()
			continue
		case isAlpha(l.cp):
			l.step()
			for isAlpha(l.cp) {
				l.step()
			}
			return token.Identifier
		default:
			l.popState()
			return l.unexpected()
		}
	}
}

func forState(l *Lexer) token.Type {
	for {
		switch {
		case l.cp == eof:
			l.popState()
			return l.unexpected()
		case isSpace(l.cp):
			l.step()
			for isSpace(l.cp) {
				l.step()
			}
			l.ignore()
			continue
		case l.cp == 'i':
			if l.accept('i', 'n', ' ') {
				l.popState()
				l.pushState(expressionState)
				return token.In
			}
			l.step()
			if l.stepUntil(' ', ',', '}') {
				return token.Expr
			}
		case isIdentifierHead(l.cp):
			if l.stepUntil(' ', ',', '}') {
				return token.Expr
			}
		case l.cp == ',':
			l.step()
			return token.Comma
		case l.cp == '}':
			l.step()
			l.popState()
			return token.RightBrace
		default:
			l.popState()
			return l.unexpected()
		}
	}
}

func isIdentifierHead(cp rune) bool {
	return isAlpha(cp) || cp == '_' || cp == '$'
}

func isAlpha(cp rune) bool {
	return (cp >= 'a' && cp <= 'z') || (cp >= 'A' && cp <= 'Z')
}

func isAlphaNumeric(cp rune) bool {
	return isAlpha(cp) || (cp >= '0' && cp <= '9')
}

func isDash(cp rune) bool {
	return cp == '-'
}

func isSpace(cp rune) bool {
	return cp == ' ' || cp == '\t' || cp == '\n' || cp == '\r'
}
