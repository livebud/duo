package evaluator

import (
	"bytes"
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
	evaluator := &evaluator{}
	if err := evaluator.evaluateDocument(&ioWriter{w}, value, e.doc); err != nil {
		return err
	}
	return nil
}

type evaluator struct {
}

type ioWriter struct {
	io.Writer
}

func (w *ioWriter) WriteRune(r rune) (int, error) {
	return w.Writer.Write([]byte(string(r)))
}

func (w *ioWriter) WriteString(s string) (int, error) {
	return w.Writer.Write([]byte(s))
}

func (w *ioWriter) WriteByte(b byte) error {
	_, err := w.Writer.Write([]byte{b})
	return err
}

type writer interface {
	io.Writer
	WriteRune(r rune) (int, error)
	WriteString(s string) (int, error)
	WriteByte(b byte) error
}

func (e *evaluator) evaluateDocument(w writer, scope reflect.Value, node *ast.Document) error {
	for _, child := range node.Children {
		if err := e.evaluateFragment(w, scope, child); err != nil {
			return err
		}
	}
	return nil
}

func (e *evaluator) evaluateFragment(w writer, scope reflect.Value, node ast.Fragment) error {
	switch n := node.(type) {
	case *ast.Element:
		return e.evaluateElement(w, scope, n)
	case *ast.Text:
		return e.evaluateText(w, scope, n)
	case *ast.Mustache:
		return e.evaluateMustache(w, scope, n)
	case *ast.Script:
		return e.evaluateScript(w, scope, n)
	default:
		return fmt.Errorf("unknown fragment %T", n)
	}
}

func (e *evaluator) evaluateElement(w writer, scope reflect.Value, node *ast.Element) error {
	w.WriteByte('<')
	w.WriteString(node.Name)
	for _, attr := range node.Attributes {
		buf := new(bytes.Buffer)
		if err := e.evaluateAttribute(buf, scope, attr); err != nil {
			return err
		}
		if buf.Len() > 0 {
			w.WriteByte(' ')
			w.Write(buf.Bytes())
		}
	}
	if node.SelfClosing {
		w.WriteString("/>")
		return nil
	}
	w.WriteString(">")
	for _, child := range node.Children {
		if err := e.evaluateFragment(w, scope, child); err != nil {
			return err
		}
	}
	w.WriteString("</")
	w.WriteString(node.Name)
	w.WriteString(">")
	return nil
}

func (e *evaluator) evaluateScript(w writer, scope reflect.Value, node *ast.Script) error {
	return nil
}

func (e *evaluator) evaluateAttribute(w writer, scope reflect.Value, node ast.Attribute) error {
	switch n := node.(type) {
	case *ast.Field:
		return e.evaluateField(w, scope, n)
	case *ast.AttributeShorthand:
		return e.evaluateAttributeShorthand(w, scope, n)
	default:
		return fmt.Errorf("unknown attribute %T", n)
	}
}

func (e *evaluator) evaluateField(w writer, scope reflect.Value, node *ast.Field) error {
	buf := new(bytes.Buffer)
	for _, value := range node.Values {
		if err := e.evaluateValue(buf, scope, value); err != nil {
			return err
		}
	}
	if buf.Len() == 0 {
		return nil
	}
	w.WriteString(node.Key)
	if len(node.Values) > 0 {
		w.WriteByte('=')
		w.WriteByte('"')
		w.Write(buf.Bytes())
		w.WriteByte('"')
	}
	return nil
}

func (e *evaluator) evaluateAttributeShorthand(w writer, scope reflect.Value, node *ast.AttributeShorthand) error {
	buf := new(bytes.Buffer)
	if err := e.evaluateIdentifier(buf, scope, &js.LiteralExpr{
		Data:      []byte(node.Key),
		TokenType: js.IdentifierToken,
	}); err != nil {
		return err
	}
	if buf.Len() == 0 {
		return nil
	}
	w.WriteString(node.Key)
	w.WriteByte('=')
	w.WriteByte('"')
	w.Write(buf.Bytes())
	w.WriteByte('"')
	return nil
}

