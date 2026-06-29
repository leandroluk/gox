// schema/object/object.go
package object

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/leandroluk/gox/validate/internal/ast"
	"github.com/leandroluk/gox/validate/internal/codec"
	"github.com/leandroluk/gox/validate/internal/defaults"
	"github.com/leandroluk/gox/validate/internal/engine"
	"github.com/leandroluk/gox/validate/internal/issues"
	"github.com/leandroluk/gox/validate/internal/ruleset"
	"github.com/leandroluk/gox/validate/schema"
	"github.com/leandroluk/gox/validate/schema/boolean"
	"github.com/leandroluk/gox/validate/schema/date"
	"github.com/leandroluk/gox/validate/schema/duration"
	"github.com/leandroluk/gox/validate/schema/text"
)

// --- rules ---

const (
	CodeRequired = "object.required"
	CodeType     = "object.type"
	CodeInvalid  = "object.invalid"

	CodeFieldDecode = "object.field.decode"

	CodeFieldRequiredIf      = "object.field.required_if"
	CodeFieldRequiredWith    = "object.field.required_with"
	CodeFieldRequiredWithout = "object.field.required_without"
	CodeFieldExcludedIf      = "object.field.excluded_if"

	CodeFieldEqField  = "object.field.eqfield"
	CodeFieldNeField  = "object.field.nefield"
	CodeFieldGtField  = "object.field.gtfield"
	CodeFieldGteField = "object.field.gtefield"
	CodeFieldLtField  = "object.field.ltfield"
	CodeFieldLteField = "object.field.ltefield"

	CodeFieldEqCSField  = "object.field.eqcsfield"
	CodeFieldNeCSField  = "object.field.necsfield"
	CodeFieldGtCSField  = "object.field.gtcsfield"
	CodeFieldGteCSField = "object.field.gtecsfield"
	CodeFieldLtCSField  = "object.field.ltcsfield"
	CodeFieldLteCSField = "object.field.ltecsfield"

	CodeFieldContains = "object.field.fieldcontains"
	CodeFieldExcludes = "object.field.fieldexcludes"
)

var ErrInvalidBuilderUsage = fmt.Errorf("invalid builder usage")

// --- field ---

type field[T any] struct {
	name      string
	offset    uintptr
	fieldType reflect.Type

	skipUnlessConditions []SkipUnlessCondition
	excludedConditions   []ExcludedIfCondition
	requiredConditions   []RequiredCondition
	required             bool

	comparators []Comparator

	validate func(context *engine.Context, value any) (any, error)
	assign   func(outputPointer unsafe.Pointer, value any)
}

func newField[T any](structPointer *T, fieldPointer any, validator func(context *engine.Context, value any) (any, error)) (field[T], error) {
	if structPointer == nil {
		return field[T]{}, fmt.Errorf("nil target")
	}
	if fieldPointer == nil {
		return field[T]{}, fmt.Errorf("nil field pointer")
	}

	structValue := reflect.ValueOf(structPointer)
	if structValue.Kind() != reflect.Pointer || structValue.Elem().Kind() != reflect.Struct {
		return field[T]{}, fmt.Errorf("target must be pointer to struct")
	}

	fieldValue := reflect.ValueOf(fieldPointer)
	if fieldValue.Kind() != reflect.Pointer {
		return field[T]{}, fmt.Errorf("fieldPointer must be a pointer")
	}

	structType := structValue.Elem().Type()

	base := unsafe.Pointer(structValue.Pointer())
	fieldAddr := unsafe.Pointer(fieldValue.Pointer())

	var matched reflect.StructField
	var matchedOffset uintptr
	found := false

	for i := 0; i < structType.NumField(); i++ {
		sf := structType.Field(i)

		if sf.Anonymous {
			continue
		}

		addr := unsafe.Pointer(uintptr(base) + sf.Offset)
		if addr == fieldAddr {
			matched = sf
			matchedOffset = sf.Offset
			found = true
			break
		}
	}

	if !found {
		return field[T]{}, fmt.Errorf("failed to resolve field (pointer does not match any direct field)")
	}

	name := jsonName(matched)
	if name == "" {
		return field[T]{}, fmt.Errorf("field has empty name (maybe json:\"-\"?)")
	}

	fieldType := matched.Type

	validateFn := func(context *engine.Context, value any) (any, error) {
		if validator != nil {
			return validator(context, value)
		}

		astValue, ok := value.(ast.Value)
		if ok {
			return decodeFallback(astValue, fieldType)
		}
		astValuePointer, ok := value.(*ast.Value)
		if ok {
			if astValuePointer == nil {
				return decodeFallback(ast.NullValue(), fieldType)
			}
			return decodeFallback(*astValuePointer, fieldType)
		}

		coerced, err := engine.InputToASTWithOptions(value, context.Options)
		if err != nil {
			return reflect.Zero(fieldType).Interface(), err
		}
		return decodeFallback(coerced, fieldType)
	}

	assignFn := buildAssignFn(fieldType, matchedOffset)

	return field[T]{
		name:      name,
		offset:    matchedOffset,
		fieldType: fieldType,

		skipUnlessConditions: make([]SkipUnlessCondition, 0),
		excludedConditions:   make([]ExcludedIfCondition, 0),
		requiredConditions:   make([]RequiredCondition, 0),
		comparators:          make([]Comparator, 0),

		validate: validateFn,
		assign:   assignFn,
	}, nil
}

type fieldInfo[T any] struct {
	name      string
	offset    uintptr
	fieldType reflect.Type
}

func resolveFieldInfo[T any](structPointer *T, fieldPointer any) (fieldInfo[T], error) {
	if structPointer == nil {
		return fieldInfo[T]{}, fmt.Errorf("nil target")
	}
	if fieldPointer == nil {
		return fieldInfo[T]{}, fmt.Errorf("nil field pointer")
	}

	structValue := reflect.ValueOf(structPointer)
	if structValue.Kind() != reflect.Pointer || structValue.Elem().Kind() != reflect.Struct {
		return fieldInfo[T]{}, fmt.Errorf("target must be pointer to struct")
	}

	fieldValue := reflect.ValueOf(fieldPointer)
	if fieldValue.Kind() != reflect.Pointer {
		return fieldInfo[T]{}, fmt.Errorf("fieldPointer must be a pointer")
	}

	structType := structValue.Elem().Type()

	base := unsafe.Pointer(structValue.Pointer())
	fieldAddr := unsafe.Pointer(fieldValue.Pointer())

	var matched reflect.StructField
	var matchedOffset uintptr
	found := false

	for i := 0; i < structType.NumField(); i++ {
		sf := structType.Field(i)

		if sf.Anonymous {
			continue
		}

		addr := unsafe.Pointer(uintptr(base) + sf.Offset)
		if addr == fieldAddr {
			matched = sf
			matchedOffset = sf.Offset
			found = true
			break
		}
	}

	if !found {
		return fieldInfo[T]{}, fmt.Errorf("failed to resolve field (pointer does not match any direct field)")
	}

	name := jsonName(matched)
	if name == "" {
		return fieldInfo[T]{}, fmt.Errorf("field has empty name (maybe json:\"-\"?)")
	}

	return fieldInfo[T]{
		name:      name,
		offset:    matchedOffset,
		fieldType: matched.Type,
	}, nil
}

func newFieldFromInfo[T any](info fieldInfo[T], validator func(context *engine.Context, value any) (any, error)) (field[T], error) {
	validateFn := func(context *engine.Context, value any) (any, error) {
		if validator != nil {
			return validator(context, value)
		}

		astValue, ok := value.(ast.Value)
		if ok {
			return decodeFallback(astValue, info.fieldType)
		}
		astValuePointer, ok := value.(*ast.Value)
		if ok {
			if astValuePointer == nil {
				return decodeFallback(ast.NullValue(), info.fieldType)
			}
			return decodeFallback(*astValuePointer, info.fieldType)
		}

		coerced, err := engine.InputToASTWithOptions(value, context.Options)
		if err != nil {
			return reflect.Zero(info.fieldType).Interface(), err
		}
		return decodeFallback(coerced, info.fieldType)
	}

	assignFn := buildAssignFn(info.fieldType, info.offset)

	return field[T]{
		name:      info.name,
		offset:    info.offset,
		fieldType: info.fieldType,

		skipUnlessConditions: make([]SkipUnlessCondition, 0),
		excludedConditions:   make([]ExcludedIfCondition, 0),
		requiredConditions:   make([]RequiredCondition, 0),
		comparators:          make([]Comparator, 0),

		validate: validateFn,
		assign:   assignFn,
	}, nil
}

func buildAssignFn(fieldType reflect.Type, offset uintptr) func(unsafe.Pointer, any) {
	return func(outputPointer unsafe.Pointer, value any) {
		if outputPointer == nil {
			return
		}

		target := reflect.NewAt(fieldType, unsafe.Pointer(uintptr(outputPointer)+offset)).Elem()

		if value == nil {
			target.Set(reflect.Zero(fieldType))
			return
		}

		v := reflect.ValueOf(value)
		if !v.IsValid() {
			target.Set(reflect.Zero(fieldType))
			return
		}

		if v.Type().AssignableTo(fieldType) {
			target.Set(v)
			return
		}

		if v.Type().ConvertibleTo(fieldType) {
			target.Set(v.Convert(fieldType))
			return
		}

		if fieldType.Kind() == reflect.Pointer {
			elemType := fieldType.Elem()
			if v.Type().AssignableTo(elemType) {
				ptr := reflect.New(elemType)
				ptr.Elem().Set(v)
				target.Set(ptr)
				return
			}
			if v.Type().ConvertibleTo(elemType) {
				ptr := reflect.New(elemType)
				ptr.Elem().Set(v.Convert(elemType))
				target.Set(ptr)
				return
			}
		}

		target.Set(reflect.Zero(fieldType))
	}
}

func jsonName(f reflect.StructField) string {
	tag := f.Tag.Get("json")
	if tag == "-" {
		return ""
	}
	if tag == "" {
		return f.Name
	}

	name := tag
	if comma := strings.IndexByte(tag, ','); comma >= 0 {
		name = tag[:comma]
	}

	if name == "" {
		return f.Name
	}

	return name
}

func decodeFallback(value ast.Value, fieldType reflect.Type) (any, error) {
	if value.IsMissing() || value.IsNull() {
		return reflect.Zero(fieldType).Interface(), nil
	}

	outPointer := reflect.New(fieldType)
	if err := codec.DecodeInto(value, outPointer.Interface()); err != nil {
		return reflect.Zero(fieldType).Interface(), err
	}

	return outPointer.Elem().Interface(), nil
}

func anyToASTValue(expected any) (ast.Value, bool) {
	switch typed := expected.(type) {
	case ast.Value:
		return typed, true

	case nil:
		return ast.NullValue(), true

	case bool:
		return ast.BooleanValue(typed), true

	case string:
		return ast.StringValue(typed), true

	case int:
		return ast.NumberValue(strconv.FormatInt(int64(typed), 10)), true
	case int8:
		return ast.NumberValue(strconv.FormatInt(int64(typed), 10)), true
	case int16:
		return ast.NumberValue(strconv.FormatInt(int64(typed), 10)), true
	case int32:
		return ast.NumberValue(strconv.FormatInt(int64(typed), 10)), true
	case int64:
		return ast.NumberValue(strconv.FormatInt(typed, 10)), true

	case uint:
		return ast.NumberValue(strconv.FormatUint(uint64(typed), 10)), true
	case uint8:
		return ast.NumberValue(strconv.FormatUint(uint64(typed), 10)), true
	case uint16:
		return ast.NumberValue(strconv.FormatUint(uint64(typed), 10)), true
	case uint32:
		return ast.NumberValue(strconv.FormatUint(uint64(typed), 10)), true
	case uint64:
		return ast.NumberValue(strconv.FormatUint(typed, 10)), true

	case float32:
		return ast.NumberValue(strconv.FormatFloat(float64(typed), 'g', -1, 32)), true
	case float64:
		return ast.NumberValue(strconv.FormatFloat(typed, 'g', -1, 64)), true

	default:
		return ast.Value{}, false
	}
}

// --- object/rule (inlined) ---

var ErrUnsupportedExpectedValue = fmt.Errorf("unsupported expected value")

type ConditionOp string

const (
	OpEq      ConditionOp = "eq"
	OpNeq     ConditionOp = "neq"
	OpPresent ConditionOp = "present"
	OpMissing ConditionOp = "missing"
	OpNull    ConditionOp = "null"
	OpNotNull ConditionOp = "notnull"

	Eq      ConditionOp = OpEq
	Ne      ConditionOp = OpNeq
	Present ConditionOp = OpPresent
	Missing ConditionOp = OpMissing
	Null    ConditionOp = OpNull
	NotNull ConditionOp = OpNotNull
)

