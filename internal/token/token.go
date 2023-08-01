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
	EOF              Type = "eof"
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

	Identifier       Type = "identifier"        // Any identifier
	PascalIdentifier Type = "pascal_identifier" // Any identifier that starts with a upper case letter
	ColonIdentifier  Type = "colon_identifier"  // Any identifier with a colon
	DashIdentifier   Type = "dash_identifier"   // Any identifier with a dash
	DotIdentifier    Type = "dot_identifier"    // Any identifier with a period

	Text   Type = "text"   // Raw text
	Expr   Type = "expr"   // Raw expression to be parsed later
	Script Type = "script" // Script to be parsed later
	Style  Type = "style"  // Style to be parsed later
	Slot   Type = "slot"   // Slot to be parsed later

	LeftBrace  Type = "{" // {
	RightBrace Type = "}" // }

	If     Type = "if"      // if
	For    Type = "for"     // for
	In     Type = "in"      // in
	ElseIf Type = "else_if" // elseif
	Else   Type = "else"    // else
	End    Type = "end"     // end

	Quote Type = "quote" // " or '
)
