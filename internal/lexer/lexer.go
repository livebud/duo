package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/livebud/duo/internal/token"
)

type state func(l *Lexer) token.Token

func New(input string) *Lexer {
	l := &Lexer{input: input, states: []state{initialState}}
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
	input  string
	start  int  // Index to the start of the current token
	end    int  // Index to the end of the current token
	cp     rune // Code point being considered
	next   int  // Index to the next rune to be considered
	states []state

	isScriptTag bool
	isStyleTag  bool
}

// Use -1 to indicate the end of the file
const eof = -1

func (l *Lexer) step() {
	codePoint, width := utf8.DecodeRuneInString(l.input[l.next:])
	if width == 0 {
		codePoint = eof
	}
	l.cp = codePoint
	l.end = l.next
	l.next += width
}

func (l *Lexer) text() string {
	return l.input[l.start:l.end]
}

func (l *Lexer) match(valids ...string) bool {
	if l.eof() {
		return false
	}
	for _, valid := range valids {
		if strings.HasPrefix(l.input[l.end:], valid) {
			for range valid {
				l.step()
			}
			return true
		}
	}
	return false
}

func (l *Lexer) stepUntil(delims ...string) {
	for !l.eof() {
		for _, delim := range delims {
			if strings.HasPrefix(l.input[l.end:l.end+len(delim)], delim) {
				return
			}
		}
		l.step()
	}
}

func (l *Lexer) Next() (t token.Token) {
	l.start = l.end
	return l.states[len(l.states)-1](l)
}

func (l *Lexer) pushState(state func(l *Lexer) token.Token) {
	l.states = append(l.states, state)
}

func (l *Lexer) popState() {
	l.states = l.states[:len(l.states)-1]
}

func (l *Lexer) eof() bool {
	return l.cp == eof
}

func (l *Lexer) lexSpace() (t token.Token) {
	for !l.eof() && l.match(" ", "\t", "\n") {
	}
	t.Type = token.Space
	return t
}

func initialState(l *Lexer) (t token.Token) {
	switch l.cp {
	case eof:
		t.Type = token.EndOfInput
		return t
	case '<':
		l.step()
		switch {
		case l.match("!--"):
			l.stepUntil("-->")
			if !l.match("-->") {
				t.Type = token.Unexpected
				return t
			}
			t.Type = token.Comment
			t.Text = l.text()
			return t
		case l.match("/"):
			t.Type = token.LessThanSlash
			l.pushState(closeTagState)
			return t
		case l.match("!"):
			t.Type = token.LessThanExclamation
			l.pushState(tagState)
			return t
		default:
			t.Type = token.LessThan
			l.pushState(tagState)
			return t
		}
	case '{':
		l.step()
		if l.match("#") {
			t.Type = token.OpenCurlyHash
			l.pushState(openBlockState)
			return t
		} else if l.match(":") {
			t.Type = token.OpenCurlyColon
			l.pushState(continueBlockState)
			return t
		} else if l.match("/") {
			t.Type = token.OpenCurlySlash
			l.pushState(closeBlockState)
			return t
		}
		t.Type = token.OpenCurly
		l.pushState(exprState)
		return t
	case ' ', '\t', '\n':
		l.step()
		return l.lexSpace()
	default:
		l.step()
		l.stepUntil("<", "{")
		t.Type = token.Text
		t.Text = l.text()
		return t
	}
}

func (l *Lexer) lexTag() (t token.Token) {
	if !unicode.IsLetter(l.cp) {
		l.step()
		t.Type = token.Unexpected
		return t
	}
	t.Type = token.Identifier
loop:
	for i := 0; ; i++ {
		switch {
		// TODO: ensure components with dashes aren't possible
		case i == 0 && unicode.IsUpper(l.cp):
			l.step()
			t.Type = token.UpperIdentifier
		case i > 0 && l.cp == '-':
			l.step()
			t.Type = token.DashIdentifier
		case i > 0 && l.cp == '.':
			l.step()
			t.Type = token.DotIdentifier
		case i > 0 && l.cp == ':':
			l.step()
			t.Type = token.ColonIdentifier
		case unicode.IsLetter(l.cp) || unicode.IsDigit(l.cp):
			l.step()
			// continue on
		default:
			break loop
		}
	}
	t.Text = l.text()
	return t
}

func tagState(l *Lexer) (t token.Token) {
	switch l.cp {
	case '>':
		l.step()
		l.popState()
		if l.isScriptTag {
			l.pushState(scriptState)
			l.isScriptTag = false
		} else if l.isStyleTag {
			l.pushState(styleState)
			l.isStyleTag = false
		}
		t.Type = token.GreaterThan
		return t
	case '/':
		l.step()
		if l.match(">") {
			l.popState()
			t.Type = token.SlashGreaterThan
			return t
		}
		t.Type = token.Unexpected
		return t
	case '=':
		l.step()
		t.Type = token.Equal
		l.pushState(attributeValueState)
		return t
	case ' ', '\t', '\n':
		l.step()
		return l.lexSpace()
	case '{':
		l.step()
		t.Type = token.OpenCurly
		l.pushState(exprState)
		return t
	case eof:
		t.Type = token.Unexpected
		return t
	default:
		t = l.lexTag()
		switch t.Text {
		case "script":
			l.isScriptTag = true
			// case "style":
			// 	l.isStyleTag = true
		}
		return t
	}
}

