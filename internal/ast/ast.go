package ast

import (
	"strings"

	"github.com/livebud/duo/internal/scope"
	css "github.com/matthewmueller/css/ast"
	"github.com/tdewolff/parse/v2/js"
)

type Node interface {
	print(indent string) string
}

var (
	_ Node = (*Document)(nil)
	_ Node = (*Element)(nil)
	_ Node = (*Component)(nil)
	_ Node = (*Script)(nil)
	_ Node = (*AttributeShorthand)(nil)
	_ Node = (*Field)(nil)
	_ Node = (*Mustache)(nil)
	_ Node = (*Text)(nil)
	_ Node = (*Comment)(nil)
	_ Node = (*IfBlock)(nil)
)

type Document struct {
	Children []Fragment
	Scope    *scope.Scope
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

// Style returns the style node if it exists.
func (d *Document) Style() (*Style, bool) {
	for _, child := range d.Children {
		if style, ok := child.(*Style); ok {
			return style, true
		}
	}
	return nil, false
}

// Script returns the script node if it exists.
func (d *Document) Script() (*Script, bool) {
	for _, child := range d.Children {
		if script, ok := child.(*Script); ok {
			return script, true
		}
	}
	return nil, false
}

type Fragment interface {
	Node
	fragment()
}

var (
	_ Fragment = (*Element)(nil)
	_ Fragment = (*Component)(nil)
	_ Fragment = (*Text)(nil)
	_ Fragment = (*Mustache)(nil)
	_ Fragment = (*Comment)(nil)
	_ Fragment = (*IfBlock)(nil)
)

type Element struct {
	Name        string
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

type Style struct {
	Attributes  []Attribute
	SelfClosing bool
	StyleSheet  *css.Stylesheet
}

func (e *Style) fragment() {}

func (e *Style) Type() string { return "Style" }

func (e *Style) print(indent string) string {
	out := new(strings.Builder)
	out.WriteString(indent)
	out.WriteString("<style")
	for _, attr := range e.Attributes {
		out.WriteString(attr.print(" "))
	}
	out.WriteString(">")
	if e.StyleSheet != nil {
		out.WriteByte('\n')
		out.WriteString(indent + "\t")
		out.WriteString(e.StyleSheet.String())
		out.WriteByte('\n')
	}
	out.WriteString(indent)
	out.WriteString("</style>")
	return out.String()
}

type Script struct {
	Attributes  []Attribute
	SelfClosing bool
	Program     *js.AST
}

func (e *Script) fragment() {}

func (e *Script) Type() string { return "Script" }

func (e *Script) print(indent string) string {
	out := new(strings.Builder)
	out.WriteString(indent)
	out.WriteString("<script")
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
	out.WriteString("</script>")
	return out.String()
}

type Component struct {
	Name        string
	Attributes  []Attribute
	Children    []Fragment
	SelfClosing bool
}

func (c *Component) fragment() {}

func (c *Component) Type() string { return "Component" }

func (c *Component) print(indent string) string {
	out := new(strings.Builder)
	out.WriteString(indent)
	out.WriteString("<")
	out.WriteString(c.Name)
	for _, attr := range c.Attributes {
		out.WriteString(" ")
		out.WriteString(attr.print(" "))
	}
	if c.SelfClosing {
		out.WriteString(" />")
		return out.String()
	}
	out.WriteString(">")
	if len(c.Children) > 0 {
		for _, child := range c.Children {
			out.WriteString(child.print(indent + "\t"))
		}
	}
	out.WriteString(indent)
	out.WriteString("</")
	out.WriteString(c.Name)
	out.WriteString(">")
	return out.String()
}

type Slot struct {
	Name        string // Empty for the default slot
	SelfClosing bool
	Fallback    []Fragment
}

func (s *Slot) fragment() {}

func (s *Slot) Type() string { return "Slot" }

func (s *Slot) print(indent string) string {
	out := new(strings.Builder)
	out.WriteString(indent)
	out.WriteString("<slot")
	if s.Name != "" {
		out.WriteString(" name=")
		out.WriteByte('"')
		out.WriteString(s.Name)
		out.WriteByte('"')
	}
	if s.SelfClosing {
		out.WriteString(" />")
		return out.String()
	}
	out.WriteString(">")
	if len(s.Fallback) > 0 {
		for _, child := range s.Fallback {
			out.WriteString(child.print(indent + "\t"))
		}
	}
	out.WriteString(indent)
	out.WriteString("</slot>")
	return out.String()
}

type Attribute interface {
	Node
	GetKey() string
	attribute()
}

var (
	_ Attribute = (*Field)(nil)
	_ Attribute = (*Binding)(nil)
	_ Attribute = (*AttributeShorthand)(nil)
	_ Attribute = (*NamedSlot)(nil)
)

type Field struct {
	Key          string
	Values       []Value
	EventHandler bool
}

func (f *Field) GetKey() string {
	return f.Key
}

func (f *Field) attribute() {}

func (f *Field) Type() string { return "Field" }

func (f *Field) print(indent string) string {
	out := new(strings.Builder)
	out.WriteString(f.Key)
	out.WriteString("=")
	if f.EventHandler && len(f.Values) > 0 {
		out.WriteString(f.Values[0].print(""))
		return out.String()
	}
	out.WriteByte('"')
	for _, v := range f.Values {
		out.WriteString(v.print(""))
	}
	out.WriteByte('"')
	return out.String()
}

type Binding struct {
	Key   string
	Value Value
}

func (f *Binding) attribute() {}

func (f *Binding) GetKey() string {
	return f.Key
}

func (f *Binding) Type() string { return "Binding" }

func (f *Binding) print(indent string) string {
	out := new(strings.Builder)
	out.WriteString("bind:")
	out.WriteString(f.Key)
	out.WriteString("=")
	out.WriteString(f.Value.print(""))
	return out.String()
}

type Class struct {
	Name  string
	Value Value
}

func (c *Class) attribute() {}

func (c *Class) GetKey() string {
	return "class"
}

func (c *Class) Type() string { return "Class" }

func (c *Class) print(indent string) string {
	out := new(strings.Builder)
	out.WriteString("class:")
	out.WriteString(c.Name)
	out.WriteString("=")
	out.WriteString(c.Value.print(""))
	return out.String()
}

type AttributeShorthand struct {
	Key          string
	EventHandler bool
}

func (a *AttributeShorthand) attribute() {}

func (a *AttributeShorthand) GetKey() string { return a.Key }

func (a *AttributeShorthand) Type() string { return "AttributeShorthand" }

func (a *AttributeShorthand) print(indent string) string {
	return "{" + a.Key + "}"
}

// NamedSlot is a named slot field.
type NamedSlot struct {
	Name string
}

func (s *NamedSlot) attribute() {}

func (s *NamedSlot) Type() string { return "NamedSlot" }

func (s *NamedSlot) GetKey() string {
	return "slot"
}

func (s *NamedSlot) print(indent string) string {
	return s.GetKey() + `="` + s.Name + `"`
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
	Value string
}

func (c *Comment) fragment() {}

func (c *Comment) Type() string { return "Comment" }

func (c *Comment) print(indent string) string {
	return indent + c.Value
}

type IfBlock struct {
	Cond js.IExpr
	Then []Fragment
	Else []Fragment
}

func (i *IfBlock) fragment() {}

func (i *IfBlock) Type() string { return "IfBlock" }

func (i *IfBlock) print(indent string) string {
	out := new(strings.Builder)
	out.WriteString(indent)
	out.WriteString("{#if ")
	out.WriteString(i.Cond.JS())
	out.WriteString("}")
	for _, child := range i.Then {
		out.WriteString(child.print(indent + "\t"))
		out.WriteByte('\n')
	}
	if len(i.Else) > 0 {
		out.WriteString("{:else}")
		for _, child := range i.Else {
			out.WriteString(child.print(indent + "\t"))
			out.WriteByte('\n')
		}
	}
	out.WriteString(indent)
	out.WriteString("{/if}")
	return out.String()
}

type EachBlock struct {
	Key   *js.Var // Can be nil
	Value *js.Var // Can be nil
	List  js.IExpr
	Body  []Fragment
	Else  []Fragment
}

func (f *EachBlock) fragment() {}

func (f *EachBlock) Type() string { return "EachBlock" }

func (f *EachBlock) print(indent string) string {
	out := new(strings.Builder)
	out.WriteString(indent)
	out.WriteString("{#each ")
	out.WriteString(f.List.JS())
	if f.Value != nil {
		out.WriteString(" as ")
		out.WriteString(string(f.Value.JS()))
	}
	if f.Key != nil {
		out.WriteString(", ")
		out.WriteString(string(f.Key.JS()))
	}
	out.WriteString("}")
	for _, child := range f.Body {
		out.WriteString(child.print(indent + "\t"))
		out.WriteByte('\n')
	}
	if len(f.Else) > 0 {
		out.WriteString("{else}")
		for _, child := range f.Else {
			out.WriteString(child.print(indent + "\t"))
			out.WriteByte('\n')
		}
	}
	out.WriteString(indent)
	out.WriteString("{/each}")
	return out.String()
}
