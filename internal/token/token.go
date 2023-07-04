package token

import (
	"strconv"
	"strings"
)

type Type string

type Token struct {
	Type  Type
	Text  string
	Start int
	Line  int
}

func (t *Token) String() string {
	s := new(strings.Builder)
	s.WriteString(string(t.Type))
	if t.Text != "" && t.Text != string(t.Type) {
		s.WriteString(":")
		s.WriteString(strconv.Quote(t.Text))
	}
	return s.String()
}

const (
	End              Type = "end"
	Error            Type = "error"
	LessThan         Type = "<"  // <
	GreaterThan      Type = ">"  // >
	Slash            Type = "/"  // /
	LessThanSlash    Type = "</" // </
	SlashGreaterThan Type = "/>" // />
	BackSlash        Type = "\\" // \

	Doctype Type = "<!doctype" // <!doctype

	Equal Type = "=" // =
	Colon Type = ":" // :
	Comma Type = "," // ,

	Comment Type = "comment" // <!-- ... -->

	Space Type = "space" // Any space character

	Identifier       Type = "identifier"        // Any identifier
	PascalIdentifier Type = "pascal_identifier" // Any identifier that starts with a upper case letter
	ColonIdentifier  Type = "colon_identifier"  // Any identifier with a colon
	DashIdentifier   Type = "dash_identifier"   // Any identifier with a dash
	DotIdentifier    Type = "dot_identifier"    // Any identifier with a period

	Text   Type = "text"   // Raw text
	Expr   Type = "expr"   // Raw expression to be parsed later
	Script Type = "script" // Script to be parsed later
	Style  Type = "style"  // Style to be parsed later

	LeftBrace  Type = "{" // {
	RightBrace Type = "}" // }

	// OpenParen  Type = "(" // (
	// CloseParen Type = ")" // )

	// OpenCurlyHash  Type = "{#" // {#
	// OpenCurlyColon Type = "{:" // {:
	// OpenCurlySlash Type = "{/" // {/

	Quote Type = "quote" // " or '

	// Pipe Type = "|" // |
)
