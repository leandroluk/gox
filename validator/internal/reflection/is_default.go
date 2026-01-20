// internal/reflection/is_default.go
package reflection

import "reflect"

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
