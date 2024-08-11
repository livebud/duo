package ssr

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/livebud/duo/internal/ast"
	"github.com/livebud/duo/internal/parser"
	"github.com/livebud/duo/internal/resolver"
	outscope "github.com/livebud/duo/internal/scope"
	"github.com/tdewolff/parse/v2/js"
)

func New(resolver resolver.Interface) *Renderer {
	return &Renderer{resolver}
}

type Renderer struct {
	Resolver resolver.Interface
}

func (e *Renderer) Render(w io.Writer, path string, v interface{}) error {
	file, err := e.Resolver.Resolve(&resolver.Resolve{
		Path: path,
	})
	if err != nil {
		return err
	}
	return e.Evaluate(w, file.Path, file.Code, v)
}

func (e *Renderer) Evaluate(w io.Writer, path string, code []byte, v interface{}) error {
	doc, err := parser.Parse(path, string(code))
	if err != nil {
		return err
	}
	value := reflect.ValueOf(v)
	scope, err := toScope(value)
	if err != nil {
		return err
	}
	evaluator := &evaluator{
		path:     path,
		scope:    doc.Scope,
		resolver: e.Resolver,
	}
	if err := evaluator.evaluateDocument(&ioWriter{w}, scope, doc); err != nil {
		return err
	}
	return nil
}

func toScope(value reflect.Value) (*scope, error) {
	scope := newScope()
	// Handles nil
	if !value.IsValid() {
		return scope, nil
	}
	switch value.Kind() {
	case reflect.Map:
		for _, key := range value.MapKeys() {
			scope.props[key.String()] = value.MapIndex(key)
		}
		return scope, nil
	default:
		return nil, fmt.Errorf("unexpected scope type %s", value.Kind().String())
	}
}

func newScope() *scope {
	return &scope{
		props: map[string]reflect.Value{},
		slots: map[string]strings.Builder{},
	}
}

type scope struct {
	parent *scope
	props  map[string]reflect.Value
	slot   strings.Builder
	slots  map[string]strings.Builder
}

func (s *scope) Lookup(name string) (reflect.Value, bool) {
	value, ok := s.props[name]
	if ok {
		return value, true
	}
	if s.parent != nil {
		return s.parent.Lookup(name)
	}
	return reflect.Value{}, false
}