type Comparator interface {
	Apply(context *engine.Context, root ast.Value, child ast.Value) bool
}

type RequiredCondition interface {
	Apply(context *engine.Context, root ast.Value, child ast.Value, childPresent bool) bool
}

type ExcludedIfCondition struct {
	code     string
	path     string
	op       ConditionOp
	expected ast.Value
}

type SkipUnlessCondition struct {
	path     string
	op       ConditionOp
	expected ast.Value
}

func conditionMet(actual ast.Value, op ConditionOp, expected ast.Value) bool {
	switch op {
	case OpPresent:
		return !actual.IsMissing() && !actual.IsNull()
	case OpMissing:
		return actual.IsMissing()
	case OpNull:
		return actual.IsNull()
	case OpNotNull:
		return !actual.IsNull()
	case OpEq:
		return astValEqual(actual, expected)
	case OpNeq:
		return !astValEqual(actual, expected)
	default:
		return false
	}
}

func astValEqual(actual ast.Value, expected ast.Value) bool {
	if expected.IsNull() {
		return actual.IsNull()
	}
	if actual.IsMissing() || actual.IsNull() {
		return false
	}
	if actual.Kind != expected.Kind {
		return false
	}
	switch expected.Kind {
	case ast.KindBoolean:
		return actual.Boolean == expected.Boolean
	case ast.KindString:
		return actual.String == expected.String
	case ast.KindNumber:
		return normalizeNumText(actual.Number) == normalizeNumText(expected.Number)
	default:
		return false
	}
}

func normalizeNumText(text string) string {
	if text == "" {
		return text
	}
	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return text
	}
	return strconv.FormatFloat(value, 'g', -1, 64)
}

func astValToMeta(value ast.Value) any {
	if value.IsMissing() {
		return "(missing)"
	}
	if value.IsNull() {
		return nil
	}
	switch value.Kind {
	case ast.KindBoolean:
		return value.Boolean
	case ast.KindString:
		return value.String
	case ast.KindNumber:
		return value.Number
	default:
		return value.Kind.String()
	}
}

func compareOrder(left ast.Value, right ast.Value) (int, bool) {
	if left.IsMissing() || left.IsNull() || right.IsMissing() || right.IsNull() {
		return 0, false
	}
	if left.Kind != right.Kind {
		return 0, false
	}
	switch left.Kind {
	case ast.KindNumber:
		lf, err1 := strconv.ParseFloat(left.Number, 64)
		rf, err2 := strconv.ParseFloat(right.Number, 64)
		if err1 != nil || err2 != nil {
			return 0, false
		}
		if lf < rf {
			return -1, true
		}
		if lf > rf {
			return 1, true
		}
		return 0, true
	case ast.KindString:
		if left.String < right.String {
			return -1, true
		}
		if left.String > right.String {
			return 1, true
		}
		return 0, true
	default:
		return 0, false
	}
}

