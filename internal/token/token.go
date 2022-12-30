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

	LessThanExclamation // <!

	Equal // =
	Colon // :

	OpenComment  // <!--
	CloseComment // -->

	Space // Any space character

	Identifier      // Any identifier
	UpperIdentifier // Any identifier that starts with a upper case letter
	ColonIdentifier // Any identifier with a colon
	DashIdentifier  // Any identifier with a dash
	DotIdentifier   // Any identifier with a period

	Text // Text that is not a token

	OpenCurly  // {
	CloseCurly // }

	Quote // " or '
)

var tokenToString = map[Type]string{
	EndOfInput:          "EndOfInput",
	Unexpected:          "Unexpected",
	LessThan:            "LessThan",
	LessThanSlash:       "LessThanSlash",
	LessThanExclamation: "LessThanExclamation",
	GreaterThan:         "GreaterThan",
	SlashGreaterThan:    "SlashGreaterThan",
	Equal:               "Equal",
	Colon:               "Colon",
	OpenComment:         "OpenComment",
	CloseComment:        "CloseComment",
	Space:               "Space",
	Identifier:          "Identifier",
	UpperIdentifier:     "UpperIdentifier",
	ColonIdentifier:     "ColonIdentifier",
	DashIdentifier:      "DashIdentifier",
	DotIdentifier:       "DotIdentifier",
	Text:                "Text",
	OpenCurly:           "OpenCurly",
	CloseCurly:          "CloseCurly",
	Quote:               "Quote",
}

func New(t Type, literal string) Token {
	return Token{
		Type:    t,
		Literal: literal,
	}
}

type Token struct {
	Type    Type
	Literal string
}

func (t *Token) String() string {
	s := tokenToString[t.Type]
	if t.Literal != "" {
		s += ":" + strconv.Quote(t.Literal)
	}
	return s
}