type evaluator struct {
	path     string
	scope    *outscope.Scope
	resolver resolver.Interface
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

func (e *evaluator) errorf(format string, args ...interface{}) error {
	return fmt.Errorf("ssr: "+format, args...)
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
	case *ast.Style:
		return e.evaluateStyle(w, sc, n)
	case *ast.IfBlock:
		return e.evaluateIfBlock(w, sc, n)
	case *ast.EachBlock:
		return e.evaluateEachBlock(w, sc, n)
	case *ast.Component:
		return e.evaluateComponent(w, sc, n)
	case *ast.Slot:
		return e.evaluateSlot(w, sc, n)
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

func (e *evaluator) evaluateScript(_ writer, _ *scope, _ *ast.Script) error {
	return nil
}

func (e *evaluator) evaluateStyle(_ writer, _ *scope, _ *ast.Style) error {
	return nil
}

func (e *evaluator) evaluateAttribute(w writer, sc *scope, node ast.Attribute) error {
	switch n := node.(type) {
	case *ast.Field:
		return e.evaluateField(w, sc, n)
	case *ast.Binding:
		return e.evaluateBinding(w, sc, n)
	case *ast.AttributeShorthand:
		return e.evaluateAttributeShorthand(w, sc, n)
	default:
		return fmt.Errorf("unknown attribute %T", n)
	}
}

func (e *evaluator) evaluateField(w writer, sc *scope, node *ast.Field) error {
	// Skip event handlers
	if node.EventHandler {
		return nil
	}
	if len(node.Values) == 0 {
		w.WriteString(node.Key)
		return nil
	}
	value, err := e.evaluateValues(sc, node.Values...)
	if err != nil {
		return err
	}
	switch value.Kind() {
	case reflect.Bool:
		if value.Bool() {
			w.WriteString(node.Key)
			return nil
		} else {
			return nil
		}
	}
	valueString, err := valueToString(value)
	if err != nil {
		return e.errorf("unable to evaluating field value: %w", err)
	}
	w.WriteString(node.Key)
	w.WriteByte('=')
	w.WriteByte('"')
	w.WriteString(valueString)
	w.WriteByte('"')
	return nil
}

func (e *evaluator) evaluateBinding(w writer, sc *scope, node *ast.Binding) error {
	w.WriteString(node.Key)
	w.WriteByte('=')
	w.WriteByte('"')
	if err := e.evaluateValue(w, sc, node.Value); err != nil {
		return err
	}
	w.WriteByte('"')
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
	if err := writeValue(w, value); err != nil {
		return err
	}
	w.WriteByte('"')
	return nil
}

func (e *evaluator) evaluateValues(sc *scope, values ...ast.Value) (reflect.Value, error) {
	if len(values) == 1 {
		return evaluateValue(sc, values[0])
	}
	buf := new(bytes.Buffer)
	for _, value := range values {
		if err := e.evaluateValue(buf, sc, value); err != nil {
			return reflect.Value{}, err
		}
	}
	return reflect.ValueOf(buf.String()), nil
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

func (e *evaluator) evaluateText(w writer, _ *scope, node *ast.Text) error {
	w.WriteString(node.Value)
	return nil
}

func (e *evaluator) evaluateMustache(w writer, sc *scope, node *ast.Mustache) error {
	value, err := evaluateExpr(sc, node.Expr)
	if err != nil {
		return err
	}
	return writeValue(w, value)
}

func writeValue(w writer, value reflect.Value) error {
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

func valueToString(value reflect.Value) (string, error) {
	if !value.IsValid() {
		return "", nil
	}
	v := value.Interface()
	switch value := v.(type) {
	case string:
		return value, nil
	case int64:
		return strconv.FormatInt(value, 10), nil
	case int:
		return strconv.Itoa(value), nil
	default:
		return "", fmt.Errorf("unexpected value %T", value)
	}
}

func evaluateAttribute(sc *scope, node ast.Attribute) (reflect.Value, error) {
	switch a := node.(type) {
	case *ast.AttributeShorthand:
		value, err := evaluateExpr(sc, &js.LiteralExpr{
			Data:      []byte(a.Key),
			TokenType: js.IdentifierToken,
		})
		if err != nil {
			return reflect.Value{}, err
		}
		return value, nil
	case *ast.Field:
		value, err := evaluateValues(sc, a.Values)
		if err != nil {
			return reflect.Value{}, err
		}
		return value, nil
	default:
		return reflect.Value{}, fmt.Errorf("unknown attribute %T", a)
	}
}

func evaluateValues(scope *scope, values []ast.Value) (reflect.Value, error) {
	switch len(values) {
	case 0:
		return reflect.Value{}, nil
	case 1:
		return evaluateValue(scope, values[0])
	default:
		buf := new(bytes.Buffer)
		for _, value := range values {
			rv, err := evaluateValue(scope, value)
			if err != nil {
				return reflect.Value{}, err
			}
			if err := writeValue(buf, rv); err != nil {
				return reflect.Value{}, err
			}
		}
		return reflect.ValueOf(buf.String()), nil
	}
}

func evaluateValue(scope *scope, node ast.Value) (reflect.Value, error) {
	switch n := node.(type) {
	case *ast.Text:
		return reflect.ValueOf(n.Value), nil
	case *ast.Mustache:
		return evaluateExpr(scope, n.Expr)
	default:
		return reflect.Value{}, fmt.Errorf("unknown attribute value %T", n)
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
	case js.EqEqToken:
		return evaluateEqual(scope, left, right)
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

func evaluateAdd(_ *scope, left, right reflect.Value) (reflect.Value, error) {
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

func evaluateOr(_ *scope, left, right reflect.Value) (reflect.Value, error) {
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

var falseValue = reflect.ValueOf(false)

func evaluateEqual(_ *scope, left, right reflect.Value) (reflect.Value, error) {
	switch left.Kind() {
	case reflect.String:
		switch right.Kind() {
		case reflect.String:
			return reflect.ValueOf(left.String() == right.String()), nil
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int8, reflect.Int16:
			rightString := strconv.FormatInt(right.Int(), 10)
			return reflect.ValueOf(left.String() == rightString), nil
		default:
			return falseValue, nil
		}
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int8, reflect.Int16:
		switch right.Kind() {
		case reflect.String:
			leftString := strconv.FormatInt(left.Int(), 10)
			return reflect.ValueOf(leftString == right.String()), nil
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int8, reflect.Int16:
			return reflect.ValueOf(left.Int() == right.Int()), nil
		case reflect.Bool:
			leftBool := left.Int() > 0
			return reflect.ValueOf(leftBool == right.Bool()), nil
		default:
			return falseValue, nil
		}
	case reflect.Bool:
		switch right.Kind() {
		case reflect.Bool:
			return reflect.ValueOf(left.Bool() == right.Bool()), nil
		case reflect.String:
			leftString := strconv.FormatBool(left.Bool())
			return reflect.ValueOf(leftString == right.String()), nil
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int8, reflect.Int16:
			rightBool := right.Int() > 0
			return reflect.ValueOf(left.Bool() == rightBool), nil
		default:
			return falseValue, nil
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

func evaluateStrictEqual(_ *scope, left, right reflect.Value) (reflect.Value, error) {
	switch left.Kind() {
	case reflect.String:
		switch right.Kind() {
		case reflect.String:
			return reflect.ValueOf(left.String() == right.String()), nil
		default:
			return falseValue, nil
		}
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int8, reflect.Int16:
		switch right.Kind() {
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int8, reflect.Int16:
			return reflect.ValueOf(left.Int() == right.Int()), nil
		default:
			return falseValue, nil
		}
	case reflect.Bool:
		switch right.Kind() {
		case reflect.Bool:
			return reflect.ValueOf(left.Bool() == right.Bool()), nil
		default:
			return falseValue, nil
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

func (e *evaluator) evaluateEachBlock(w writer, sc *scope, node *ast.EachBlock) error {
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
		return fmt.Errorf("each must be a slice of values, but got %s", list.Kind())
	}
	// Loop over the elements of the slice
	// TODO: handle maps too
	for i := 0; i < slice.Len(); i++ {
		forScope := newScope()
		forScope.parent = sc
		if node.Key != nil {
			forScope.props[string(node.Key.Data)] = reflect.ValueOf(i)
		}
		if node.Value != nil {
			forScope.props[string(node.Value.Data)] = slice.Index(i)
		}
		if err := e.evaluateFragments(w, forScope, node.Body...); err != nil {
			return err
		}
	}
	return nil
}

func (e *evaluator) evaluateComponent(w writer, sc *scope, node *ast.Component) error {
	symbol, ok := e.scope.LookupByName(node.Name)
	if !ok {
		return fmt.Errorf("component %s not found", node.Name)
	} else if symbol.Import == nil {
		return fmt.Errorf("component %s not imported", node.Name)
	}
	file, err := e.resolver.Resolve(&resolver.Resolve{
		From: e.path,
		Path: symbol.Import.Path,
	})
	if err != nil {
		return err
	}
	doc, err := parser.Parse(file.Path, string(file.Code))
	if err != nil {
		return err
	}
	// Build props from attributes
	componentScope := newScope()
	for _, attr := range node.Attributes {
		value, err := evaluateAttribute(sc, attr)
		if err != nil {
			return err
		}
		// Skip undefined values
		if !value.IsValid() {
			continue
		}
		componentScope.props[attr.GetKey()] = value
	}
	// Build the slots
	for _, fragment := range node.Children {
		switch n := fragment.(type) {
		case *ast.Slot:
			return fmt.Errorf("named slots not implemented yet")
		default:
			if err := e.evaluateFragment(&componentScope.slot, sc, n); err != nil {
				return err
			}
		}
	}
	// Evaluate the component
	return e.evaluateDocument(w, componentScope, doc)
}

func (e *evaluator) evaluateSlot(w writer, sc *scope, node *ast.Slot) error {
	if node.Name != "" {
		return fmt.Errorf("named slots not implemented yet")
	}
	if sc.slot.Len() == 0 && len(node.Fallback) > 0 {
		if err := e.evaluateFragments(&sc.slot, sc, node.Fallback...); err != nil {
			return err
		}
	}
	w.WriteString(sc.slot.String())
	return nil
}