func condContains(text string, needle string) bool {
	if needle == "" {
		return true
	}
	for i := 0; i+len(needle) <= len(text); i++ {
		if text[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

// ExcludedIf returns a condition that excludes the field when the path matches op/expected.
func ExcludedIf(code string, path string, op ConditionOp, expected ast.Value) ExcludedIfCondition {
	return ExcludedIfCondition{code: code, path: path, op: op, expected: expected}
}

func (c ExcludedIfCondition) Apply(context *engine.Context, root ast.Value, _ ast.Value, childPresent bool) (skip bool, stop bool) {
	actual := ast.Query(root, c.path)
	if !conditionMet(actual, c.op, c.expected) {
		return false, false
	}
	if childPresent {
		stop = context.AddIssue(c.code, "excluded", map[string]any{
			"path": c.path, "op": string(c.op), "expected": astValToMeta(c.expected),
		})
		return true, stop
	}
	return true, false
}

func SkipUnless(path string, op ConditionOp, expected ast.Value) SkipUnlessCondition {
	return SkipUnlessCondition{path: path, op: op, expected: expected}
}

func (c SkipUnlessCondition) ShouldSkip(_ *engine.Context, root ast.Value) bool {
	actual := ast.Query(root, c.path)
	return !conditionMet(actual, c.op, c.expected)
}

type requiredIfCondition struct {
	code     string
	path     string
	op       ConditionOp
	expected ast.Value
}

func RequiredIf(code string, path string, op ConditionOp, expected ast.Value) RequiredCondition {
	return requiredIfCondition{code: code, path: path, op: op, expected: expected}
}

func (c requiredIfCondition) Apply(context *engine.Context, root ast.Value, _ ast.Value, childPresent bool) bool {
	if childPresent {
		return false
	}
	actual := ast.Query(root, c.path)
	if !conditionMet(actual, c.op, c.expected) {
		return false
	}
	return context.AddIssue(c.code, "required", map[string]any{
		"path": c.path, "op": string(c.op), "expected": astValToMeta(c.expected),
	})
}

type requiredWithCondition struct {
	code  string
	paths []string
}

func RequiredWith(code string, paths ...string) RequiredCondition {
	return requiredWithCondition{code: code, paths: append([]string(nil), paths...)}
}

func (c requiredWithCondition) Apply(context *engine.Context, root ast.Value, _ ast.Value, childPresent bool) bool {
	if childPresent {
		return false
	}
	for _, path := range c.paths {
		actual := ast.Query(root, path)
		if !actual.IsMissing() && !actual.IsNull() {
			return context.AddIssue(c.code, "required", map[string]any{"paths": append([]string(nil), c.paths...)})
		}
	}
	return false
}

type requiredWithoutCondition struct {
	code  string
	paths []string
}

func RequiredWithout(code string, paths ...string) RequiredCondition {
	return requiredWithoutCondition{code: code, paths: append([]string(nil), paths...)}
}

func (c requiredWithoutCondition) Apply(context *engine.Context, root ast.Value, _ ast.Value, childPresent bool) bool {
	if childPresent {
		return false
	}
	for _, path := range c.paths {
		actual := ast.Query(root, path)
		if actual.IsMissing() || actual.IsNull() {
			return context.AddIssue(c.code, "required", map[string]any{"paths": append([]string(nil), c.paths...)})
		}
	}
	return false
}

// comparators

type eqFieldComparator struct{ code, other string }

func EqField(code string, other string) Comparator { return eqFieldComparator{code, other} }
func (c eqFieldComparator) Apply(context *engine.Context, root ast.Value, child ast.Value) bool {
	other := ast.Query(root, c.other)
	if astValEqual(child, other) {
		return false
	}
	return context.AddIssue(c.code, "must be equal", map[string]any{"other": c.other, "actual": astValToMeta(child), "value": astValToMeta(other)})
}

type neFieldComparator struct{ code, other string }

func NeField(code string, other string) Comparator { return neFieldComparator{code, other} }
func (c neFieldComparator) Apply(context *engine.Context, root ast.Value, child ast.Value) bool {
	other := ast.Query(root, c.other)
	if !astValEqual(child, other) {
		return false
	}
	return context.AddIssue(c.code, "must not be equal", map[string]any{"other": c.other, "actual": astValToMeta(child), "value": astValToMeta(other)})
}

type gtFieldComparator struct{ code, other string }

func GtField(code string, other string) Comparator { return gtFieldComparator{code, other} }
func (c gtFieldComparator) Apply(context *engine.Context, root ast.Value, child ast.Value) bool {
	other := ast.Query(root, c.other)
	if order, ok := compareOrder(child, other); ok && order > 0 {
		return false
	}
	return context.AddIssue(c.code, "must be greater", map[string]any{"other": c.other, "actual": astValToMeta(child), "value": astValToMeta(other)})
}

type gteFieldComparator struct{ code, other string }

func GteField(code string, other string) Comparator { return gteFieldComparator{code, other} }
func (c gteFieldComparator) Apply(context *engine.Context, root ast.Value, child ast.Value) bool {
	other := ast.Query(root, c.other)
	if order, ok := compareOrder(child, other); ok && order >= 0 {
		return false
	}
	return context.AddIssue(c.code, "must be greater or equal", map[string]any{"other": c.other, "actual": astValToMeta(child), "value": astValToMeta(other)})
}

type ltFieldComparator struct{ code, other string }

func LtField(code string, other string) Comparator { return ltFieldComparator{code, other} }
func (c ltFieldComparator) Apply(context *engine.Context, root ast.Value, child ast.Value) bool {
	other := ast.Query(root, c.other)
	if order, ok := compareOrder(child, other); ok && order < 0 {
		return false
	}
	return context.AddIssue(c.code, "must be lower", map[string]any{"other": c.other, "actual": astValToMeta(child), "value": astValToMeta(other)})
}

type lteFieldComparator struct{ code, other string }

func LteField(code string, other string) Comparator { return lteFieldComparator{code, other} }
func (c lteFieldComparator) Apply(context *engine.Context, root ast.Value, child ast.Value) bool {
	other := ast.Query(root, c.other)
	if order, ok := compareOrder(child, other); ok && order <= 0 {
		return false
	}
	return context.AddIssue(c.code, "must be lower or equal", map[string]any{"other": c.other, "actual": astValToMeta(child), "value": astValToMeta(other)})
}

type eqCSFieldComparator struct{ code, path string }

func EqCSField(code string, path string) Comparator { return eqCSFieldComparator{code, path} }
func (c eqCSFieldComparator) Apply(context *engine.Context, root ast.Value, child ast.Value) bool {
	other := ast.Query(root, c.path)
	if astValEqual(child, other) {
		return false
	}
	return context.AddIssue(c.code, "must be equal", map[string]any{"path": c.path, "actual": astValToMeta(child), "value": astValToMeta(other)})
}

type neCSFieldComparator struct{ code, path string }

func NeCSField(code string, path string) Comparator { return neCSFieldComparator{code, path} }
func (c neCSFieldComparator) Apply(context *engine.Context, root ast.Value, child ast.Value) bool {
	other := ast.Query(root, c.path)
	if !astValEqual(child, other) {
		return false
	}
	return context.AddIssue(c.code, "must not be equal", map[string]any{"path": c.path, "actual": astValToMeta(child), "value": astValToMeta(other)})
}

type gtCSFieldComparator struct{ code, path string }

func GtCSField(code string, path string) Comparator { return gtCSFieldComparator{code, path} }
func (c gtCSFieldComparator) Apply(context *engine.Context, root ast.Value, child ast.Value) bool {
	other := ast.Query(root, c.path)
	if order, ok := compareOrder(child, other); ok && order > 0 {
		return false
	}
	return context.AddIssue(c.code, "must be greater", map[string]any{"path": c.path, "actual": astValToMeta(child), "value": astValToMeta(other)})
}

type gteCSFieldComparator struct{ code, path string }

func GteCSField(code string, path string) Comparator { return gteCSFieldComparator{code, path} }
func (c gteCSFieldComparator) Apply(context *engine.Context, root ast.Value, child ast.Value) bool {
	other := ast.Query(root, c.path)
	if order, ok := compareOrder(child, other); ok && order >= 0 {
		return false
	}
	return context.AddIssue(c.code, "must be greater or equal", map[string]any{"path": c.path, "actual": astValToMeta(child), "value": astValToMeta(other)})
}

type ltCSFieldComparator struct{ code, path string }

func LtCSField(code string, path string) Comparator { return ltCSFieldComparator{code, path} }
func (c ltCSFieldComparator) Apply(context *engine.Context, root ast.Value, child ast.Value) bool {
	other := ast.Query(root, c.path)
	if order, ok := compareOrder(child, other); ok && order < 0 {
		return false
	}
	return context.AddIssue(c.code, "must be lower", map[string]any{"path": c.path, "actual": astValToMeta(child), "value": astValToMeta(other)})
}

type lteCSFieldComparator struct{ code, path string }

func LteCSField(code string, path string) Comparator { return lteCSFieldComparator{code, path} }
func (c lteCSFieldComparator) Apply(context *engine.Context, root ast.Value, child ast.Value) bool {
	other := ast.Query(root, c.path)
	if order, ok := compareOrder(child, other); ok && order <= 0 {
		return false
	}
	return context.AddIssue(c.code, "must be lower or equal", map[string]any{"path": c.path, "actual": astValToMeta(child), "value": astValToMeta(other)})
}

type fieldContainsComparator struct{ code, other string }

func FieldContains(code string, other string) Comparator { return fieldContainsComparator{code, other} }
func (c fieldContainsComparator) Apply(context *engine.Context, root ast.Value, child ast.Value) bool {
	other := ast.Query(root, c.other)
	if child.Kind == ast.KindString && other.Kind == ast.KindString {
		if other.String == "" || condContains(child.String, other.String) {
			return false
		}
	}
	return context.AddIssue(c.code, "must contain", map[string]any{"other": c.other, "actual": astValToMeta(child), "value": astValToMeta(other)})
}

type fieldExcludesComparator struct{ code, other string }

func FieldExcludes(code string, other string) Comparator {
	return fieldExcludesComparator{code, other}
}
func (c fieldExcludesComparator) Apply(context *engine.Context, root ast.Value, child ast.Value) bool {
	other := ast.Query(root, c.other)
	if child.Kind == ast.KindString && other.Kind == ast.KindString {
		if other.String == "" || !condContains(child.String, other.String) {
			return false
		}
	}
	return context.AddIssue(c.code, "must not contain", map[string]any{"other": c.other, "actual": astValToMeta(child), "value": astValToMeta(other)})
}

// --- schema ---

type Schema[T any] struct {
	required bool

	structOnly    bool
	noStructLevel bool

	defaultProvider defaults.Provider[T]

	fields []field[T]
	rules  []ruleset.RuleFn[T]

	lastFieldIndex int

	buildTarget *T
	buildError  error
}

func New[T any](builder func(target *T, schemaValue *Schema[T])) *Schema[T] {
	schemaValue := &Schema[T]{
		defaultProvider: defaults.None[T](),
		fields:          make([]field[T], 0),
		rules:           make([]ruleset.RuleFn[T], 0),
		lastFieldIndex:  -1,
	}

	var target T
	schemaValue.buildTarget = &target

	if builder != nil {
		builder(&target, schemaValue)
	}

	schemaValue.buildTarget = nil
	return schemaValue
}

func (s *Schema[T]) Required() *Schema[T] {
	s.required = true
	return s
}

func (s *Schema[T]) StructOnly() *Schema[T] {
	s.structOnly = true
	s.noStructLevel = false
	return s
}

func (s *Schema[T]) NoStructLevel() *Schema[T] {
	s.noStructLevel = true
	s.structOnly = false
	return s
}

func (s *Schema[T]) Default(value T) *Schema[T] {
	s.defaultProvider = defaults.Value(value)
	return s
}

func (s *Schema[T]) DefaultFunc(fn func() T) *Schema[T] {
	s.defaultProvider = defaults.Func(fn)
	return s
}

func (s *Schema[T]) Custom(ruleValue ruleset.RuleFn[T]) *Schema[T] {
	if ruleValue != nil {
		s.rules = append(s.rules, ruleValue)
	}
	return s
}

func (s *Schema[T]) Field(fieldPointer any) *FieldBuilder[T] {
	return newFieldBuilder(s, fieldPointer)
}

func (s *Schema[T]) RequiredIf(path string, op ConditionOp, expected any) *Schema[T] {
	fieldPointer := s.lastField()
	if fieldPointer == nil {
		s.buildError = ErrInvalidBuilderUsage
		return s
	}

	expectedValue, ok := anyToASTValue(expected)
	if !ok {
		s.buildError = ErrUnsupportedExpectedValue
		return s
	}

	fieldPointer.requiredConditions = append(fieldPointer.requiredConditions, RequiredIf(CodeFieldRequiredIf, path, op, expectedValue))
	return s
}

func (s *Schema[T]) RequiredWith(paths ...string) *Schema[T] {
	fieldPointer := s.lastField()
	if fieldPointer == nil {
		s.buildError = ErrInvalidBuilderUsage
		return s
	}

	fieldPointer.requiredConditions = append(fieldPointer.requiredConditions, RequiredWith(CodeFieldRequiredWith, paths...))
	return s
}

func (s *Schema[T]) RequiredWithout(paths ...string) *Schema[T] {
	fieldPointer := s.lastField()
	if fieldPointer == nil {
		s.buildError = ErrInvalidBuilderUsage
		return s
	}

	fieldPointer.requiredConditions = append(fieldPointer.requiredConditions, RequiredWithout(CodeFieldRequiredWithout, paths...))
	return s
}

func (s *Schema[T]) ExcludedIf(path string, op ConditionOp, expected any) *Schema[T] {
	fieldPointer := s.lastField()
	if fieldPointer == nil {
		s.buildError = ErrInvalidBuilderUsage
		return s
	}

	expectedValue, ok := anyToASTValue(expected)
	if !ok {
		s.buildError = ErrUnsupportedExpectedValue
		return s
	}

	fieldPointer.excludedConditions = append(fieldPointer.excludedConditions, ExcludedIf(CodeFieldExcludedIf, path, op, expectedValue))
	return s
}

func (s *Schema[T]) SkipUnless(path string, op ConditionOp, expected any) *Schema[T] {
	fieldPointer := s.lastField()
	if fieldPointer == nil {
		s.buildError = ErrInvalidBuilderUsage
		return s
	}

	expectedValue, ok := anyToASTValue(expected)
	if !ok {
		s.buildError = ErrUnsupportedExpectedValue
		return s
	}

	fieldPointer.skipUnlessConditions = append(fieldPointer.skipUnlessConditions, SkipUnless(path, op, expectedValue))
	return s
}

func (s *Schema[T]) EqField(other string) *Schema[T] {
	return s.addComparator(EqField(CodeFieldEqField, other))
}
func (s *Schema[T]) NeField(other string) *Schema[T] {
	return s.addComparator(NeField(CodeFieldNeField, other))
}
func (s *Schema[T]) GtField(other string) *Schema[T] {
	return s.addComparator(GtField(CodeFieldGtField, other))
}
func (s *Schema[T]) GteField(other string) *Schema[T] {
	return s.addComparator(GteField(CodeFieldGteField, other))
}
func (s *Schema[T]) LtField(other string) *Schema[T] {
	return s.addComparator(LtField(CodeFieldLtField, other))
}
func (s *Schema[T]) LteField(other string) *Schema[T] {
	return s.addComparator(LteField(CodeFieldLteField, other))
}

func (s *Schema[T]) EqCSField(path string) *Schema[T] {
	return s.addComparator(EqCSField(CodeFieldEqCSField, path))
}
func (s *Schema[T]) NeCSField(path string) *Schema[T] {
	return s.addComparator(NeCSField(CodeFieldNeCSField, path))
}
func (s *Schema[T]) GtCSField(path string) *Schema[T] {
	return s.addComparator(GtCSField(CodeFieldGtCSField, path))
}
func (s *Schema[T]) GteCSField(path string) *Schema[T] {
	return s.addComparator(GteCSField(CodeFieldGteCSField, path))
}
func (s *Schema[T]) LtCSField(path string) *Schema[T] {
	return s.addComparator(LtCSField(CodeFieldLtCSField, path))
}
func (s *Schema[T]) LteCSField(path string) *Schema[T] {
	return s.addComparator(LteCSField(CodeFieldLteCSField, path))
}

func (s *Schema[T]) FieldContains(other string) *Schema[T] {
	return s.addComparator(FieldContains(CodeFieldContains, other))
}
func (s *Schema[T]) FieldExcludes(other string) *Schema[T] {
	return s.addComparator(FieldExcludes(CodeFieldExcludes, other))
}

func (s *Schema[T]) addComparator(comparator Comparator) *Schema[T] {
	fieldPointer := s.lastField()
	if fieldPointer == nil {
		s.buildError = ErrInvalidBuilderUsage
		return s
	}
	fieldPointer.comparators = append(fieldPointer.comparators, comparator)
	return s
}

func (s *Schema[T]) lastField() *field[T] {
	if s == nil {
		return nil
	}
	if s.lastFieldIndex < 0 || s.lastFieldIndex >= len(s.fields) {
		return nil
	}
	return &s.fields[s.lastFieldIndex]
}

func (s *Schema[T]) Validate(input any, optionList ...schema.Option) (T, error) {
	options := schema.ApplyOptions(optionList...)
	return s.validateWithOptions(input, options)
}

func (s *Schema[T]) MustValidate(input any, optionList ...schema.Option) T {
	result, err := s.Validate(input, optionList...)
	if err != nil {
		panic(err)
	}
	return result
}

func (s *Schema[T]) ValidateAny(input any, options schema.Options) (any, error) {
	return s.validateWithOptions(input, options)
}

func (s *Schema[T]) OutputType() reflect.Type {
	var pointer *T
	return reflect.TypeOf(pointer).Elem()
}

// --- validate ---

func (s *Schema[T]) validateWithOptions(input any, options schema.Options) (T, error) {
	var zero T

	if s == nil {
		return zero, fmt.Errorf("schema is nil")
	}
	if s.buildError != nil {
		return zero, s.buildError
	}

	context := engine.NewContext(options)

	value, err := engine.InputToASTWithOptions(input, options)
	if err != nil {
		return zero, err
	}

	output, _ := s.validateAST(context, value)
	if ptr, ok := input.(*T); ok && ptr != nil {
		*ptr = output
	}
	return output, context.Error()
}

func (s *Schema[T]) validateAST(context *engine.Context, value ast.Value) (T, bool) {
	if defaultValue, ok := defaults.Apply(value.Presence, context.Options, s.defaultProvider); ok {
		return defaultValue, false
	}

	if value.IsMissing() || value.IsNull() {
		if s.required {
			stop := context.AddIssue(CodeRequired, "required")
			var zero T
			return zero, stop
		}
		var zero T
		return zero, false
	}

	if value.Kind != ast.KindObject {
		stop := context.AddIssue(CodeType, "expected object", map[string]any{
			"expected": "object",
			"actual":   value.Kind.String(),
		})
		var zero T
		return zero, stop
	}

	if s.structOnly {
		var output T
		if err := codec.DecodeInto(value, &output); err != nil {
			stop := context.AddIssue(CodeInvalid, "invalid object", map[string]any{
				"error": err.Error(),
			})
			var zero T
			return zero, stop
		}

		if !s.noStructLevel && len(s.rules) > 0 {
			for _, ruleValue := range s.rules {
				if ruleValue(output, context) {
					return output, true
				}
			}
		}

		return output, false
	}

	var output T
	outputPointer := unsafe.Pointer(&output)

	for _, compiledField := range s.fields {
		context.PushField(compiledField.name)

		child, ok := value.Object[compiledField.name]
		if !ok {
			child = ast.MissingValue()
		}

		action, stop := s.applyFieldPlan(context, value, compiledField, child)
		if action == fieldActionSkip {
			compiledField.assign(outputPointer, reflect.Zero(compiledField.fieldType).Interface())
			context.Pop()
			if stop {
				return output, true
			}
			continue
		}
		if stop {
			compiledField.assign(outputPointer, reflect.Zero(compiledField.fieldType).Interface())
			context.Pop()
			return output, true
		}

		fieldValue, fieldError := compiledField.validate(context, child)
		if fieldError != nil {
			stop := context.AddIssue(CodeFieldDecode, "invalid field", map[string]any{
				"error": fieldError.Error(),
			})

			compiledField.assign(outputPointer, reflect.Zero(compiledField.fieldType).Interface())
			context.Pop()

			if stop {
				return output, true
			}
			continue
		}

		compiledField.assign(outputPointer, fieldValue)
		context.Pop()
	}

	if !s.noStructLevel && len(s.rules) > 0 {
		if ruleset.Apply(output, context, s.rules...) {
			return output, true
		}
	}

	return output, false
}

type fieldAction uint8

const (
	fieldActionValidate fieldAction = iota
	fieldActionSkip
)

func (s *Schema[T]) applyFieldPlan(context *engine.Context, root ast.Value, fieldValue field[T], child ast.Value) (fieldAction, bool) {
	if len(fieldValue.skipUnlessConditions) == 0 &&
		len(fieldValue.excludedConditions) == 0 &&
		len(fieldValue.requiredConditions) == 0 &&
		len(fieldValue.comparators) == 0 &&
		!fieldValue.required {
		return fieldActionValidate, false
	}

	for _, cond := range fieldValue.skipUnlessConditions {
		if cond.ShouldSkip(context, root) {
			return fieldActionSkip, false
		}
	}

	childPresent := !child.IsMissing() && !child.IsNull()

	for _, cond := range fieldValue.excludedConditions {
		skip, stop := cond.Apply(context, root, child, childPresent)
		if stop {
			return fieldActionSkip, true
		}
		if skip {
			return fieldActionSkip, false
		}
	}

	for _, cond := range fieldValue.requiredConditions {
		if cond.Apply(context, root, child, childPresent) {
			return fieldActionSkip, true
		}
	}

	if !childPresent && !fieldValue.required {
		return fieldActionSkip, false
	}

	for _, comparator := range fieldValue.comparators {
		if comparator.Apply(context, root, child) {
			return fieldActionValidate, true
		}
	}

	return fieldActionValidate, false
}

// --- field_builder ---

type FieldBuilder[T any] struct {
	schema    *Schema[T]
	fieldInfo fieldInfo[T]
}

func newFieldBuilder[T any](s *Schema[T], fieldPointer any) *FieldBuilder[T] {
	if s == nil || s.buildTarget == nil {
		s.buildError = ErrInvalidBuilderUsage
		return &FieldBuilder[T]{schema: s}
	}

	info, err := resolveFieldInfo(s.buildTarget, fieldPointer)
	if err != nil {
		s.buildError = err
		return &FieldBuilder[T]{schema: s}
	}

	return &FieldBuilder[T]{
		schema:    s,
		fieldInfo: info,
	}
}

func (fb *FieldBuilder[T]) Text() *TextFieldBuilder[T] {
	if fb.schema.buildError != nil {
		return &TextFieldBuilder[T]{schema: fb.schema, fieldIndex: -1}
	}

	b := &TextFieldBuilder[T]{
		schema:     fb.schema,
		fieldInfo:  fb.fieldInfo,
		textSchema: text.New(),
		fieldIndex: -1,
	}
	b.build()
	return b
}

func (fb *FieldBuilder[T]) Number() *NumberFieldBuilder[T] {
	if fb.schema.buildError != nil {
		return &NumberFieldBuilder[T]{schema: fb.schema, fieldIndex: -1}
	}

	b := &NumberFieldBuilder[T]{
		schema:     fb.schema,
		fieldInfo:  fb.fieldInfo,
		fieldIndex: -1,
	}
	b.build()
	return b
}

func (fb *FieldBuilder[T]) Boolean() *BooleanFieldBuilder[T] {
	if fb.schema.buildError != nil {
		return &BooleanFieldBuilder[T]{schema: fb.schema, fieldIndex: -1}
	}

	b := &BooleanFieldBuilder[T]{
		schema:        fb.schema,
		fieldInfo:     fb.fieldInfo,
		booleanSchema: boolean.New(),
		fieldIndex:    -1,
	}
	b.build()
	return b
}

func (fb *FieldBuilder[T]) Date() *DateFieldBuilder[T] {
	if fb.schema.buildError != nil {
		return &DateFieldBuilder[T]{schema: fb.schema, fieldIndex: -1}
	}

	b := &DateFieldBuilder[T]{
		schema:     fb.schema,
		fieldInfo:  fb.fieldInfo,
		dateSchema: date.New(),
		fieldIndex: -1,
	}
	b.build()
	return b
}

func (fb *FieldBuilder[T]) Duration() *DurationFieldBuilder[T] {
	if fb.schema.buildError != nil {
		return &DurationFieldBuilder[T]{schema: fb.schema, fieldIndex: -1}
	}

	b := &DurationFieldBuilder[T]{
		schema:         fb.schema,
		fieldInfo:      fb.fieldInfo,
		durationSchema: duration.New(),
		fieldIndex:     -1,
	}
	b.build()
	return b
}

func (fb *FieldBuilder[T]) Array(items ...schema.AnySchema) *ArrayFieldBuilder[T] {
	if fb.schema.buildError != nil {
		return &ArrayFieldBuilder[T]{schema: fb.schema, fieldIndex: -1}
	}

	var itemSchema schema.AnySchema
	if len(items) > 0 {
		itemSchema = items[0]
	}

	b := &ArrayFieldBuilder[T]{
		schema:     fb.schema,
		fieldInfo:  fb.fieldInfo,
		itemSchema: itemSchema,
		fieldIndex: -1,
	}
	b.build()
	return b
}

func (fb *FieldBuilder[T]) Object(builderFunc any) *ObjectFieldBuilder[T] {
	if fb.schema.buildError != nil {
		return &ObjectFieldBuilder[T]{schema: fb.schema, fieldIndex: -1}
	}

	b := &ObjectFieldBuilder[T]{
		schema:      fb.schema,
		fieldInfo:   fb.fieldInfo,
		builderFunc: builderFunc,
		fieldIndex:  -1,
	}
	b.build()
	return b
}

func (fb *FieldBuilder[T]) Record(items ...schema.AnySchema) *RecordFieldBuilder[T] {
	if fb.schema.buildError != nil {
		return &RecordFieldBuilder[T]{schema: fb.schema, fieldIndex: -1}
	}

	var valueSchema schema.AnySchema
	if len(items) > 0 {
		valueSchema = items[0]
	}

	b := &RecordFieldBuilder[T]{
		schema:      fb.schema,
		fieldInfo:   fb.fieldInfo,
		valueSchema: valueSchema,
		fieldIndex:  -1,
	}
	b.build()
	return b
}

func (fb *FieldBuilder[T]) Transform(fn func(value any) (any, error)) *Schema[T] {
	return fb.Custom(func(ctx *engine.Context, value any) (any, error) {
		return fn(value)
	})
}

func (fb *FieldBuilder[T]) Custom(validator func(ctx *engine.Context, value any) (any, error)) *Schema[T] {
	if fb.schema.buildError != nil {
		return fb.schema
	}

	compiled, err := newFieldFromInfo(fb.fieldInfo, validator)
	if err != nil {
		fb.schema.buildError = err
		return fb.schema
	}

	fb.schema.fields = append(fb.schema.fields, compiled)
	fb.schema.lastFieldIndex = len(fb.schema.fields) - 1
	return fb.schema
}

// --- text_field_builder ---

type TextFieldBuilder[T any] struct {
	schema     *Schema[T]
	fieldInfo  fieldInfo[T]
	textSchema *text.Schema
	fieldIndex int
	required   bool
}

func (b *TextFieldBuilder[T]) build() {
	if b.schema.buildError != nil {
		return
	}

	validator := func(ctx *engine.Context, value any) (any, error) {
		if b.required {
			astVal, ok := value.(ast.Value)
			if ok && (astVal.IsMissing() || astVal.IsNull()) {
				ctx.AddIssue("text.required", "required")
				return nil, nil
			}
		}
		out, err := b.textSchema.ValidateAny(value, ctx.Options)
		if err != nil {
			if vErr, ok := err.(*issues.ValidationError); ok {
				basePath := ctx.PathString()
				for _, issue := range vErr.Issues {
					issue.Path = joinIssuePath(basePath, issue.Path)
					ctx.Issues.Add(issue)
				}
				return nil, nil
			}
			return nil, err
		}
		return out, nil
	}

	compiled, err := newFieldFromInfo(b.fieldInfo, validator)
	if err != nil {
		b.schema.buildError = err
		return
	}
	compiled.required = b.required

	if b.fieldIndex == -1 {
		b.schema.fields = append(b.schema.fields, compiled)
		b.schema.lastFieldIndex = len(b.schema.fields) - 1
		b.fieldIndex = b.schema.lastFieldIndex
	} else {
		b.schema.fields[b.fieldIndex] = compiled
		b.schema.lastFieldIndex = b.fieldIndex
	}
}

func (b *TextFieldBuilder[T]) Required() *TextFieldBuilder[T] {
	b.textSchema.Required()
	b.required = true
	b.build()
	return b
}

func (b *TextFieldBuilder[T]) IsDefault() *TextFieldBuilder[T] {
	b.textSchema.IsDefault()
	b.build()
	return b
}

func (b *TextFieldBuilder[T]) Default(value string) *TextFieldBuilder[T] {
	b.textSchema.Default(value)
	b.build()
	return b
}

func (b *TextFieldBuilder[T]) DefaultFunc(fn func() string) *TextFieldBuilder[T] {
	b.textSchema.DefaultFunc(fn)
	b.build()
	return b
}

func (b *TextFieldBuilder[T]) Len(n int) *TextFieldBuilder[T] {
	b.textSchema.Len(n)
	b.build()
	return b
}

func (b *TextFieldBuilder[T]) Min(n int) *TextFieldBuilder[T] {
	b.textSchema.Min(n)
	b.build()
	return b
}

func (b *TextFieldBuilder[T]) Max(n int) *TextFieldBuilder[T] {
	b.textSchema.Max(n)
	b.build()
	return b
}

func (b *TextFieldBuilder[T]) Email() *TextFieldBuilder[T] {
	b.textSchema.Email()
	b.build()
	return b
}

func (b *TextFieldBuilder[T]) URL() *TextFieldBuilder[T] {
	b.textSchema.URL()
	b.build()
	return b
}

func (b *TextFieldBuilder[T]) URI() *TextFieldBuilder[T] {
	b.textSchema.URI()
	b.build()
	return b
}

func (b *TextFieldBuilder[T]) UUID() *TextFieldBuilder[T] {
	b.textSchema.UUID()
	b.build()
	return b
}

func (b *TextFieldBuilder[T]) Pattern(values ...string) *TextFieldBuilder[T] {
	b.textSchema.Pattern(values...)
	b.build()
	return b
}

func (b *TextFieldBuilder[T]) OneOf(values ...string) *TextFieldBuilder[T] {
	b.textSchema.OneOf(values...)
	b.build()
	return b
}

func (b *TextFieldBuilder[T]) Enum(values ...any) *TextFieldBuilder[T] {
	b.textSchema.Enum(values...)
	b.build()
	return b
}

func (b *TextFieldBuilder[T]) RequiredIf(path string, op ConditionOp, expected any) *TextFieldBuilder[T] {
	b.build()
	b.schema.RequiredIf(path, op, expected)
	return b
}

func (b *TextFieldBuilder[T]) RequiredWith(paths ...string) *TextFieldBuilder[T] {
	b.build()
	b.schema.RequiredWith(paths...)
	return b
}

func (b *TextFieldBuilder[T]) RequiredWithout(paths ...string) *TextFieldBuilder[T] {
	b.build()
	b.schema.RequiredWithout(paths...)
	return b
}

func (b *TextFieldBuilder[T]) ExcludedIf(path string, op ConditionOp, expected any) *TextFieldBuilder[T] {
	b.build()
	b.schema.ExcludedIf(path, op, expected)
	return b
}

func (b *TextFieldBuilder[T]) SkipUnless(path string, op ConditionOp, expected any) *TextFieldBuilder[T] {
	b.build()
	b.schema.SkipUnless(path, op, expected)
	return b
}

func (b *TextFieldBuilder[T]) EqField(other string) *TextFieldBuilder[T] {
	b.build()
	b.schema.EqField(other)
	return b
}

func (b *TextFieldBuilder[T]) NeField(other string) *TextFieldBuilder[T] {
	b.build()
	b.schema.NeField(other)
	return b
}

func (b *TextFieldBuilder[T]) GtField(other string) *TextFieldBuilder[T] {
	b.build()
	b.schema.GtField(other)
	return b
}

func (b *TextFieldBuilder[T]) GteField(other string) *TextFieldBuilder[T] {
	b.build()
	b.schema.GteField(other)
	return b
}

func (b *TextFieldBuilder[T]) LtField(other string) *TextFieldBuilder[T] {
	b.build()
	b.schema.LtField(other)
	return b
}

func (b *TextFieldBuilder[T]) LteField(other string) *TextFieldBuilder[T] {
	b.build()
	b.schema.LteField(other)
	return b
}

func (b *TextFieldBuilder[T]) EqCSField(path string) *TextFieldBuilder[T] {
	b.build()
	b.schema.EqCSField(path)
	return b
}

func (b *TextFieldBuilder[T]) NeCSField(path string) *TextFieldBuilder[T] {
	b.build()
	b.schema.NeCSField(path)
	return b
}

func (b *TextFieldBuilder[T]) GtCSField(path string) *TextFieldBuilder[T] {
	b.build()
	b.schema.GtCSField(path)
	return b
}

func (b *TextFieldBuilder[T]) GteCSField(path string) *TextFieldBuilder[T] {
	b.build()
	b.schema.GteCSField(path)
	return b
}

func (b *TextFieldBuilder[T]) LtCSField(path string) *TextFieldBuilder[T] {
	b.build()
	b.schema.LtCSField(path)
	return b
}

func (b *TextFieldBuilder[T]) LteCSField(path string) *TextFieldBuilder[T] {
	b.build()
	b.schema.LteCSField(path)
	return b
}

func (b *TextFieldBuilder[T]) FieldContains(other string) *TextFieldBuilder[T] {
	b.build()
	b.schema.FieldContains(other)
	return b
}

func (b *TextFieldBuilder[T]) FieldExcludes(other string) *TextFieldBuilder[T] {
	b.build()
	b.schema.FieldExcludes(other)
	return b
}

func (b *TextFieldBuilder[T]) ValidateAny(value any, options schema.Options) (any, error) {
	return b.textSchema.ValidateAny(value, options)
}

func (b *TextFieldBuilder[T]) Build() *Schema[T] {
	b.build()
	return b.schema
}

func (b *TextFieldBuilder[T]) Transform(fn func(value any) (any, error)) *Schema[T] {
	validator := func(ctx *engine.Context, value any) (any, error) {
		if b.required {
			astVal, ok := value.(ast.Value)
			if ok && (astVal.IsMissing() || astVal.IsNull()) {
				ctx.AddIssue("text.required", "required")
				return nil, ctx.Error()
			}
		}

		out, err := b.textSchema.ValidateAny(value, ctx.Options)
		if err != nil {
			if vErr, ok := err.(*issues.ValidationError); ok {
				basePath := ctx.PathString()
				for _, issue := range vErr.Issues {
					issue.Path = joinIssuePath(basePath, issue.Path)
					ctx.Issues.Add(issue)
				}
				return nil, nil
			}
			return nil, err
		}
		return fn(out)
	}

	compiled, err := newFieldFromInfo(b.fieldInfo, validator)
	if err != nil {
		b.schema.buildError = err
		return b.schema
	}
	compiled.required = b.required

	if b.fieldIndex == -1 {
		b.schema.fields = append(b.schema.fields, compiled)
		b.schema.lastFieldIndex = len(b.schema.fields) - 1
	} else {
		b.schema.fields[b.fieldIndex] = compiled
		b.schema.lastFieldIndex = b.fieldIndex
	}

	return b.schema
}

// --- boolean_field_builder ---

type BooleanFieldBuilder[T any] struct {
	schema        *Schema[T]
	fieldInfo     fieldInfo[T]
	booleanSchema *boolean.Schema
	fieldIndex    int
	required      bool
}

func (b *BooleanFieldBuilder[T]) build() {
	if b.schema.buildError != nil {
		return
	}

	validator := func(ctx *engine.Context, value any) (any, error) {
		if b.required {
			astVal, ok := value.(ast.Value)
			if ok && (astVal.IsMissing() || astVal.IsNull()) {
				ctx.AddIssue("boolean.required", "required")
				return nil, nil
			}
		}
		out, err := b.booleanSchema.ValidateAny(value, ctx.Options)
		if err != nil {
			if vErr, ok := err.(*issues.ValidationError); ok {
				basePath := ctx.PathString()
				for _, issue := range vErr.Issues {
					issue.Path = joinIssuePath(basePath, issue.Path)
					ctx.Issues.Add(issue)
				}
				return nil, nil
			}
			return nil, err
		}
		return out, nil
	}

	compiled, err := newFieldFromInfo(b.fieldInfo, validator)
	if err != nil {
		b.schema.buildError = err
		return
	}
	compiled.required = b.required

	if b.fieldIndex == -1 {
		b.schema.fields = append(b.schema.fields, compiled)
		b.schema.lastFieldIndex = len(b.schema.fields) - 1
		b.fieldIndex = b.schema.lastFieldIndex
	} else {
		b.schema.fields[b.fieldIndex] = compiled
		b.schema.lastFieldIndex = b.fieldIndex
	}
}

func (b *BooleanFieldBuilder[T]) Required() *BooleanFieldBuilder[T] {
	b.booleanSchema.Required()
	b.required = true
	b.build()
	return b
}

func (b *BooleanFieldBuilder[T]) IsDefault() *BooleanFieldBuilder[T] {
	b.booleanSchema.IsDefault()
	b.build()
	return b
}

func (b *BooleanFieldBuilder[T]) Default(value bool) *BooleanFieldBuilder[T] {
	b.booleanSchema.Default(value)
	b.build()
	return b
}

func (b *BooleanFieldBuilder[T]) DefaultFunc(fn func() bool) *BooleanFieldBuilder[T] {
	b.booleanSchema.DefaultFunc(fn)
	b.build()
	return b
}

func (b *BooleanFieldBuilder[T]) RequiredIf(path string, op ConditionOp, expected any) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.RequiredIf(path, op, expected)
	return b
}

func (b *BooleanFieldBuilder[T]) RequiredWith(paths ...string) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.RequiredWith(paths...)
	return b
}

func (b *BooleanFieldBuilder[T]) RequiredWithout(paths ...string) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.RequiredWithout(paths...)
	return b
}

