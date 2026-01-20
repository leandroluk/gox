// internal/reflection/field_offset.go
package reflection

import (
	"fmt"
	"reflect"
	"unsafe"
)

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
