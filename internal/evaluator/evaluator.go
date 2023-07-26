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
	scope, err := toScope(value)
	if err != nil {
		return err
	}
	evaluator := &evaluator{}
	if err := evaluator.evaluateDocument(&ioWriter{w}, scope, e.doc); err != nil {
		return err
	}
	return nil
}

func toScope(value reflect.Value) (*scope, error) {
	values := make(map[string]reflect.Value)
	// Handles nil
	if !value.IsValid() {
		return &scope{
			values: values,
		}, nil
	}
	switch value.Kind() {
	case reflect.Map:
		for _, key := range value.MapKeys() {
			values[key.String()] = value.MapIndex(key)
		}
		return &scope{
			values: values,
		}, nil
	default:
		return nil, fmt.Errorf("unexpected scope type %s", value.Kind().String())
	}

}

type scope struct {
	parent *scope
	values map[string]reflect.Value
}

func (s *scope) Lookup(name string) (reflect.Value, bool) {
	value, ok := s.values[name]
	if ok {
		return value, true
	}
	if s.parent != nil {
		return s.parent.Lookup(name)
	}
	return reflect.Value{}, false
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

func (e *evaluator) evaluateDocument(w writer, sc *scope, node *ast.Document) error {
	return e.evaluateFragments(w, sc, node.Children...)
}

func (e *evaluator) evaluateFragments(w writer, sc *scope, nodes ...ast.Fragment) error {
	for _, node := range nodes {
		if err := e.evaluateFragment(w, sc, node); err != nil {
			return err
		}
	}
	return nil
}

func (e *evaluator) evaluateFragment(w writer, sc *scope, node ast.Fragment) error {
	switch n := node.(type) {
	case *ast.Element:
		return e.evaluateElement(w, sc, n)
	case *ast.Text:
		return e.evaluateText(w, sc, n)
	case *ast.Mustache:
		return e.evaluateMustache(w, sc, n)
	case *ast.Script:
		return e.evaluateScript(w, sc, n)
	case *ast.IfBlock:
		return e.evaluateIfBlock(w, sc, n)
	case *ast.ForBlock:
		return e.evaluateForBlock(w, sc, n)
	default:
		return fmt.Errorf("unknown fragment %T", n)
	}
}

func (e *evaluator) evaluateElement(w writer, sc *scope, node *ast.Element) error {
	w.WriteByte('<')
	w.WriteString(node.Name)
	for _, attr := range node.Attributes {
		buf := new(bytes.Buffer)
		if err := e.evaluateAttribute(buf, sc, attr); err != nil {
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
		if err := e.evaluateFragment(w, sc, child); err != nil {
			return err
		}
	}
	w.WriteString("</")
	w.WriteString(node.Name)
	w.WriteString(">")
	return nil
}

func (e *evaluator) evaluateScript(w writer, sc *scope, node *ast.Script) error {
	return nil
}

func (e *evaluator) evaluateAttribute(w writer, sc *scope, node ast.Attribute) error {
	switch n := node.(type) {
	case *ast.Field:
		return e.evaluateField(w, sc, n)
	case *ast.AttributeShorthand:
		return e.evaluateAttributeShorthand(w, sc, n)
	default:
		return fmt.Errorf("unknown attribute %T", n)
	}
}

func (e *evaluator) evaluateField(w writer, sc *scope, node *ast.Field) error {
	buf := new(bytes.Buffer)
	// Skip event handlers
	if node.EventHandler {
		return nil
	}
	for _, value := range node.Values {
		if err := e.evaluateValue(buf, sc, value); err != nil {
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

func (e *evaluator) evaluateAttributeShorthand(w writer, sc *scope, node *ast.AttributeShorthand) error {
	// buf := new(bytes.Buffer)
	// Skip event handlers
	if node.EventHandler {
		return nil
	}
	value, err := evaluateExpr(sc, &js.LiteralExpr{
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

func (e *evaluator) evaluateValue(w writer, sc *scope, node ast.Value) error {
	switch n := node.(type) {
	case *ast.Text:
		return e.evaluateText(w, sc, n)
	case *ast.Mustache:
		return e.evaluateMustache(w, sc, n)
	default:
		return fmt.Errorf("unknown attribute value %T", n)
	}
}

func (e *evaluator) evaluateText(w writer, sc *scope, node *ast.Text) error {
	w.WriteString(node.Value)
	return nil
}

func (e *evaluator) evaluateMustache(w writer, sc *scope, node *ast.Mustache) error {
	value, err := evaluateExpr(sc, node.Expr)
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

func evaluateExpr(scope *scope, node js.IExpr) (reflect.Value, error) {
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

func evaluateLiteralExpr(scope *scope, node *js.LiteralExpr) (reflect.Value, error) {
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

func evaluateVar(scope *scope, node *js.Var) (reflect.Value, error) {
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

func evaluateBinaryExpr(scope *scope, node *js.BinaryExpr) (reflect.Value, error) {
	left, err := evaluateExpr(scope, node.X)
	if err != nil {
		return reflect.Value{}, err
	}
	if left.Kind() == reflect.Interface {
		left = left.Elem()
	}
	right, err := evaluateExpr(scope, node.Y)
	if err != nil {
		return reflect.Value{}, err
	}
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

func evaluateCondExpr(scope *scope, node *js.CondExpr) (reflect.Value, error) {
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

func evaluateAdd(scope *scope, left, right reflect.Value) (reflect.Value, error) {
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

func evaluateOr(scope *scope, left, right reflect.Value) (reflect.Value, error) {
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

func evaluateStrictEqual(scope *scope, left, right reflect.Value) (reflect.Value, error) {
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

func evaluateIdentifier(scope *scope, node *js.LiteralExpr) (reflect.Value, error) {
	value, ok := scope.Lookup(string(node.Data))
	if !ok {
		// return reflect.Value{}, fmt.Errorf("identifier %s not found", string(node.Data))
		return reflect.Value{}, nil
	}
	return value, nil
}

func (e *evaluator) evaluateIfBlock(w writer, sc *scope, node *ast.IfBlock) error {
	cond, err := evaluateExpr(sc, node.Cond)
	if err != nil {
		return err
	}
	if isTruthy(cond) {
		return e.evaluateFragments(w, sc, node.Then...)
	}
	return e.evaluateFragments(w, sc, node.Else...)
}

func isTruthy(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.Interface:
		return isTruthy(value.Elem())
	case reflect.Bool:
		return value.Bool()
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int8, reflect.Int16:
		return value.Int() != 0
	case reflect.String:
		return value.String() != ""
	case reflect.Invalid:
		return false
	default:
		return false
	}
}

func toSlice(value reflect.Value) (reflect.Value, bool) {
	switch value.Kind() {
	case reflect.Slice:
		return value, true
	case reflect.Interface:
		return toSlice(value.Elem())
	default:
		return value, false
	}
}

func (e *evaluator) evaluateForBlock(w writer, sc *scope, node *ast.ForBlock) error {
	list, err := evaluateExpr(sc, node.List)
	if err != nil {
		return err
	}
	// Skip over undefined lists
	if !list.IsValid() {
		return nil
	}
	// Convert the list to a slice
	slice, ok := toSlice(list)
	if !ok {
		return fmt.Errorf("for: must be a slice of values, but got %s", list.Kind())
	}
	// Loop over the elements of the slice
	// TODO: handle maps too
	for i := 0; i < slice.Len(); i++ {
		newScope := &scope{
			parent: sc,
			values: map[string]reflect.Value{},
		}
		if node.Key != nil {
			newScope.values[string(node.Key.Data)] = reflect.ValueOf(i)
		}
		if node.Value != nil {
			newScope.values[string(node.Value.Data)] = slice.Index(i)
		}
		if err := e.evaluateFragments(w, newScope, node.Body...); err != nil {
			return err
		}
	}
	return nil
}