func (b *BooleanFieldBuilder[T]) ExcludedIf(path string, op ConditionOp, expected any) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.ExcludedIf(path, op, expected)
	return b
}

func (b *BooleanFieldBuilder[T]) SkipUnless(path string, op ConditionOp, expected any) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.SkipUnless(path, op, expected)
	return b
}

func (b *BooleanFieldBuilder[T]) EqField(other string) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.EqField(other)
	return b
}

func (b *BooleanFieldBuilder[T]) NeField(other string) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.NeField(other)
	return b
}

func (b *BooleanFieldBuilder[T]) GtField(other string) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.GtField(other)
	return b
}

func (b *BooleanFieldBuilder[T]) GteField(other string) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.GteField(other)
	return b
}

func (b *BooleanFieldBuilder[T]) LtField(other string) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.LtField(other)
	return b
}

func (b *BooleanFieldBuilder[T]) LteField(other string) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.LteField(other)
	return b
}

func (b *BooleanFieldBuilder[T]) EqCSField(path string) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.EqCSField(path)
	return b
}

func (b *BooleanFieldBuilder[T]) NeCSField(path string) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.NeCSField(path)
	return b
}

func (b *BooleanFieldBuilder[T]) GtCSField(path string) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.GtCSField(path)
	return b
}

func (b *BooleanFieldBuilder[T]) GteCSField(path string) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.GteCSField(path)
	return b
}

