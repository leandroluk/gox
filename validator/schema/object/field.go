// schema/object/field.go
package object

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"github.com/leandroluk/go/validator/internal/ast"
	"github.com/leandroluk/go/validator/internal/codec"
	"github.com/leandroluk/go/validator/internal/engine"
	"github.com/leandroluk/go/validator/schema/object/rule"
)

type field[T any] struct {
	name      string
	offset    uintptr
	fieldType reflect.Type

	skipUnlessConditions []rule.SkipUnlessCondition
	excludedConditions   []rule.ExcludedIfCondition
	requiredConditions   []rule.RequiredCondition
	required             bool

	comparators []rule.Comparator

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

	assignFn := func(outputPointer unsafe.Pointer, value any) {
		if outputPointer == nil {
			return
		}

		target := reflect.NewAt(info.fieldType, unsafe.Pointer(uintptr(outputPointer)+info.offset)).Elem()

		if value == nil {
			target.Set(reflect.Zero(info.fieldType))
			return
		}

		v := reflect.ValueOf(value)
		if !v.IsValid() {
			target.Set(reflect.Zero(info.fieldType))
			return
		}

		if v.Type().AssignableTo(info.fieldType) {
			target.Set(v)
			return
		}

		if v.Type().ConvertibleTo(info.fieldType) {
			target.Set(v.Convert(info.fieldType))
			return
		}

		if info.fieldType.Kind() == reflect.Pointer {
			elemType := info.fieldType.Elem()
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

		target.Set(reflect.Zero(info.fieldType))
	}

	return field[T]{
		name:      info.name,
		offset:    info.offset,
		fieldType: info.fieldType,

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
