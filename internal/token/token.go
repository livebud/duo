package token

import "strconv"

type Type uint8

// If you add a new token, remember to add it to "tokenToString" too
const (
	EndOfInput Type = iota
	Unexpected

	LessThan         // <
	GreaterThan      // >
	LessThanSlash    // </
	SlashGreaterThan // />
	BackSlash        // \

	LessThanExclamation // <!

	Equal // =
	Colon // :
	Comma // ,

	Comment // <!-- ... -->

	Space // Any space character

	Identifier      // Any identifier
	UpperIdentifier // Any identifier that starts with a upper case letter
	ColonIdentifier // Any identifier with a colon
	DashIdentifier  // Any identifier with a dash
	DotIdentifier   // Any identifier with a period

	Text   // Raw text
	Expr   // Raw expression to be parsed later
	Script // Script to be parsed later
	Style  // Style to be parsed later

	OpenCurly  // {
	CloseCurly // }

	OpenParen  // (
	CloseParen // )

	OpenCurlyHash  // {#
	OpenCurlyColon // {:
	OpenCurlySlash // {/

	Quote // " or '

)

var tokenToString = map[Type]string{
	EndOfInput:          "end_of_input",
	Unexpected:          "unexpected",
	LessThan:            "less_than",
	LessThanSlash:       "less_than_slash",
	LessThanExclamation: "less_than_exclamation",
	GreaterThan:         "greater_than",
	SlashGreaterThan:    "slash_greater_than",
	BackSlash:           "back_slash",
	Equal:               "equal",
	Colon:               "colon",
	Comma:               "comma",
	Comment:             "comment",
	Space:               "space",
	Identifier:          "identifier",
	UpperIdentifier:     "upper_identifier",
	ColonIdentifier:     "colon_identifier",
	DashIdentifier:      "dash_identifier",
	DotIdentifier:       "dot_identifier",
	Text:                "text",
	Expr:                "expr",
	Script:              "script",
	Style:               "style",
	OpenCurly:           "open_curly",
	CloseCurly:          "close_curly",
	OpenParen:           "open_paren",
	CloseParen:          "close_paren",
	OpenCurlyHash:       "open_curly_hash",
	OpenCurlyColon:      "open_curly_colon",
	OpenCurlySlash:      "open_curly_slash",
	Quote:               "quote",
}

type Token struct {
	Type Type
	Text string
}

func (t *Token) String() string {
	s := tokenToString[t.Type]
	if t.Text != "" {
		s += ":" + strconv.Quote(t.Text)
	}
	return s
}