func (e *evaluator) evaluateValue(w writer, scope reflect.Value, node ast.Value) error {
	switch n := node.(type) {
	case *ast.Text:
		return e.evaluateText(w, scope, n)
	case *ast.Mustache:
		return e.evaluateMustache(w, scope, n)
	default:
		return fmt.Errorf("unknown attribute value %T", n)
	}
}

func (e *evaluator) evaluateText(w writer, scope reflect.Value, node *ast.Text) error {
	w.WriteString(node.Value)
	return nil
}

func (e *evaluator) evaluateMustache(w writer, scope reflect.Value, node *ast.Mustache) error {
	return e.evaluateExpr(w, scope, node.Expr)
}

func (e *evaluator) evaluateExpr(w writer, scope reflect.Value, node js.IExpr) error {
	switch n := node.(type) {
	case *js.LiteralExpr:
		return e.evaluateLiteralExpr(w, scope, n)
	case *js.Var:
		return e.evaluateVar(w, scope, n)
	// case *js.Identifier:
	// 	return e.evaluateIdentifierExpr(w,scope, n)
	// case *js.
	// case *js.MemberExpr:
	// 	return e.evaluateMemberExpr(w,scope, n)
	// case *js.CallExpr:
	// 	return e.evaluateCallExpr(w,scope, n)
	// case *js.BinaryExpr:
	// 	return e.evaluateBinaryExpr(w,scope, n)
	// case *js.UnaryExpr:
	// 	return e.evaluateUnaryExpr(w,scope, n)
	// case *js.ConditionalExpr:
	// 	return e.evaluateConditionalExpr(w,scope, n)
	// case *js.ArrayExpr:
	// 	return e.evaluateArrayExpr(w,scope, n)
	// case *js.ObjectExpr:
	// 	return e.evaluateObjectExpr(w,scope, n)
	// case *js.FunctionExpr:
	// 	return e.evaluateFunctionExpr(w,scope, n)
	// case *js.TemplateExpr:
	// 	return e.evaluateTemplateExpr(w,scope, n)
	// case *js.TaggedTemplateExpr:
	// 	return e.evaluateTaggedTemplateExpr(w,scope, n)
	// case *js.ParenExpr:
	// 	return e.evaluateParenExpr(w,scope, n)
	default:
		return fmt.Errorf("unknown expression %T", n)
	}
}

func (e *evaluator) evaluateVar(w writer, scope reflect.Value, node *js.Var) error {
	switch node.Decl {
	case js.NoDecl:
		// Since there's no parse expression function, identifiers are considered a non-declared variable
		return e.evaluateIdentifier(w, scope, &js.LiteralExpr{
			Data:      node.Data,
			TokenType: js.IdentifierToken,
		})
	default:
		return fmt.Errorf("unknown decl %s", node.Decl.String())
	}
}

func (e *evaluator) evaluateIdentifier(w writer, scope reflect.Value, node *js.LiteralExpr) error {
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
		return e.writeValue(w, value.Interface())
	default:
		return fmt.Errorf("unexpected scope type %s", scope.Kind().String())
	}
}

func (e *evaluator) writeValue(w writer, v interface{}) error {
	switch value := v.(type) {
	case string:
		w.WriteString(value)
		return nil
	default:
		return fmt.Errorf("unexpected value %T", value)
	}
}

func (e *evaluator) evaluateLiteralExpr(w writer, scope reflect.Value, node *js.LiteralExpr) error {
	switch node.TokenType {
	case js.IdentifierToken:
		w.Write(node.Data)
		return nil
	case js.StringToken:
		w.Write(node.Data)
		return nil
	// case js.NumericToken:
	// 	w.Write(node.Data)
	// case js.RegexpLiteralToken:
	// 	w.Write(node.Data)
	// case js.TrueToken:
	// 	w.WriteString("true")
	// case js.FalseToken:
	// 	w.WriteString("false")
	// case js.NullToken:
	// 	w.WriteString("null")
	// case js.UndefinedToken:
	// 	w.WriteString("undefined")
	// case js.NaNToken:
	// 	w.WriteString("NaN")
	// case js.InfinityToken:
	// 	w.WriteString("Infinity")
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