func (b *BooleanFieldBuilder[T]) LtCSField(path string) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.LtCSField(path)
	return b
}

func (b *BooleanFieldBuilder[T]) LteCSField(path string) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.LteCSField(path)
	return b
}

func (b *BooleanFieldBuilder[T]) FieldContains(other string) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.FieldContains(other)
	return b
}

func (b *BooleanFieldBuilder[T]) FieldExcludes(other string) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.FieldExcludes(other)
	return b
}

func (b *BooleanFieldBuilder[T]) ValidateAny(value any, options schema.Options) (any, error) {
	return b.booleanSchema.ValidateAny(value, options)
}

func (b *BooleanFieldBuilder[T]) Build() *Schema[T] {
	b.build()
	return b.schema
}

func (b *BooleanFieldBuilder[T]) Transform(fn func(value any) (any, error)) *Schema[T] {
	validator := func(ctx *engine.Context, value any) (any, error) {
		if b.required {
			astVal, ok := value.(ast.Value)
			if ok && (astVal.IsMissing() || astVal.IsNull()) {
				ctx.AddIssue("boolean.required", "required")
				return nil, nil
			}
		}
		out, err := b.booleanSchema.ValidateAny(value, ctx.Options)
		if err != nil {
			if vErr, ok := err.(*issues.ValidationError); ok {
				basePath := ctx.PathString()
				for _, issue := range vErr.Issues {
					issue.Path = joinIssuePath(basePath, issue.Path)
					ctx.Issues.Add(issue)
				}
				return nil, nil
			}
			return nil, err
		}
		return fn(out)
	}

	compiled, err := newFieldFromInfo(b.fieldInfo, validator)
	if err != nil {
		b.schema.buildError = err
		return b.schema
	}
	compiled.required = b.required

	if b.fieldIndex == -1 {
		b.schema.fields = append(b.schema.fields, compiled)
		b.schema.lastFieldIndex = len(b.schema.fields) - 1
	} else {
		b.schema.fields[b.fieldIndex] = compiled
		b.schema.lastFieldIndex = b.fieldIndex
	}

	return b.schema
}

// --- date_field_builder ---

type DateFieldBuilder[T any] struct {
	schema     *Schema[T]
	fieldInfo  fieldInfo[T]
	dateSchema *date.Schema
	fieldIndex int
	required   bool
}

func (b *DateFieldBuilder[T]) build() {
	if b.schema.buildError != nil {
		return
	}

	validator := func(ctx *engine.Context, value any) (any, error) {
		if b.required {
			astVal, ok := value.(ast.Value)
			if ok && (astVal.IsMissing() || astVal.IsNull()) {
				ctx.AddIssue("date.required", "required")
				return nil, nil
			}
		}
		out, err := b.dateSchema.ValidateAny(value, ctx.Options)
		if err != nil {
			if vErr, ok := err.(*issues.ValidationError); ok {
				basePath := ctx.PathString()
				for _, issue := range vErr.Issues {
					issue.Path = joinIssuePath(basePath, issue.Path)
					ctx.Issues.Add(issue)
				}
				return nil, nil
			}
			return nil, err
		}
		return out, nil
	}

	compiled, err := newFieldFromInfo(b.fieldInfo, validator)
	if err != nil {
		b.schema.buildError = err
		return
	}
	compiled.required = b.required

	if b.fieldIndex == -1 {
		b.schema.fields = append(b.schema.fields, compiled)
		b.schema.lastFieldIndex = len(b.schema.fields) - 1
		b.fieldIndex = b.schema.lastFieldIndex
	} else {
		b.schema.fields[b.fieldIndex] = compiled
		b.schema.lastFieldIndex = b.fieldIndex
	}
}

func (b *DateFieldBuilder[T]) Required() *DateFieldBuilder[T] {
	b.dateSchema.Required()
	b.required = true
	b.build()
	return b
}

func (b *DateFieldBuilder[T]) IsDefault() *DateFieldBuilder[T] {
	b.dateSchema.IsDefault()
	b.build()
	return b
}

func (b *DateFieldBuilder[T]) Default(value time.Time) *DateFieldBuilder[T] {
	b.dateSchema.Default(value)
	b.build()
	return b
}

func (b *DateFieldBuilder[T]) DefaultFunc(fn func() time.Time) *DateFieldBuilder[T] {
	b.dateSchema.DefaultFunc(fn)
	b.build()
	return b
}

func (b *DateFieldBuilder[T]) Min(value time.Time) *DateFieldBuilder[T] {
	b.dateSchema.Min(value)
	b.build()
	return b
}

func (b *DateFieldBuilder[T]) Max(value time.Time) *DateFieldBuilder[T] {
	b.dateSchema.Max(value)
	b.build()
	return b
}

func (b *DateFieldBuilder[T]) RequiredIf(path string, op ConditionOp, expected any) *DateFieldBuilder[T] {
	b.build()
	b.schema.RequiredIf(path, op, expected)
	return b
}

func (b *DateFieldBuilder[T]) RequiredWith(paths ...string) *DateFieldBuilder[T] {
	b.build()
	b.schema.RequiredWith(paths...)
	return b
}

func (b *DateFieldBuilder[T]) RequiredWithout(paths ...string) *DateFieldBuilder[T] {
	b.build()
	b.schema.RequiredWithout(paths...)
	return b
}

func (b *DateFieldBuilder[T]) ExcludedIf(path string, op ConditionOp, expected any) *DateFieldBuilder[T] {
	b.build()
	b.schema.ExcludedIf(path, op, expected)
	return b
}

func (b *DateFieldBuilder[T]) SkipUnless(path string, op ConditionOp, expected any) *DateFieldBuilder[T] {
	b.build()
	b.schema.SkipUnless(path, op, expected)
	return b
}

func (b *DateFieldBuilder[T]) EqField(other string) *DateFieldBuilder[T] {
	b.build()
	b.schema.EqField(other)
	return b
}

func (b *DateFieldBuilder[T]) NeField(other string) *DateFieldBuilder[T] {
	b.build()
	b.schema.NeField(other)
	return b
}

func (b *DateFieldBuilder[T]) GtField(other string) *DateFieldBuilder[T] {
	b.build()
	b.schema.GtField(other)
	return b
}

func (b *DateFieldBuilder[T]) GteField(other string) *DateFieldBuilder[T] {
	b.build()
	b.schema.GteField(other)
	return b
}

func (b *DateFieldBuilder[T]) LtField(other string) *DateFieldBuilder[T] {
	b.build()
	b.schema.LtField(other)
	return b
}

func (b *DateFieldBuilder[T]) LteField(other string) *DateFieldBuilder[T] {
	b.build()
	b.schema.LteField(other)
	return b
}

func (b *DateFieldBuilder[T]) EqCSField(path string) *DateFieldBuilder[T] {
	b.build()
	b.schema.EqCSField(path)
	return b
}

func (b *DateFieldBuilder[T]) NeCSField(path string) *DateFieldBuilder[T] {
	b.build()
	b.schema.NeCSField(path)
	return b
}

func (b *DateFieldBuilder[T]) GtCSField(path string) *DateFieldBuilder[T] {
	b.build()
	b.schema.GtCSField(path)
	return b
}

func (b *DateFieldBuilder[T]) GteCSField(path string) *DateFieldBuilder[T] {
	b.build()
	b.schema.GteCSField(path)
	return b
}

func (b *DateFieldBuilder[T]) LtCSField(path string) *DateFieldBuilder[T] {
	b.build()
	b.schema.LtCSField(path)
	return b
}

func (b *DateFieldBuilder[T]) LteCSField(path string) *DateFieldBuilder[T] {
	b.build()
	b.schema.LteCSField(path)
	return b
}

func (b *DateFieldBuilder[T]) FieldContains(other string) *DateFieldBuilder[T] {
	b.build()
	b.schema.FieldContains(other)
	return b
}

