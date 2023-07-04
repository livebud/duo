package evaluator

import (
	"fmt"
	"io"
	"reflect"

	"github.com/livebud/duo/internal/ast"
	"github.com/tdewolff/parse/v2/js"
)

func New(doc *ast.Document) *Evaluator {
	return &Evaluator{doc}
}

type Evaluator struct {
	doc *ast.Document
}

func (e *Evaluator) Evaluate(w io.Writer, v interface{}) error {
	value := reflect.ValueOf(v)
	evaluator := &evaluator{&writer{w}}
	if err := evaluator.evaluateDocument(value, e.doc); err != nil {
		return err
	}
	return nil
}

type evaluator struct {
	*writer
}

type writer struct {
	io.Writer
}

func (w *writer) WriteRune(r rune) (int, error) {
	return w.Writer.Write([]byte(string(r)))
}

func (w *writer) WriteString(s string) (int, error) {
	return w.Writer.Write([]byte(s))
}

func (w *writer) WriteByte(b byte) error {
	_, err := w.Writer.Write([]byte{b})
	return err
}

func (e *evaluator) evaluateDocument(scope reflect.Value, node *ast.Document) error {
	for _, child := range node.Children {
		if err := e.evaluateFragment(scope, child); err != nil {
			return err
		}
	}
	return nil
}

func (e *evaluator) evaluateFragment(scope reflect.Value, node ast.Fragment) error {
	switch n := node.(type) {
	case *ast.Element:
		return e.evaluateElement(scope, n)
	case *ast.Text:
		return e.evaluateText(scope, n)
	case *ast.Mustache:
		return e.evaluateMustache(scope, n)
	case *ast.Script:
		return e.evaluateScript(scope, n)
	default:
		return fmt.Errorf("unknown fragment %T", n)
	}
}

func (e *evaluator) evaluateElement(scope reflect.Value, node *ast.Element) error {
	e.WriteByte('<')
	e.WriteString(node.Name)
	for _, attr := range node.Attributes {
		if err := e.evaluateAttribute(scope, attr); err != nil {
			return err
		}
	}
	if node.SelfClosing {
		e.WriteString("/>")
		return nil
	}
	e.WriteString(">")
	for _, child := range node.Children {
		if err := e.evaluateFragment(scope, child); err != nil {
			return err
		}
	}
	e.WriteString("</")
	e.WriteString(node.Name)
	e.WriteString(">")
	return nil
}

func (e *evaluator) evaluateScript(scope reflect.Value, node *ast.Script) error {
	return nil
}

func (e *evaluator) evaluateAttribute(scope reflect.Value, node ast.Attribute) error {
	switch n := node.(type) {
	case *ast.Field:
		return e.evaluateField(scope, n)
	default:
		return fmt.Errorf("unknown attribute %T", n)
	}
}

func (e *evaluator) evaluateField(scope reflect.Value, node *ast.Field) error {
	e.WriteString(node.Key)
	e.WriteString("=\"")
	e.WriteString(node.Value.JS())
	return nil
}

func (e *evaluator) evaluateText(scope reflect.Value, node *ast.Text) error {
	e.WriteString(node.Value)
	return nil
}

func (e *evaluator) evaluateMustache(scope reflect.Value, node *ast.Mustache) error {
	return e.evaluateExpr(scope, node.Expr)
}

func (e *evaluator) evaluateExpr(scope reflect.Value, node js.IExpr) error {
	switch n := node.(type) {
	case *js.LiteralExpr:
		return e.evaluateLiteralExpr(scope, n)
	case *js.Var:
		return e.evaluateVar(scope, n)
	// case *js.Identifier:
	// 	return e.evaluateIdentifierExpr(scope, n)
	// case *js.
	// case *js.MemberExpr:
	// 	return e.evaluateMemberExpr(scope, n)
	// case *js.CallExpr:
	// 	return e.evaluateCallExpr(scope, n)
	// case *js.BinaryExpr:
	// 	return e.evaluateBinaryExpr(scope, n)
	// case *js.UnaryExpr:
	// 	return e.evaluateUnaryExpr(scope, n)
	// case *js.ConditionalExpr:
	// 	return e.evaluateConditionalExpr(scope, n)
	// case *js.ArrayExpr:
	// 	return e.evaluateArrayExpr(scope, n)
	// case *js.ObjectExpr:
	// 	return e.evaluateObjectExpr(scope, n)
	// case *js.FunctionExpr:
	// 	return e.evaluateFunctionExpr(scope, n)
	// case *js.TemplateExpr:
	// 	return e.evaluateTemplateExpr(scope, n)
	// case *js.TaggedTemplateExpr:
	// 	return e.evaluateTaggedTemplateExpr(scope, n)
	// case *js.ParenExpr:
	// 	return e.evaluateParenExpr(scope, n)
	default:
		return fmt.Errorf("unknown expression %T", n)
	}
}

func (e *evaluator) evaluateVar(scope reflect.Value, node *js.Var) error {
	switch node.Decl {
	case js.NoDecl:
		// Since there's no parse expression function, identifiers are considered a non-declared variable
		return e.evaluateIdentifier(scope, &js.LiteralExpr{
			Data:      node.Data,
			TokenType: js.IdentifierToken,
		})
	default:
		return fmt.Errorf("unknown decl %s", node.Decl.String())
	}
}

func (e *evaluator) evaluateIdentifier(scope reflect.Value, node *js.LiteralExpr) error {
	// Handles nil
	if !scope.IsValid() {
		return nil
	}
	switch scope.Kind() {
	case reflect.Map:
		value := scope.MapIndex(reflect.ValueOf(string(node.Data)))
		if !value.IsValid() {
			// Don't write anything
			return nil
		}
		return e.writeValue(value.Interface())
	default:
		return fmt.Errorf("unexpected scope type %s", scope.Kind().String())
	}
}

func (e *evaluator) writeValue(v interface{}) error {
	switch value := v.(type) {
	case string:
		e.WriteString(value)
		return nil
	default:
		return fmt.Errorf("unexpected value %T", value)
	}
}

func (e *evaluator) evaluateLiteralExpr(scope reflect.Value, node *js.LiteralExpr) error {
	switch node.TokenType {
	case js.IdentifierToken:
		e.Write(node.Data)
		return nil
	case js.StringToken:
		e.Write(node.Data)
		return nil
	// case js.NumericToken:
	// 	e.Write(node.Data)
	// case js.RegexpLiteralToken:
	// 	e.Write(node.Data)
	// case js.TrueToken:
	// 	e.WriteString("true")
	// case js.FalseToken:
	// 	e.WriteString("false")
	// case js.NullToken:
	// 	e.WriteString("null")
	// case js.UndefinedToken:
	// 	e.WriteString("undefined")
	// case js.NaNToken:
	// 	e.WriteString("NaN")
	// case js.InfinityToken:
	// 	e.WriteString("Infinity")
	default:
		return fmt.Errorf("unknown literal %s", node.TokenType.String())
	}
}

// var _ walker.Interface = (*evaluation)(nil)

// func (e *evaluation) EnterDocument(node *ast.Document) (walker.Interface, error) {
// 	return e, nil
// }

// func (e *evaluation) EnterText(node *ast.Text) (walker.Interface, error) {
// 	return e, nil
// }

// func (e *evaluation) EnterMustache(node *ast.Mustache) (walker.Interface, error) {
// 	return e, nil
// }
