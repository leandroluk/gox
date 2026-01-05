package cqrs

import (
	"fmt"
	"reflect"
)

// normalizeType removes the pointer from a type if it exists.
func normalizeType(targetType reflect.Type) reflect.Type {
	if targetType == nil {
		return nil
	}
	if targetType.Kind() == reflect.Pointer {
		return targetType.Elem()
	}
	return targetType
}

// normalizedTypeKeyOfValue extracts the reflect.Type of a value and normalizes it.
func normalizedTypeKeyOfValue(value any, valueName string) (reflect.Type, error) {
	if value == nil {
		return nil, fmt.Errorf("cqrs: nil %s", valueName)
	}

	valueType := reflect.TypeOf(value)

	if valueType.Kind() == reflect.Pointer && reflect.ValueOf(value).IsNil() {
		return nil, fmt.Errorf("cqrs: nil %s pointer", valueName)
	}

	return normalizeType(valueType), nil
}

// coerce attempts to convert a value to TExpected, handling pointer/value conversions.
func coerce[TExpected any](value any, valueName string) (TExpected, error) {
	var zero TExpected

	if value == nil {
		return zero, fmt.Errorf("cqrs: nil %s", valueName)
	}

	expectedType := reflect.TypeFor[TExpected]()
	gotValue := reflect.ValueOf(value)
	gotType := gotValue.Type()

	if gotType.AssignableTo(expectedType) {
		return gotValue.Interface().(TExpected), nil
	}

	// Handle Pointer to Value conversion
	if expectedType.Kind() != reflect.Pointer && gotType.Kind() == reflect.Pointer {
		if gotValue.IsNil() {
			return zero, fmt.Errorf("cqrs: nil %s pointer", valueName)
		}
		elem := gotValue.Elem()
		if elem.Type().AssignableTo(expectedType) {
			return elem.Interface().(TExpected), nil
		}
	}

	// Handle Value to Pointer conversion
	if expectedType.Kind() == reflect.Pointer && gotType.Kind() != reflect.Pointer {
		expectedElem := expectedType.Elem()
		if gotType.AssignableTo(expectedElem) {
			pointer := reflect.New(expectedElem)
			pointer.Elem().Set(gotValue)
			return pointer.Interface().(TExpected), nil
		}
	}

	return zero, fmt.Errorf("cqrs: expected %s compatible with %v, got %T", valueName, expectedType, value)
}