func (b *DateFieldBuilder[T]) FieldExcludes(other string) *DateFieldBuilder[T] {
	b.build()
	b.schema.FieldExcludes(other)
	return b
}

func (b *DateFieldBuilder[T]) ValidateAny(value any, options schema.Options) (any, error) {
	return b.dateSchema.ValidateAny(value, options)
}

func (b *DateFieldBuilder[T]) Build() *Schema[T] {
	b.build()
	return b.schema
}

func (b *DateFieldBuilder[T]) Transform(fn func(value any) (any, error)) *Schema[T] {
	validator := func(ctx *engine.Context, value any) (any, error) {
		if b.required {
			astVal, ok := value.(ast.Value)
			if ok && (astVal.IsMissing() || astVal.IsNull()) {
				ctx.AddIssue("date.required", "required")
				return nil, nil
			}
		}
		out, err := b.dateSchema.ValidateAny(value, ctx.Options)
		if err != nil {
			if vErr, ok := err.(*issues.ValidationError); ok {
				basePath := ctx.PathString()
				for _, issue := range vErr.Issues {
					issue.Path = joinIssuePath(basePath, issue.Path)
					ctx.Issues.Add(issue)
				}
				return nil, nil
			}
			return nil, err
		}
		return fn(out)
	}

	compiled, err := newFieldFromInfo(b.fieldInfo, validator)
	if err != nil {
		b.schema.buildError = err
		return b.schema
	}
	compiled.required = b.required

	if b.fieldIndex == -1 {
		b.schema.fields = append(b.schema.fields, compiled)
		b.schema.lastFieldIndex = len(b.schema.fields) - 1
	} else {
		b.schema.fields[b.fieldIndex] = compiled
		b.schema.lastFieldIndex = b.fieldIndex
	}

	return b.schema
}

// --- duration_field_builder ---

type DurationFieldBuilder[T any] struct {
	schema         *Schema[T]
	fieldInfo      fieldInfo[T]
	durationSchema *duration.Schema
	fieldIndex     int
	required       bool
}

func (b *DurationFieldBuilder[T]) build() {
	if b.schema.buildError != nil {
		return
	}

	validator := func(ctx *engine.Context, value any) (any, error) {
		if b.required {
			astVal, ok := value.(ast.Value)
			if ok && (astVal.IsMissing() || astVal.IsNull()) {
				ctx.AddIssue("duration.required", "required")
				return nil, nil
			}
		}
		out, err := b.durationSchema.ValidateAny(value, ctx.Options)
		if err != nil {
			if vErr, ok := err.(*issues.ValidationError); ok {
				basePath := ctx.PathString()
				for _, issue := range vErr.Issues {
					issue.Path = joinIssuePath(basePath, issue.Path)
					ctx.Issues.Add(issue)
				}
				return nil, nil
			}
			return nil, err
		}
		return out, nil
	}

	compiled, err := newFieldFromInfo(b.fieldInfo, validator)
	if err != nil {
		b.schema.buildError = err
		return
	}
	compiled.required = b.required

	if b.fieldIndex == -1 {
		b.schema.fields = append(b.schema.fields, compiled)
		b.schema.lastFieldIndex = len(b.schema.fields) - 1
		b.fieldIndex = b.schema.lastFieldIndex
	} else {
		b.schema.fields[b.fieldIndex] = compiled
		b.schema.lastFieldIndex = b.fieldIndex
	}
}

func (b *DurationFieldBuilder[T]) Required() *DurationFieldBuilder[T] {
	b.durationSchema.Required()
	b.required = true
	b.build()
	return b
}

func (b *DurationFieldBuilder[T]) IsDefault() *DurationFieldBuilder[T] {
	b.durationSchema.IsDefault()
	b.build()
	return b
}

func (b *DurationFieldBuilder[T]) Default(value time.Duration) *DurationFieldBuilder[T] {
	b.durationSchema.Default(value)
	b.build()
	return b
}

func (b *DurationFieldBuilder[T]) DefaultFunc(fn func() time.Duration) *DurationFieldBuilder[T] {
	b.durationSchema.DefaultFunc(fn)
	b.build()
	return b
}

func (b *DurationFieldBuilder[T]) Min(value time.Duration) *DurationFieldBuilder[T] {
	b.durationSchema.Min(value)
	b.build()
	return b
}

func (b *DurationFieldBuilder[T]) Max(value time.Duration) *DurationFieldBuilder[T] {
	b.durationSchema.Max(value)
	b.build()
	return b
}

func (b *DurationFieldBuilder[T]) RequiredIf(path string, op ConditionOp, expected any) *DurationFieldBuilder[T] {
	b.build()
	b.schema.RequiredIf(path, op, expected)
	return b
}

func (b *DurationFieldBuilder[T]) RequiredWith(paths ...string) *DurationFieldBuilder[T] {
	b.build()
	b.schema.RequiredWith(paths...)
	return b
}

func (b *DurationFieldBuilder[T]) RequiredWithout(paths ...string) *DurationFieldBuilder[T] {
	b.build()
	b.schema.RequiredWithout(paths...)
	return b
}

func (b *DurationFieldBuilder[T]) ExcludedIf(path string, op ConditionOp, expected any) *DurationFieldBuilder[T] {
	b.build()
	b.schema.ExcludedIf(path, op, expected)
	return b
}

func (b *DurationFieldBuilder[T]) SkipUnless(path string, op ConditionOp, expected any) *DurationFieldBuilder[T] {
	b.build()
	b.schema.SkipUnless(path, op, expected)
	return b
}

func (b *DurationFieldBuilder[T]) EqField(other string) *DurationFieldBuilder[T] {
	b.build()
	b.schema.EqField(other)
	return b
}

func (b *DurationFieldBuilder[T]) NeField(other string) *DurationFieldBuilder[T] {
	b.build()
	b.schema.NeField(other)
	return b
}

func (b *DurationFieldBuilder[T]) GtField(other string) *DurationFieldBuilder[T] {
	b.build()
	b.schema.GtField(other)
	return b
}

func (b *DurationFieldBuilder[T]) GteField(other string) *DurationFieldBuilder[T] {
	b.build()
	b.schema.GteField(other)
	return b
}

func (b *DurationFieldBuilder[T]) LtField(other string) *DurationFieldBuilder[T] {
	b.build()
	b.schema.LtField(other)
	return b
}

func (b *DurationFieldBuilder[T]) LteField(other string) *DurationFieldBuilder[T] {
	b.build()
	b.schema.LteField(other)
	return b
}

func (b *DurationFieldBuilder[T]) EqCSField(path string) *DurationFieldBuilder[T] {
	b.build()
	b.schema.EqCSField(path)
	return b
}

func (b *DurationFieldBuilder[T]) NeCSField(path string) *DurationFieldBuilder[T] {
	b.build()
	b.schema.NeCSField(path)
	return b
}

func (b *DurationFieldBuilder[T]) GtCSField(path string) *DurationFieldBuilder[T] {
	b.build()
	b.schema.GtCSField(path)
	return b
}

func (b *DurationFieldBuilder[T]) GteCSField(path string) *DurationFieldBuilder[T] {
	b.build()
	b.schema.GteCSField(path)
	return b
}

func (b *DurationFieldBuilder[T]) LtCSField(path string) *DurationFieldBuilder[T] {
	b.build()
	b.schema.LtCSField(path)
	return b
}

func (b *DurationFieldBuilder[T]) LteCSField(path string) *DurationFieldBuilder[T] {
	b.build()
	b.schema.LteCSField(path)
	return b
}

func (b *DurationFieldBuilder[T]) FieldContains(other string) *DurationFieldBuilder[T] {
	b.build()
	b.schema.FieldContains(other)
	return b
}

func (b *DurationFieldBuilder[T]) FieldExcludes(other string) *DurationFieldBuilder[T] {
	b.build()
	b.schema.FieldExcludes(other)
	return b
}

func (b *DurationFieldBuilder[T]) ValidateAny(value any, options schema.Options) (any, error) {
	return b.durationSchema.ValidateAny(value, options)
}

func (b *DurationFieldBuilder[T]) Build() *Schema[T] {
	b.build()
	return b.schema
}

func (b *DurationFieldBuilder[T]) Transform(fn func(value any) (any, error)) *Schema[T] {
	validator := func(ctx *engine.Context, value any) (any, error) {
		if b.required {
			astVal, ok := value.(ast.Value)
			if ok && (astVal.IsMissing() || astVal.IsNull()) {
				ctx.AddIssue("duration.required", "required")
				return nil, nil
			}
		}
		out, err := b.durationSchema.ValidateAny(value, ctx.Options)
		if err != nil {
			if vErr, ok := err.(*issues.ValidationError); ok {
				basePath := ctx.PathString()
				for _, issue := range vErr.Issues {
					issue.Path = joinIssuePath(basePath, issue.Path)
					ctx.Issues.Add(issue)
				}
				return nil, nil
			}
			return nil, err
		}
		return fn(out)
	}

	compiled, err := newFieldFromInfo(b.fieldInfo, validator)
	if err != nil {
		b.schema.buildError = err
		return b.schema
	}
	compiled.required = b.required

	if b.fieldIndex == -1 {
		b.schema.fields = append(b.schema.fields, compiled)
		b.schema.lastFieldIndex = len(b.schema.fields) - 1
	} else {
		b.schema.fields[b.fieldIndex] = compiled
		b.schema.lastFieldIndex = b.fieldIndex
	}

	return b.schema
}

// --- number_field_builder ---

type NumberFieldBuilder[T any] struct {
	schema       *Schema[T]
	fieldInfo    fieldInfo[T]
	required     bool
	isDefault    bool
	minSet       bool
	maxSet       bool
	minValue     float64
	maxValue     float64
	fieldIndex   int
	hasDefault   bool
	defaultValue float64
	defaultFunc  func() float64
}

func (b *NumberFieldBuilder[T]) build() {
	if b.schema.buildError != nil {
		return
	}

	fieldType := b.fieldInfo.fieldType
	minSet := b.minSet
	maxSet := b.maxSet
	minValue := b.minValue
	maxValue := b.maxValue
	required := b.required

	validator := func(ctx *engine.Context, value any) (any, error) {
		astVal, ok := value.(ast.Value)
		if !ok {
			coerced, err := engine.InputToASTWithOptions(value, ctx.Options)
			if err != nil {
				return nil, err
			}
			astVal = coerced
		}

		if astVal.IsMissing() || astVal.IsNull() {
			if required {
				ctx.AddIssue("number.required", "required")
				return nil, ctx.Error()
			}
			if b.defaultFunc != nil {
				return convertToFieldType(b.defaultFunc(), fieldType), nil
			}
			if b.hasDefault {
				return convertToFieldType(b.defaultValue, fieldType), nil
			}
			return reflect.Zero(fieldType).Interface(), nil
		}

		if astVal.Kind != ast.KindNumber {
			ctx.AddIssue("number.type", "expected number", map[string]any{
				"expected": "number",
				"actual":   astVal.Kind.String(),
			})
			return nil, ctx.Error()
		}

		floatVal, err := strconv.ParseFloat(astVal.Number, 64)
		if err != nil {
			ctx.AddIssue("number.parse", "invalid number", map[string]any{"error": err.Error()})
			return nil, ctx.Error()
		}

		if minSet && floatVal < minValue {
			ctx.AddIssue("number.min", "too small", map[string]any{"min": minValue, "actual": floatVal})
			return nil, ctx.Error()
		}

		if maxSet && floatVal > maxValue {
			ctx.AddIssue("number.max", "too large", map[string]any{"max": maxValue, "actual": floatVal})
			return nil, ctx.Error()
		}

		return convertToFieldType(floatVal, fieldType), nil
	}

	compiled, err := newFieldFromInfo(b.fieldInfo, validator)
	if err != nil {
		b.schema.buildError = err
		return
	}
	compiled.required = b.required

	if b.fieldIndex == -1 {
		b.schema.fields = append(b.schema.fields, compiled)
		b.schema.lastFieldIndex = len(b.schema.fields) - 1
		b.fieldIndex = b.schema.lastFieldIndex
	} else {
		b.schema.fields[b.fieldIndex] = compiled
		b.schema.lastFieldIndex = b.fieldIndex
	}
}

func (b *NumberFieldBuilder[T]) Required() *NumberFieldBuilder[T] {
	b.required = true
	b.build()
	return b
}

func (b *NumberFieldBuilder[T]) IsDefault() *NumberFieldBuilder[T] {
	b.isDefault = true
	b.build()
	return b
}

func (b *NumberFieldBuilder[T]) Default(value float64) *NumberFieldBuilder[T] {
	b.hasDefault = true
	b.defaultValue = value
	b.defaultFunc = nil
	b.build()
	return b
}

func (b *NumberFieldBuilder[T]) DefaultFunc(fn func() float64) *NumberFieldBuilder[T] {
	b.defaultFunc = fn
	b.hasDefault = false
	b.build()
	return b
}