func closeTagState(l *Lexer) (t token.Token) {
	switch l.cp {
	case '>':
		l.step()
		l.popState()
		t.Type = token.GreaterThan
		return t
	case ' ', '\t', '\n':
		l.step()
		return l.lexSpace()
	case eof:
		t.Type = token.Unexpected
		return t
	default:
		return l.lexTag()
	}
}

func scriptState(l *Lexer) (t token.Token) {
	inString := false
loop:
	for {
		switch l.cp {
		case eof:
			l.popState()
			t.Type = token.Unexpected
			return t
		case '"':
			l.step()
			inString = !inString
		case '<':
			if !inString && strings.HasPrefix(l.input[l.end:], "</script>") {
				break loop
			}
			l.step()
		default:
			l.step()
		}
	}
	l.popState()
	t.Type = token.Script
	t.Text = l.text()
	return t
}

func styleState(l *Lexer) (t token.Token) {
loop:
	for {
		fmt.Println(l.cp)
		switch l.cp {
		case eof:
			l.popState()
			t.Type = token.Unexpected
			return t
		case '<':
			if strings.HasPrefix(l.input[l.end:], "</style>") {
				break loop
			}
			l.step()
		default:
			l.step()
		}
	}
	fmt.Println("Here...")
	l.popState()
	t.Type = token.Style
	t.Text = l.text()
	return t
}

func exprState(l *Lexer) (t token.Token) {
	switch l.cp {
	case '}':
		l.step()
		l.popState()
		t.Type = token.CloseCurly
		return t
	case eof:
		t.Type = token.Unexpected
		return t
	default:
		depth := 0
	loop:
		for {
			switch l.cp {
			case '{':
				depth++
			case '}':
				if depth == 0 {
					break loop
				}
				depth--
			}
			l.step()
		}
		t.Type = token.Expr
		t.Text = l.text()
		return t
	}
}

// func isEndOfAttributeValue(ch byte) bool {
// 	switch ch {
// 	case ' ', '\t', '\n', '>', 0:
// 		return true
// 	default:
// 		return false
// 	}
// }

// func (l *Lexer) readAttributeValue() string {
// 	position := l.position
// 	for !isEndOfAttributeValue(l.ch) {
// 		l.step()
// 	}
// 	return l.input[position:l.position]
// }

func attributeValueState(l *Lexer) (t token.Token) {
	switch l.cp {
	case '{':
		l.step()
		t.Type = token.OpenCurly
		l.popState()
		l.pushState(exprState)
		return t
	case '"':
		l.step()
		l.popState()
		l.pushState(doubleQuoteState)
		t.Type = token.Quote
		return t
	case '\'':
		l.step()
		l.popState()
		l.pushState(singleQuoteState)
		t.Type = token.Quote
		return t
	case ' ', '\t', '\n':
		l.step()
		return l.lexSpace()
	case eof:
		t.Type = token.Unexpected
		return t
	default:
		l.stepUntil(" ", "\t", "\n", ">")
		l.popState()
		t.Type = token.Text
		t.Text = l.text()
		return t
	}
}

func doubleQuoteState(l *Lexer) (t token.Token) {
	switch l.cp {
	case '"':
		l.step()
		l.popState()
		t.Type = token.Quote
		return t
	case '{':
		l.step()
		l.pushState(exprState)
		t.Type = token.OpenCurly
		return t
	case '\\':
		l.step()
		t.Type = token.BackSlash
		return t
	case eof:
		t.Type = token.Unexpected
		return t
	default:
		l.stepUntil("\"", "{")
		t.Type = token.Text
		t.Text = l.text()
		return t
	}
}

func singleQuoteState(l *Lexer) (t token.Token) {
	switch l.cp {

	case '\'':
		l.step()
		l.popState()
		t.Type = token.Quote
		return t
	case '{':
		l.step()
		l.pushState(exprState)
		t.Type = token.OpenCurly
		return t
	case '\\':
		l.step()
		t.Type = token.BackSlash
		return t
	case eof:
		t.Type = token.Unexpected
		return t
	default:
		l.stepUntil("'", "{")
		t.Type = token.Text
		t.Text = l.text()
		return t
	}
}

func openBlockState(l *Lexer) (t token.Token) {
	switch {
	case l.match("if "):
		l.popState()
		l.pushState(exprState)
		t.Type = token.Identifier
		t.Text = "if"
		return t
	case l.match("each "):
		l.popState()
		l.pushState(eachState)
		l.pushState(eachExprState)
		t.Type = token.Identifier
		t.Text = "each"
		return t
	case l.match("await "):
		l.popState()
		l.pushState(exprState)
		t.Type = token.Identifier
		t.Text = "await"
		return t
	default:
		l.step()
		l.popState()
		t.Type = token.Unexpected
		return t
	}
}

