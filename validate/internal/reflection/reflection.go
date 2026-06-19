// internal/reflection/reflection.go
package reflection

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

func FieldJSONName(field reflect.StructField) string {
	tag := ParseJSONTag(field.Tag.Get("json"))
	if tag.Ignored {
		return ""
	}
	if tag.HasTag && tag.Name != "" {
		return tag.Name
	}
	return field.Name
}

func IsExported(field reflect.StructField) bool {
	return field.PkgPath == ""
}

type ResolvedField struct {
	Field  reflect.StructField
	Offset uintptr
	Name   string
}

func ResolveField(structPointer any, fieldPointer any) (ResolvedField, bool, error) {
	if structPointer == nil {
		return ResolvedField{}, false, fmt.Errorf("struct pointer is nil")
	}
	if fieldPointer == nil {
		return ResolvedField{}, false, fmt.Errorf("field pointer is nil")
	}

	structValue := reflect.ValueOf(structPointer)
	if structValue.Kind() != reflect.Pointer || structValue.IsNil() {
		return ResolvedField{}, false, fmt.Errorf("structPointer must be a non-nil pointer to struct")
	}

	structType := structValue.Type().Elem()
	if structType.Kind() != reflect.Struct {
		return ResolvedField{}, false, fmt.Errorf("structPointer must point to a struct, got %s", structType.Kind().String())
	}

	fieldValue := reflect.ValueOf(fieldPointer)
	if fieldValue.Kind() != reflect.Pointer || fieldValue.IsNil() {
		return ResolvedField{}, false, fmt.Errorf("fieldPointer must be a non-nil pointer")
	}

	structAddress := unsafe.Pointer(structValue.Pointer())
	fieldAddress := unsafe.Pointer(fieldValue.Pointer())

	offset := uintptr(fieldAddress) - uintptr(structAddress)

	for index := 0; index < structType.NumField(); index++ {
		field := structType.Field(index)
		if field.Offset != offset {
			continue
		}

		name := FieldJSONName(field)
		return ResolvedField{
			Field:  field,
			Offset: offset,
			Name:   name,
		}, true, nil
	}

	return ResolvedField{}, false, nil
}

func IsDefault(value any) bool {
	if value == nil {
		return true
	}
	return IsDefaultValue(reflect.ValueOf(value))
}

func IsDefaultValue(value reflect.Value) bool {
	if !value.IsValid() {
		return true
	}

	for value.Kind() == reflect.Interface {
		if value.IsNil() {
			return true
		}
		value = value.Elem()
	}

	return value.IsZero()
}

func UnwrapInterface(value reflect.Value) (reflect.Value, bool) {
	if !value.IsValid() {
		return value, true
	}

	for value.Kind() == reflect.Interface {
		if value.IsNil() {
			return value, true
		}
		value = value.Elem()
	}

	return value, false
}

func IsNilLike(value reflect.Value) bool {
	unwrapped, nilLike := UnwrapInterface(value)
	if nilLike {
		return true
	}

	switch unwrapped.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func, reflect.Interface:
		return unwrapped.IsNil()
	default:
		return false
	}
}

func IsLengthZero(value reflect.Value) bool {
	unwrapped, nilLike := UnwrapInterface(value)
	if nilLike {
		return true
	}

	switch unwrapped.Kind() {
	case reflect.String, reflect.Slice, reflect.Array, reflect.Map:
		return unwrapped.Len() == 0
	default:
		return false
	}
}

func IsEmpty(value reflect.Value) bool {
	unwrapped, nilLike := UnwrapInterface(value)
	if nilLike {
		return true
	}

	if IsNilLike(unwrapped) {
		return true
	}

	if unwrapped.IsZero() {
		return true
	}

	return IsLengthZero(unwrapped)
}

type JSONTag struct {
	Name      string
	OmitEmpty bool
	Ignored   bool
	HasTag    bool
}

func ParseJSONTag(tag string) JSONTag {
	if tag == "" {
		return JSONTag{}
	}

	head, tail, _ := strings.Cut(tag, ",")
	if head == "-" {
		return JSONTag{Ignored: true, HasTag: true}
	}

	result := JSONTag{
		Name:   head,
		HasTag: true,
	}

	for tail != "" {
		var part string
		part, tail, _ = strings.Cut(tail, ",")
		if part == "omitempty" {
			result.OmitEmpty = true
		}
	}

	return result
}