func (b *NumberFieldBuilder[T]) Min(value float64) *NumberFieldBuilder[T] {
	b.minSet = true
	b.minValue = value
	b.build()
	return b
}

func (b *NumberFieldBuilder[T]) Max(value float64) *NumberFieldBuilder[T] {
	b.maxSet = true
	b.maxValue = value
	b.build()
	return b
}

func (b *NumberFieldBuilder[T]) Integer() *NumberFieldBuilder[T] {
	b.build()
	return b
}

func (b *NumberFieldBuilder[T]) RequiredIf(path string, op ConditionOp, expected any) *NumberFieldBuilder[T] {
	b.build()
	b.schema.RequiredIf(path, op, expected)
	return b
}

func (b *NumberFieldBuilder[T]) RequiredWith(paths ...string) *NumberFieldBuilder[T] {
	b.build()
	b.schema.RequiredWith(paths...)
	return b
}

func (b *NumberFieldBuilder[T]) RequiredWithout(paths ...string) *NumberFieldBuilder[T] {
	b.build()
	b.schema.RequiredWithout(paths...)
	return b
}

func (b *NumberFieldBuilder[T]) ExcludedIf(path string, op ConditionOp, expected any) *NumberFieldBuilder[T] {
	b.build()
	b.schema.ExcludedIf(path, op, expected)
	return b
}

func (b *NumberFieldBuilder[T]) SkipUnless(path string, op ConditionOp, expected any) *NumberFieldBuilder[T] {
	b.build()
	b.schema.SkipUnless(path, op, expected)
	return b
}

func (b *NumberFieldBuilder[T]) EqField(other string) *NumberFieldBuilder[T] {
	b.build()
	b.schema.EqField(other)
	return b
}

func (b *NumberFieldBuilder[T]) NeField(other string) *NumberFieldBuilder[T] {
	b.build()
	b.schema.NeField(other)
	return b
}

func (b *NumberFieldBuilder[T]) GtField(other string) *NumberFieldBuilder[T] {
	b.build()
	b.schema.GtField(other)
	return b
}

func (b *NumberFieldBuilder[T]) GteField(other string) *NumberFieldBuilder[T] {
	b.build()
	b.schema.GteField(other)
	return b
}

func (b *NumberFieldBuilder[T]) LtField(other string) *NumberFieldBuilder[T] {
	b.build()
	b.schema.LtField(other)
	return b
}

func (b *NumberFieldBuilder[T]) LteField(other string) *NumberFieldBuilder[T] {
	b.build()
	b.schema.LteField(other)
	return b
}

func (b *NumberFieldBuilder[T]) EqCSField(path string) *NumberFieldBuilder[T] {
	b.build()
	b.schema.EqCSField(path)
	return b
}

func (b *NumberFieldBuilder[T]) NeCSField(path string) *NumberFieldBuilder[T] {
	b.build()
	b.schema.NeCSField(path)
	return b
}

func (b *NumberFieldBuilder[T]) GtCSField(path string) *NumberFieldBuilder[T] {
	b.build()
	b.schema.GtCSField(path)
	return b
}

func (b *NumberFieldBuilder[T]) GteCSField(path string) *NumberFieldBuilder[T] {
	b.build()
	b.schema.GteCSField(path)
	return b
}

func (b *NumberFieldBuilder[T]) LtCSField(path string) *NumberFieldBuilder[T] {
	b.build()
	b.schema.LtCSField(path)
	return b
}

func (b *NumberFieldBuilder[T]) LteCSField(path string) *NumberFieldBuilder[T] {
	b.build()
	b.schema.LteCSField(path)
	return b
}

func (b *NumberFieldBuilder[T]) FieldContains(other string) *NumberFieldBuilder[T] {
	b.build()
	b.schema.FieldContains(other)
	return b
}

func (b *NumberFieldBuilder[T]) FieldExcludes(other string) *NumberFieldBuilder[T] {
	b.build()
	b.schema.FieldExcludes(other)
	return b
}

func (b *NumberFieldBuilder[T]) ValidateAny(value any, options schema.Options) (any, error) {
	return nil, nil
}

func (b *NumberFieldBuilder[T]) Build() *Schema[T] {
	b.build()
	return b.schema
}

func (b *NumberFieldBuilder[T]) Transform(fn func(value any) (any, error)) *Schema[T] {
	fieldType := b.fieldInfo.fieldType
	minSet := b.minSet
	maxSet := b.maxSet
	minValue := b.minValue
	maxValue := b.maxValue
	required := b.required

	validator := func(ctx *engine.Context, value any) (any, error) {
		astVal, ok := value.(ast.Value)
		if !ok {
			coerced, err := engine.InputToASTWithOptions(value, ctx.Options)
			if err != nil {
				return nil, err
			}
			astVal = coerced
		}

		if astVal.IsMissing() || astVal.IsNull() {
			if required {
				ctx.AddIssue("number.required", "required")
				return nil, ctx.Error()
			}
			if b.defaultFunc != nil {
				return convertToFieldType(b.defaultFunc(), fieldType), nil
			}
			if b.hasDefault {
				return convertToFieldType(b.defaultValue, fieldType), nil
			}
			return reflect.Zero(fieldType).Interface(), nil
		}

		if astVal.Kind != ast.KindNumber {
			ctx.AddIssue("number.type", "expected number", map[string]any{
				"expected": "number",
				"actual":   astVal.Kind.String(),
			})
			return nil, ctx.Error()
		}

		floatVal, err := strconv.ParseFloat(astVal.Number, 64)
		if err != nil {
			ctx.AddIssue("number.parse", "invalid number", map[string]any{"error": err.Error()})
			return nil, ctx.Error()
		}

		if minSet && floatVal < minValue {
			ctx.AddIssue("number.min", "too small", map[string]any{"min": minValue, "actual": floatVal})
			return nil, ctx.Error()
		}

		if maxSet && floatVal > maxValue {
			ctx.AddIssue("number.max", "too large", map[string]any{"max": maxValue, "actual": floatVal})
			return nil, ctx.Error()
		}

		return fn(convertToFieldType(floatVal, fieldType))
	}

	compiled, err := newFieldFromInfo(b.fieldInfo, validator)
	if err != nil {
		b.schema.buildError = err
		return b.schema
	}
	compiled.required = b.required

	if b.fieldIndex == -1 {
		b.schema.fields = append(b.schema.fields, compiled)
		b.schema.lastFieldIndex = len(b.schema.fields) - 1
	} else {
		b.schema.fields[b.fieldIndex] = compiled
		b.schema.lastFieldIndex = b.fieldIndex
	}

	return b.schema
}

func convertToFieldType(floatVal float64, fieldType reflect.Type) any {
	switch fieldType.Kind() {
	case reflect.Int:
		return int(floatVal)
	case reflect.Int8:
		return int8(floatVal)
	case reflect.Int16:
		return int16(floatVal)
	case reflect.Int32:
		return int32(floatVal)
	case reflect.Int64:
		return int64(floatVal)
	case reflect.Uint:
		return uint(floatVal)
	case reflect.Uint8:
		return uint8(floatVal)
	case reflect.Uint16:
		return uint16(floatVal)
	case reflect.Uint32:
		return uint32(floatVal)
	case reflect.Uint64:
		return uint64(floatVal)
	case reflect.Float32:
		return float32(floatVal)
	case reflect.Float64:
		return floatVal
	default:
		return floatVal
	}
}

// --- array_field_builder ---

type ArrayFieldBuilder[T any] struct {
	schema    *Schema[T]
	fieldInfo fieldInfo[T]

	required  bool
	isDefault bool
	min       *int
	max       *int
	len       *int
	eq        *int
	ne        *int
	gt        *int
	gte       *int
	lt        *int
	lte       *int
	unique    bool

	itemSchema schema.AnySchema

	fieldIndex int
}

func (b *ArrayFieldBuilder[T]) Required() *ArrayFieldBuilder[T] {
	b.required = true
	return b.build()
}

func (b *ArrayFieldBuilder[T]) IsDefault() *ArrayFieldBuilder[T] {
	b.isDefault = true
	return b.build()
}

func (b *ArrayFieldBuilder[T]) Min(min int) *ArrayFieldBuilder[T] {
	b.min = &min
	return b.build()
}

func (b *ArrayFieldBuilder[T]) Max(max int) *ArrayFieldBuilder[T] {
	b.max = &max
	return b.build()
}

func (b *ArrayFieldBuilder[T]) Len(len int) *ArrayFieldBuilder[T] {
	b.len = &len
	return b.build()
}

func (b *ArrayFieldBuilder[T]) Eq(len int) *ArrayFieldBuilder[T] {
	b.eq = &len
	return b.build()
}

func (b *ArrayFieldBuilder[T]) Ne(len int) *ArrayFieldBuilder[T] {
	b.ne = &len
	return b.build()
}

func (b *ArrayFieldBuilder[T]) Gt(len int) *ArrayFieldBuilder[T] {
	b.gt = &len
	return b.build()
}

func (b *ArrayFieldBuilder[T]) Gte(len int) *ArrayFieldBuilder[T] {
	b.gte = &len
	return b.build()
}

func (b *ArrayFieldBuilder[T]) Lt(len int) *ArrayFieldBuilder[T] {
	b.lt = &len
	return b.build()
}

func (b *ArrayFieldBuilder[T]) Lte(len int) *ArrayFieldBuilder[T] {
	b.lte = &len
	return b.build()
}

func (b *ArrayFieldBuilder[T]) Unique() *ArrayFieldBuilder[T] {
	b.unique = true
	return b.build()
}

func (b *ArrayFieldBuilder[T]) build() *ArrayFieldBuilder[T] {
	sliceType := b.fieldInfo.fieldType
	if sliceType.Kind() != reflect.Slice {
		return b
	}

	validator := func(context *engine.Context, value any) (any, error) {
		val, ok := value.(ast.Value)
		if !ok {
			return nil, nil
		}

		if val.IsMissing() {
			if b.required {
				context.AddIssue("object.required", "required")
				return nil, nil
			}
			return nil, nil
		}
		if val.IsNull() {
			return nil, nil
		}

		if val.Kind != ast.KindArray {
			context.AddIssue("array.type", "expected array")
			return nil, nil
		}

		arr := val.Array
		count := len(arr)

		if b.min != nil && count < *b.min {
			context.AddIssue("array.min", "too short", map[string]any{"min": *b.min, "actual": count})
			return nil, nil
		}
		if b.max != nil && count > *b.max {
			context.AddIssue("array.max", "too long", map[string]any{"max": *b.max, "actual": count})
			return nil, nil
		}
		if b.len != nil && count != *b.len {
			context.AddIssue("array.len", "invalid length", map[string]any{"len": *b.len, "actual": count})
			return nil, nil
		}

		resultSlice := reflect.MakeSlice(sliceType, 0, count)

		if b.itemSchema != nil {
			basePath := context.PathString()

			for i, item := range arr {
				itemRes, err := b.itemSchema.ValidateAny(item, context.Options)
				if err != nil {
					if vErr, ok := err.(*issues.ValidationError); ok {
						for _, issue := range vErr.Issues {
							indexPath := fmt.Sprintf("[%d]", i)

							var itemRelPath string
							if issue.Path != "" {
								if issue.Path[0] == '[' {
									itemRelPath = indexPath + issue.Path
								} else {
									itemRelPath = indexPath + "." + issue.Path
								}
							} else {
								itemRelPath = indexPath
							}

							var fullPath string
							if basePath != "" {
								fullPath = basePath + itemRelPath
							} else {
								fullPath = itemRelPath
							}

							issue.Path = fullPath
							context.Issues.Add(issue)
						}
					} else {
						return nil, err
					}
				}
				if itemRes != nil {
					resultSlice = reflect.Append(resultSlice, reflect.ValueOf(itemRes))
				}
			}
		}

		return resultSlice.Interface(), nil
	}

	compiled, err := newFieldFromInfo(b.fieldInfo, validator)
	if err != nil {
		b.schema.buildError = err
		return b
	}
	compiled.required = b.required

	if b.fieldIndex == -1 {
		b.schema.fields = append(b.schema.fields, compiled)
		b.schema.lastFieldIndex = len(b.schema.fields) - 1
		b.fieldIndex = b.schema.lastFieldIndex
	} else {
		b.schema.fields[b.fieldIndex] = compiled
		b.schema.lastFieldIndex = b.fieldIndex
	}

	return b
}

func (b *ArrayFieldBuilder[T]) RequiredIf(path string, op ConditionOp, expected any) *ArrayFieldBuilder[T] {
	b.build()
	b.schema.RequiredIf(path, op, expected)
	return b
}

func (b *ArrayFieldBuilder[T]) RequiredWith(paths ...string) *ArrayFieldBuilder[T] {
	b.build()
	b.schema.RequiredWith(paths...)
	return b
}

