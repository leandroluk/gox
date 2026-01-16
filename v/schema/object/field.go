// schema/object/field.go
package object

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"github.com/leandroluk/go/v/internal/ast"
	"github.com/leandroluk/go/v/internal/codec"
	"github.com/leandroluk/go/v/internal/engine"
	"github.com/leandroluk/go/v/schema/object/rule"
)

type field[T any] struct {
	name      string
	offset    uintptr
	fieldType reflect.Type

	skipUnlessConditions []rule.SkipUnlessCondition
	excludedConditions   []rule.ExcludedIfCondition
	requiredConditions   []rule.RequiredCondition

	comparators []rule.Comparator

	validate func(context *engine.Context, value any) (any, bool)
	assign   func(outputPointer unsafe.Pointer, value any)
}

func newField[T any](structPointer *T, fieldPointer any, validator func(context *engine.Context, value any) (any, bool)) (field[T], error) {
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

	validateFn := func(context *engine.Context, value any) (any, bool) {
		if validator != nil {
			return validator(context, value)
		}

		astValue, ok := value.(ast.Value)
		if ok {
			return decodeFallback(context, astValue, fieldType)
		}

		astValuePointer, ok := value.(*ast.Value)
		if ok {
			if astValuePointer == nil {
				return decodeFallback(context, ast.NullValue(), fieldType)
			}
			return decodeFallback(context, *astValuePointer, fieldType)
		}

		astValue, err := engine.InputToASTWithOptions(value, context.Options)
		if err != nil {
			stop := context.AddIssueWithMeta(CodeFieldDecode, "invalid field", map[string]any{
				"error": err.Error(),
			})
			return reflect.Zero(fieldType).Interface(), stop
		}
		return decodeFallback(context, astValue, fieldType)
	}

	assignFn := func(outputPointer unsafe.Pointer, value any) {
		if outputPointer == nil {
			return
		}

		target := reflect.NewAt(fieldType, unsafe.Pointer(uintptr(outputPointer)+matchedOffset)).Elem()

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

		target.Set(reflect.Zero(fieldType))
	}

	return field[T]{
		name:      name,
		offset:    matchedOffset,
		fieldType: fieldType,

		skipUnlessConditions: make([]rule.SkipUnlessCondition, 0),
		excludedConditions:   make([]rule.ExcludedIfCondition, 0),
		requiredConditions:   make([]rule.RequiredCondition, 0),
		comparators:          make([]rule.Comparator, 0),

		validate: validateFn,
		assign:   assignFn,
	}, nil
}

func jsonName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "-" {
		return ""
	}
	if tag == "" {
		return field.Name
	}

	name := tag
	if comma := strings.IndexByte(tag, ','); comma >= 0 {
		name = tag[:comma]
	}

	if name == "" {
		return field.Name
	}

	return name
}

func decodeFallback(context *engine.Context, value ast.Value, fieldType reflect.Type) (any, bool) {
	if value.IsMissing() || value.IsNull() {
		return reflect.Zero(fieldType).Interface(), false
	}

	outPointer := reflect.New(fieldType)
	if err := codec.DecodeInto(value, outPointer.Interface()); err != nil {
		stop := context.AddIssueWithMeta(CodeFieldDecode, "invalid field", map[string]any{
			"error": err.Error(),
		})
		return reflect.Zero(fieldType).Interface(), stop
	}

	return outPointer.Elem().Interface(), false
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
