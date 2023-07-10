package evaluator

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

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
	// Skip event handlers
	if node.EventHandler {
		return nil
	}
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
	// buf := new(bytes.Buffer)
	// Skip event handlers
	if node.EventHandler {
		return nil
	}
	value, err := evaluateExpr(scope, &js.LiteralExpr{
		Data:      []byte(node.Key),
		TokenType: js.IdentifierToken,
	})
	if err != nil {
		return err
	} else if !value.IsValid() {
		return nil
	}
	w.WriteString(node.Key)
	w.WriteByte('=')
	w.WriteByte('"')
	if err := e.writeValue(w, value); err != nil {
		return err
	}
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
	value, err := evaluateExpr(scope, node.Expr)
	if err != nil {
		return err
	}
	return e.writeValue(w, value)
}

func (e *evaluator) writeValue(w writer, value reflect.Value) error {
	if !value.IsValid() {
		return nil
	}
	v := value.Interface()
	switch value := v.(type) {
	case string:
		w.WriteString(value)
		return nil
	case int64:
		w.WriteString(strconv.FormatInt(value, 10))
		return nil
	case int:
		w.WriteString(strconv.Itoa(value))
		return nil
	default:
		return fmt.Errorf("unexpected value %T", value)
	}
}

func evaluateExpr(scope reflect.Value, node js.IExpr) (reflect.Value, error) {
	switch n := node.(type) {
	case *js.LiteralExpr:
		return evaluateLiteralExpr(scope, n)
	case *js.Var:
		return evaluateVar(scope, n)
	case *js.BinaryExpr:
		return evaluateBinaryExpr(scope, n)
	case *js.CondExpr:
		return evaluateCondExpr(scope, n)
	default:
		return reflect.Value{}, fmt.Errorf("unknown expression %T", n)
	}
}

