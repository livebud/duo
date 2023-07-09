package ast

import (
	"strings"

	"github.com/tdewolff/parse/v2/js"
)

type Node interface {
	print(indent string) string
}

var (
	_ Node = (*Document)(nil)
	_ Node = (*Element)(nil)
	// _ Node = (*Component)(nil)
	_ Node = (*Field)(nil)
	_ Node = (*Mustache)(nil)
	_ Node = (*Text)(nil)
	_ Node = (*Comment)(nil)
)

type Document struct {
	Children []Fragment
}

func (d *Document) Type() string { return "Document" }

func (d *Document) String() string {
	return d.print("")
}

func (d *Document) print(indent string) string {
	out := new(strings.Builder)
	for _, child := range d.Children {
		out.WriteString(child.print(indent))
		out.WriteByte('\n')
	}
	return out.String()
}

type Fragment interface {
	Node
	fragment()
}

var (
	_ Fragment = (*Element)(nil)
	// _ Fragment = (*Component)(nil)
	_ Fragment = (*Text)(nil)
	_ Fragment = (*Mustache)(nil)
	_ Fragment = (*Comment)(nil)
)

type ElementKind int8

const (
	ElementKindTag ElementKind = iota
	ElementKindComponent
)

type Element struct {
	Name        string
	Kind        ElementKind
	Attributes  []Attribute
	Children    []Fragment
	SelfClosing bool
}

func (e *Element) fragment() {}

func (e *Element) Type() string { return "Element" }

func (e *Element) print(indent string) string {
	out := new(strings.Builder)
	out.WriteString(indent)
	out.WriteString("<")
	out.WriteString(e.Name)
	for _, attr := range e.Attributes {
		out.WriteString(" ")
		out.WriteString(attr.print(" "))
	}
	if e.SelfClosing {
		out.WriteString(" />")
		return out.String()
	}
	out.WriteString(">")
	if len(e.Children) > 0 {
		for _, child := range e.Children {
			out.WriteString(child.print(indent + "\t"))
		}
	}
	out.WriteString(indent)
	out.WriteString("</")
	out.WriteString(e.Name)
	out.WriteString(">")
	return out.String()
}

type Script struct {
	Name       string
	Attributes []Attribute
	Program    *js.AST
}

func (e *Script) fragment() {}

func (e *Script) Type() string { return "Script" }

func (e *Script) print(indent string) string {
	out := new(strings.Builder)
	out.WriteString(indent)
	out.WriteString("<")
	out.WriteString(e.Name)
	for _, attr := range e.Attributes {
		out.WriteString(attr.print(" "))
	}
	out.WriteString(">")
	if e.Program != nil {
		out.WriteByte('\n')
		out.WriteString(indent + "\t")
		out.WriteString(e.Program.JS())
		out.WriteByte('\n')
	}
	out.WriteString(indent)
	out.WriteString("</")
	out.WriteString(e.Name)
	out.WriteString(">")
	return out.String()
}

// type Attribute struct {
// 	Key   string
// 	Value []Value
// }

type Attribute interface {
	Node
	attribute()
}

var (
	_ Attribute = (*Field)(nil)
	_ Attribute = (*AttributeShorthand)(nil)
)

type Field struct {
	Key    string
	Values []Value
}

func (f *Field) attribute() {}

func (f *Field) Type() string { return "Field" }

func (f *Field) print(indent string) string {
	out := new(strings.Builder)
	out.WriteString(f.Key)
	out.WriteString("=")
	out.WriteByte('"')
	for _, v := range f.Values {
		out.WriteString(v.print(""))
	}
	out.WriteByte('"')
	return out.String()
}

type AttributeShorthand struct {
	Key string
}

func (a *AttributeShorthand) attribute() {}

func (a *AttributeShorthand) Type() string { return "AttributeShorthand" }

func (a *AttributeShorthand) print(indent string) string {
	return "{" + a.Key + "}"
}

type Value interface {
	Node
	value()
}

var (
	_ Value = (*Mustache)(nil)
	_ Value = (*Text)(nil)
)

type Mustache struct {
	Expr js.IExpr
}

func (m *Mustache) fragment() {}
func (m *Mustache) value()    {}

func (m *Mustache) Type() string { return "Mustache" }

func (m *Mustache) print(indent string) string {
	return "{" + m.Expr.JS() + "}"
}

type Text struct {
	Value string
}

func (t *Text) fragment() {}
func (t *Text) value()    {}

func (t *Text) Type() string { return "Text" }

func (t *Text) print(indent string) string {
	return t.Value
}

type Comment struct {
}

func (c *Comment) fragment() {}

func (c *Comment) Type() string { return "Comment" }

func (c *Comment) print(indent string) string {
	return ""
}
