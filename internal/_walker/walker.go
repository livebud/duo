package walker

import (
	"fmt"

	"github.com/livebud/duo/internal/ast"
)

type Interface interface {
	EnterDocument(node *ast.Document) (Interface, error)
	ExitDocument(node *ast.Document) error

	EnterElement(node *ast.Element) (Interface, error)
	ExitElement(node *ast.Element) error

	EnterScript(node *ast.Script) (Interface, error)
	ExitScript(node *ast.Script) error

	EnterField(node *ast.Field) (Interface, error)
	ExitField(node *ast.Field) error

	EnterMustache(node *ast.Mustache) (Interface, error)
	ExitMustache(node *ast.Mustache) error

	EnterText(node *ast.Text) (Interface, error)
	ExitText(node *ast.Text) error

	EnterComment(node *ast.Comment) (Interface, error)
	ExitComment(node *ast.Comment) error
}

// Base struct implements the Interface interface with empty methods.
type Base struct {
}

var _ Interface = (*Base)(nil)

func (s Base) EnterDocument(node *ast.Document) (Interface, error) { return s, nil }
func (s Base) ExitDocument(node *ast.Document) error               { return nil }
func (s Base) EnterElement(node *ast.Element) (Interface, error)   { return s, nil }
func (s Base) ExitElement(node *ast.Element) error                 { return nil }
func (s Base) EnterScript(node *ast.Script) (Interface, error)     { return s, nil }
func (s Base) ExitScript(node *ast.Script) error                   { return nil }
func (s Base) EnterField(node *ast.Field) (Interface, error)       { return s, nil }
func (s Base) ExitField(node *ast.Field) error                     { return nil }
func (s Base) EnterMustache(node *ast.Mustache) (Interface, error) { return s, nil }
func (s Base) ExitMustache(node *ast.Mustache) error               { return nil }
func (s Base) EnterText(node *ast.Text) (Interface, error)         { return s, nil }
func (s Base) ExitText(node *ast.Text) error                       { return nil }
func (s Base) EnterComment(node *ast.Comment) (Interface, error)   { return s, nil }
func (s Base) ExitComment(node *ast.Comment) error                 { return nil }

func Walk(node ast.Node, walker Interface) error {
	switch n := node.(type) {
	case *ast.Document:
		if w, err := walker.EnterDocument(n); err != nil {
			return err
		} else if w == nil {
			return nil
		}
		for _, child := range n.Children {
			if err := Walk(child, walker); err != nil {
				return err
			}
		}
		return walker.ExitDocument(n)
	case *ast.Element:
		if w, err := walker.EnterElement(n); err != nil {
			return err
		} else if w == nil {
			return nil
		}
		for _, attr := range n.Attributes {
			if err := Walk(attr, walker); err != nil {
				return err
			}
		}
		for _, child := range n.Children {
			if err := Walk(child, walker); err != nil {
				return err
			}
		}
		return walker.ExitElement(n)
	case *ast.Field:
		if w, err := walker.EnterField(n); err != nil {
			return err
		} else if w == nil {
			return nil
		}
		return walker.ExitField(n)
	case *ast.Mustache:
		if w, err := walker.EnterMustache(n); err != nil {
			return err
		} else if w == nil {
			return nil
		}
		return walker.ExitMustache(n)
	case *ast.Script:
		if w, err := walker.EnterScript(n); err != nil {
			return err
		} else if w == nil {
			return nil
		}
		for _, attr := range n.Attributes {
			if err := Walk(attr, walker); err != nil {
				return err
			}
		}
		return walker.ExitScript(n)
	case *ast.Text:
		if w, err := walker.EnterText(n); err != nil {
			return err
		} else if w == nil {
			return nil
		}
		return walker.ExitText(n)
	default:
		return fmt.Errorf("unknown node type %T", n)
	}
}