func evaluateLiteralExpr(scope reflect.Value, node *js.LiteralExpr) (reflect.Value, error) {
	switch node.TokenType {
	case js.IdentifierToken:
		return evaluateIdentifier(scope, node)
	case js.StringToken:
		// TODO: more robust unquoting.
		// strconv.Unquote doesn't support multi-char single quotes
		value := strings.Trim(string(node.Data), `'"`)
		return reflect.ValueOf(value), nil
	case js.DecimalToken:
		// TODO: handle floats
		// TODO: make sure 64b is right
		n, err := strconv.ParseInt(string(node.Data), 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(n), nil
	case js.TrueToken:
		return reflect.ValueOf(true), nil
	case js.FalseToken:
		return reflect.ValueOf(false), nil
	default:
		return reflect.Value{}, fmt.Errorf("unknown literal expression %s", node.TokenType.String())
	}
}

func evaluateVar(scope reflect.Value, node *js.Var) (reflect.Value, error) {
	switch node.Decl {
	case js.NoDecl:
		// Since there's no parse expression function, identifiers are considered a non-declared variable
		return evaluateIdentifier(scope, &js.LiteralExpr{
			Data:      node.Data,
			TokenType: js.IdentifierToken,
		})
	default:
		return reflect.Value{}, fmt.Errorf("unknown decl %s", node.Decl.String())
	}
}

func evaluateBinaryExpr(scope reflect.Value, node *js.BinaryExpr) (reflect.Value, error) {
	left, err := evaluateExpr(scope, node.X)
	if err != nil {
		return reflect.Value{}, err
	}
	// if !left.IsValid() {
	// 	return reflect.Value{}, fmt.Errorf("%s is undefined", node.X.JS())
	// }
	if left.Kind() == reflect.Interface {
		left = left.Elem()
	}
	right, err := evaluateExpr(scope, node.Y)
	if err != nil {
		return reflect.Value{}, err
	}
	// if !right.IsValid() {
	// 	return reflect.Value{}, fmt.Errorf("%s is undefined", node.Y.JS())
	// }
	if right.Kind() == reflect.Interface {
		right = right.Elem()
	}
	switch node.Op {
	case js.AddToken:
		return evaluateAdd(scope, left, right)
	case js.EqEqEqToken:
		return evaluateStrictEqual(scope, left, right)
	case js.OrToken:
		return evaluateOr(scope, left, right)
	default:
		return reflect.Value{}, fmt.Errorf("unknown binary expression %s", node.Op.String())
	}
}

func evaluateCondExpr(scope reflect.Value, node *js.CondExpr) (reflect.Value, error) {
	cond, err := evaluateExpr(scope, node.Cond)
	if err != nil {
		return reflect.Value{}, err
	}
	left := true
	switch cond.Kind() {
	case reflect.Bool:
		left = cond.Bool()
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int8, reflect.Int16:
		left = cond.Int() != 0
	case reflect.String:
		left = cond.String() != ""
	}
	if left {
		return evaluateExpr(scope, node.X)
	}
	return evaluateExpr(scope, node.Y)
}

func evaluateAdd(scope reflect.Value, left, right reflect.Value) (reflect.Value, error) {
	switch left.Kind() {
	case reflect.String:
		switch right.Kind() {
		case reflect.String:
			return reflect.ValueOf(left.String() + right.String()), nil
		default:
			return reflect.Value{}, fmt.Errorf("unexpected right value %s", right.Kind().String())
		}
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int8, reflect.Int16:
		switch right.Kind() {
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int8, reflect.Int16:
			return reflect.ValueOf(left.Int() + right.Int()), nil
		default:
			return reflect.Value{}, fmt.Errorf("unexpected right value %s", right.Kind().String())
		}
	default:
		return reflect.Value{}, fmt.Errorf("unexpected left value %s", left.String())
	}
}

func evaluateOr(scope reflect.Value, left, right reflect.Value) (reflect.Value, error) {
	switch left.Kind() {
	case reflect.Bool:
		if left.Bool() {
			return left, nil
		}
		return right, nil
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int8, reflect.Int16:
		if left.Int() != 0 {
			return left, nil
		}
		return right, nil
	case reflect.String:
		if left.String() != "" {
			return left, nil
		}
		return right, nil
	case reflect.Invalid:
		return right, nil
	default:
		return reflect.Value{}, fmt.Errorf("unexpected left value %s", left.String())
	}
}

func evaluateStrictEqual(scope reflect.Value, left, right reflect.Value) (reflect.Value, error) {
	switch left.Kind() {
	case reflect.String:
		switch right.Kind() {
		case reflect.String:
			return reflect.ValueOf(left.String() == right.String()), nil
		default:
			return reflect.Value{}, fmt.Errorf("unexpected right value %s", right.Kind().String())
		}
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int8, reflect.Int16:
		switch right.Kind() {
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int8, reflect.Int16:
			return reflect.ValueOf(left.Int() == right.Int()), nil
		default:
			return reflect.Value{}, fmt.Errorf("unexpected right value %s", right.Kind().String())
		}
	case reflect.Bool:
		switch right.Kind() {
		case reflect.Bool:
			return reflect.ValueOf(left.Bool() == right.Bool()), nil
		default:
			return reflect.Value{}, fmt.Errorf("unexpected right value %s", right.Kind().String())
		}
	case reflect.Invalid:
		switch right.Kind() {
		case reflect.Invalid:
			return reflect.ValueOf(true), nil
		default:
			return reflect.ValueOf(false), nil
		}
	default:
		return reflect.Value{}, fmt.Errorf("unexpected left value %s", left.Kind().String())
	}
}

func evaluateIdentifier(scope reflect.Value, node *js.LiteralExpr) (reflect.Value, error) {
	// Handles nil
	if !scope.IsValid() {
		return reflect.Value{}, nil
	}
	switch scope.Kind() {
	case reflect.Map:
		value := scope.MapIndex(reflect.ValueOf(string(node.Data)))
		return value, nil
	default:
		return reflect.Value{}, fmt.Errorf("unexpected scope type %s", scope.Kind().String())
	}
}