func closeBlockState(l *Lexer) (t token.Token) {
	switch {
	case l.match("if", "each", "await"):
		t.Type = token.Identifier
		t.Text = l.text()
		return t
	case l.cp == '}':
		l.step()
		l.popState()
		t.Type = token.CloseCurly
		return t
	default:
		l.step()
		l.popState()
		t.Type = token.Unexpected
		return t
	}
}

func continueBlockState(l *Lexer) (t token.Token) {
	switch {
	case l.match("else if "):
		l.popState()
		l.pushState(exprState)
		t.Type = token.Identifier
		t.Text = "else if"
		return t
	case l.match("else"):
		l.popState()
		l.pushState(exprState)
		t.Type = token.Identifier
		t.Text = "else"
		return t
	case l.match("then"):
		l.match(" ") // Optionally match a space after then
		l.popState()
		l.pushState(exprState)
		t.Type = token.Identifier
		t.Text = "then"
		return t
	case l.match("catch"):
		l.match(" ") // Optionally match a space after catch
		l.popState()
		l.pushState(exprState)
		t.Type = token.Identifier
		t.Text = "catch"
		return t
	default:
		l.step()
		l.popState()
		t.Type = token.Unexpected
		return t
	}
}

func eachExprState(l *Lexer) (t token.Token) {
	switch l.cp {
	case ' ':
		l.step()
		l.popState()
		t.Type = token.Space
		return t
	case eof:
		l.popState()
		t.Type = token.Unexpected
		return t
	default:
		l.stepUntil(" ")
		t.Type = token.Expr
		t.Text = l.text()
		return t
	}
}

func (l *Lexer) lexVariable() (t token.Token) {
	t.Type = token.Identifier
	for !l.eof() && (unicode.IsLetter(l.cp) || unicode.IsDigit(l.cp) || l.cp == '_') {
		l.step()
	}
	t.Text = l.text()
	return t
}

func eachState(l *Lexer) (t token.Token) {
	switch {
	case l.match(" as "):
		t.Type = token.Identifier
		t.Text = "as"
		return t
	case l.cp == ' ':
		l.step()
		return l.lexSpace()
	case l.cp == eof:
		l.popState()
		t.Type = token.Unexpected
		return t
	case l.cp == '}':
		l.step()
		l.popState()
		t.Type = token.CloseCurly
		return t
	case unicode.IsLetter(l.cp):
		l.step()
		return l.lexVariable()
	case l.cp == ',':
		l.step()
		l.lexSpace()
		t.Type = token.Comma
		return t
	case l.cp == '(':
		l.step()
		l.pushState(eachKeyState)
		t.Type = token.OpenParen
		return t
	default:
		l.step()
		l.popState()
		t.Type = token.Unexpected
		return t
	}
}

func eachKeyState(l *Lexer) (t token.Token) {
	switch l.cp {
	case ')':
		l.step()
		l.popState()
		t.Type = token.CloseParen
		return t
	case eof:
		l.popState()
		t.Type = token.Unexpected
		return t
	default:
		l.stepUntil(")")
		t.Type = token.Expr
		t.Text = l.text()
		return t
	}
}

// func (l *Lexer) lexEachExpr() (t token.Token) {
// 	if !unicode.IsLetter(l.cp) {
// 		l.step()
// 		t.Type = token.Unexpected
// 		return t
// 	}
// 	l.step()
// loop:
// 	for i := 0; l.cp != eof; i++ {
// 		switch {
// 		case l.cp == ' ':
// 			break loop
// 		case unicode.IsLetter(l.cp) || unicode.IsDigit(l.cp) || l.cp == '.' || l.cp == '_':
// 		}
// 	}

// 	// Match the variable name up until the next space
// 	for i := 0; ; i++ {
// 		switch {
// 		case i == 0 && unicode.IsLetter(l.cp):

// 		case i > 0 && l.cp == ' ':
// 			l.step()
// 			l.popState()
// 			l.pushState(eachAsState)
// 			t.Type = token.Space
// 			return t
// 		case l.cp == eof:
// 			t.Type = token.Unexpected
// 			return t
// 		default:
// 			l.step()
// 		}
// 	}
// }

// func (l *Lexer) lexQuote(expected rune) (t token.Token) {
// 	escaped := false
// 	for !l.eof() {
// 		switch l.cp {
// 		// Handle escaped quotes
// 		case '{':
// 			l.step()
// 			l.pushState(exprState)
// 			t.Type = token.OpenCurly
// 			return t
// 		case '\\':
// 			l.step()
// 			escaped = true
// 			continue
// 		case expected:
// 			l.step()
// 			if escaped {
// 				escaped = false
// 				continue
// 			}
// 			t.Type = token.Text
// 			t.Text = l.text()
// 			return t
// 		default:
// 			escaped = false
// 			l.step()
// 		}
// 	}
// 	t.Type = token.Unexpected
// 	return t
// }

// func (l *Lexer) lexBareAttributeValue() (t token.Token) {
// 	t.Type = token.Text
// 	l.stepUntil(" ", "\t", "\n", ">", "}")
// 	t.Text = l.text()
// 	return t
// }