func (b *ArrayFieldBuilder[T]) RequiredWithout(paths ...string) *ArrayFieldBuilder[T] {
	b.build()
	b.schema.RequiredWithout(paths...)
	return b
}

func (b *ArrayFieldBuilder[T]) ExcludedIf(path string, op ConditionOp, expected any) *ArrayFieldBuilder[T] {
	b.build()
	b.schema.ExcludedIf(path, op, expected)
	return b
}

func (b *ArrayFieldBuilder[T]) SkipUnless(path string, op ConditionOp, expected any) *ArrayFieldBuilder[T] {
	b.build()
	b.schema.SkipUnless(path, op, expected)
	return b
}

func (b *ArrayFieldBuilder[T]) EqField(other string) *ArrayFieldBuilder[T] {
	b.build()
	b.schema.EqField(other)
	return b
}

func (b *ArrayFieldBuilder[T]) NeField(other string) *ArrayFieldBuilder[T] {
	b.build()
	b.schema.NeField(other)
	return b
}

func (b *ArrayFieldBuilder[T]) GtField(other string) *ArrayFieldBuilder[T] {
	b.build()
	b.schema.GtField(other)
	return b
}

func (b *ArrayFieldBuilder[T]) GteField(other string) *ArrayFieldBuilder[T] {
	b.build()
	b.schema.GteField(other)
	return b
}

func (b *ArrayFieldBuilder[T]) LtField(other string) *ArrayFieldBuilder[T] {
	b.build()
	b.schema.LtField(other)
	return b
}

func (b *ArrayFieldBuilder[T]) LteField(other string) *ArrayFieldBuilder[T] {
	b.build()
	b.schema.LteField(other)
	return b
}

func (b *ArrayFieldBuilder[T]) EqCSField(path string) *ArrayFieldBuilder[T] {
	b.build()
	b.schema.EqCSField(path)
	return b
}

func (b *ArrayFieldBuilder[T]) NeCSField(path string) *ArrayFieldBuilder[T] {
	b.build()
	b.schema.NeCSField(path)
	return b
}

func (b *ArrayFieldBuilder[T]) GtCSField(path string) *ArrayFieldBuilder[T] {
	b.build()
	b.schema.GtCSField(path)
	return b
}

func (b *ArrayFieldBuilder[T]) GteCSField(path string) *ArrayFieldBuilder[T] {
	b.build()
	b.schema.GteCSField(path)
	return b
}

func (b *ArrayFieldBuilder[T]) LtCSField(path string) *ArrayFieldBuilder[T] {
	b.build()
	b.schema.LtCSField(path)
	return b
}

func (b *ArrayFieldBuilder[T]) LteCSField(path string) *ArrayFieldBuilder[T] {
	b.build()
	b.schema.LteCSField(path)
	return b
}

func (b *ArrayFieldBuilder[T]) FieldContains(other string) *ArrayFieldBuilder[T] {
	b.build()
	b.schema.FieldContains(other)
	return b
}

func (b *ArrayFieldBuilder[T]) FieldExcludes(other string) *ArrayFieldBuilder[T] {
	b.build()
	b.schema.FieldExcludes(other)
	return b
}

func (b *ArrayFieldBuilder[T]) Transform(fn func(value any) (any, error)) *Schema[T] {
	b.build()

	idx := b.fieldIndex
	if idx < 0 || idx >= len(b.schema.fields) {
		return b.schema
	}

	currentField := b.schema.fields[idx]
	originalValidator := currentField.validate

	newValidator := func(ctx *engine.Context, value any) (any, error) {
		out, err := originalValidator(ctx, value)
		if err != nil {
			return nil, err
		}
		return fn(out)
	}

	b.schema.fields[idx].validate = newValidator

	return b.schema
}

// --- record_field_builder ---

type RecordFieldBuilder[T any] struct {
	schema    *Schema[T]
	fieldInfo fieldInfo[T]

	required bool
	min      *int
	max      *int
	len      *int

	keySchema   schema.AnySchema
	valueSchema schema.AnySchema

	fieldIndex int
}

func (b *RecordFieldBuilder[T]) Required() *RecordFieldBuilder[T] {
	b.required = true
	return b.build()
}

func (b *RecordFieldBuilder[T]) Min(min int) *RecordFieldBuilder[T] {
	b.min = &min
	return b.build()
}

func (b *RecordFieldBuilder[T]) Max(max int) *RecordFieldBuilder[T] {
	b.max = &max
	return b.build()
}

func (b *RecordFieldBuilder[T]) Len(len int) *RecordFieldBuilder[T] {
	b.len = &len
	return b.build()
}

func (b *RecordFieldBuilder[T]) build() *RecordFieldBuilder[T] {
	mapType := b.fieldInfo.fieldType
	if mapType.Kind() != reflect.Map {
		return b
	}

	validator := func(context *engine.Context, value any) (any, error) {
		val, ok := value.(ast.Value)
		if !ok {
			return nil, nil
		}

		if val.IsMissing() {
			if b.required {
				context.AddIssue("object.required", "required")
				return nil, nil
			}
			return nil, nil
		}
		if val.IsNull() {
			return nil, nil
		}

		if val.Kind != ast.KindObject {
			context.AddIssue("record.type", "expected object/map")
			return nil, nil
		}

		obj := val.Object
		count := len(obj)

		if b.min != nil && count < *b.min {
			context.AddIssue("record.min", "too few items", map[string]any{"min": *b.min, "actual": count})
			return nil, nil
		}
		if b.max != nil && count > *b.max {
			context.AddIssue("record.max", "too many items", map[string]any{"max": *b.max, "actual": count})
			return nil, nil
		}
		if b.len != nil && count != *b.len {
			context.AddIssue("record.len", "invalid length", map[string]any{"len": *b.len, "actual": count})
			return nil, nil
		}

		resultMap := reflect.MakeMapWithSize(mapType, count)
		basePath := context.PathString()

		for key, item := range obj {
			if b.keySchema != nil {
				keyVal := ast.StringValue(key)
				_, err := b.keySchema.ValidateAny(keyVal, context.Options)
				if err != nil {
					if vErr, ok := err.(*issues.ValidationError); ok {
						for _, issue := range vErr.Issues {
							context.AddIssue("record.key", "invalid key", map[string]any{"key": key, "details": issue.Message})
						}
					}
				}
			}

			if b.valueSchema != nil {
				itemRes, err := b.valueSchema.ValidateAny(item, context.Options)
				if err != nil {
					if vErr, ok := err.(*issues.ValidationError); ok {
						for _, issue := range vErr.Issues {
							var itemRelPath string
							if issue.Path != "" {
								if issue.Path[0] == '[' {
									itemRelPath = key + issue.Path
								} else {
									itemRelPath = key + "." + issue.Path
								}
							} else {
								itemRelPath = key
							}

							var fullPath string
							if basePath != "" {
								fullPath = basePath + "." + itemRelPath
							} else {
								fullPath = itemRelPath
							}

							issue.Path = fullPath
							context.Issues.Add(issue)
						}
					} else {
						return nil, err
					}
				}
				if itemRes != nil {
					keyType := mapType.Key()
					var keyVal reflect.Value
					if keyType.Kind() == reflect.String {
						keyVal = reflect.ValueOf(key)
					} else {
						continue
					}

					resultMap.SetMapIndex(keyVal, reflect.ValueOf(itemRes))
				}
			}
		}

		return resultMap.Interface(), nil
	}

	compiled, err := newFieldFromInfo(b.fieldInfo, validator)
	if err != nil {
		b.schema.buildError = err
		return b
	}
	compiled.required = b.required

	if b.fieldIndex == -1 {
		b.schema.fields = append(b.schema.fields, compiled)
		b.schema.lastFieldIndex = len(b.schema.fields) - 1
		b.fieldIndex = b.schema.lastFieldIndex
	} else {
		b.schema.fields[b.fieldIndex] = compiled
		b.schema.lastFieldIndex = b.fieldIndex
	}

	return b
}

func (b *RecordFieldBuilder[T]) Transform(fn func(value any) (any, error)) *Schema[T] {
	b.build()

	idx := b.fieldIndex
	if idx < 0 || idx >= len(b.schema.fields) {
		return b.schema
	}

	currentField := b.schema.fields[idx]
	originalValidator := currentField.validate

	newValidator := func(ctx *engine.Context, value any) (any, error) {
		out, err := originalValidator(ctx, value)
		if err != nil {
			return nil, err
		}
		return fn(out)
	}

	b.schema.fields[idx].validate = newValidator

	return b.schema
}

// --- object_field_builder ---

type ObjectFieldBuilder[T any] struct {
	schema    *Schema[T]
	fieldInfo fieldInfo[T]

	required      bool
	structOnly    bool
	noStructLevel bool
	builderFunc   any

	fieldIndex int
}

func (b *ObjectFieldBuilder[T]) Required() *ObjectFieldBuilder[T] {
	b.required = true
	return b.build()
}

func (b *ObjectFieldBuilder[T]) StructOnly() *ObjectFieldBuilder[T] {
	b.structOnly = true
	b.noStructLevel = false
	return b.build()
}

func (b *ObjectFieldBuilder[T]) NoStructLevel() *ObjectFieldBuilder[T] {
	b.noStructLevel = true
	b.structOnly = false
	return b.build()
}

func (b *ObjectFieldBuilder[T]) build() *ObjectFieldBuilder[T] {
	targetType := b.fieldInfo.fieldType

	validator := func(context *engine.Context, value any) (any, error) {
		val, ok := value.(ast.Value)
		if !ok {
			return nil, nil
		}

		if val.IsMissing() {
			if b.required {
				context.AddIssue("object.required", "required")
				return nil, nil
			}
			return nil, nil
		}
		if val.IsNull() {
			return nil, nil
		}

		if b.builderFunc == nil {
			return nil, nil
		}

		fnVal := reflect.ValueOf(b.builderFunc)

		nestedType := targetType
		if nestedType.Kind() == reflect.Ptr {
			nestedType = nestedType.Elem()
		}
		nestedPtr := reflect.New(nestedType)

		funcType := fnVal.Type()
		if funcType.NumIn() != 2 {
			return nil, fmt.Errorf("invalid builder signature: expected 2 arguments")
		}

		schemaTypePtr := funcType.In(1)
		schemaType := schemaTypePtr.Elem()

		nestedSchemaVal := reflect.New(schemaType)

		fLastFieldIndex := nestedSchemaVal.Elem().FieldByName("lastFieldIndex")
		if fLastFieldIndex.IsValid() && fLastFieldIndex.CanSet() {
			fLastFieldIndex.SetInt(-1)
		}

		fBuildTarget := nestedSchemaVal.Elem().FieldByName("buildTarget")
		if fBuildTarget.IsValid() && fBuildTarget.CanSet() {
			fBuildTarget.Set(nestedPtr)
		}

		args := []reflect.Value{nestedPtr, nestedSchemaVal}
		fnVal.Call(args)

		if fBuildTarget.IsValid() && fBuildTarget.CanSet() {
			fBuildTarget.Set(reflect.Zero(fBuildTarget.Type()))
		}

		mValidateAny := nestedSchemaVal.MethodByName("ValidateAny")
		if !mValidateAny.IsValid() {
			return nil, fmt.Errorf("ValidateAny not found on schema")
		}

		res := mValidateAny.Call([]reflect.Value{
			reflect.ValueOf(val),
			reflect.ValueOf(context.Options),
		})

		resVal := res[0].Interface()
		resErr := res[1].Interface()

		if resErr != nil {
			errCompat, _ := resErr.(error)

			if vErr, ok := errCompat.(*issues.ValidationError); ok {
				basePath := context.PathString()
				for _, issue := range vErr.Issues {
					issue.Path = joinIssuePath(basePath, issue.Path)
					context.Issues.Add(issue)
				}
				return nil, nil
			}

			return nil, errCompat
		}

		return resVal, nil
	}

	compiled, err := newFieldFromInfo(b.fieldInfo, validator)
	if err != nil {
		b.schema.buildError = err
		return b
	}
	compiled.required = b.required

	if b.fieldIndex == -1 {
		b.schema.fields = append(b.schema.fields, compiled)
		b.schema.lastFieldIndex = len(b.schema.fields) - 1
		b.fieldIndex = b.schema.lastFieldIndex
	} else {
		b.schema.fields[b.fieldIndex] = compiled
		b.schema.lastFieldIndex = b.fieldIndex
	}

	return b
}

// joinIssuePath joins a base path with an issue's relative path.
func joinIssuePath(basePath, issuePath string) string {
	if basePath == "" {
		return issuePath
	}
	if issuePath == "" {
		return basePath
	}
	if issuePath[0] == '[' {
		return basePath + issuePath
	}
	return basePath + "." + issuePath
}
