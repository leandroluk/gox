package env

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// convertStringToType converts a string to the specified type T.
// Supports pointer types: Get[*string], Get[*int32], etc.
func convertStringToType[T any](raw string) (T, error) {
	var zero T
	targetType := reflect.TypeFor[T]()

	if targetType.Kind() == reflect.Ptr {
		elemType := targetType.Elem()
		elemVal, err := convertStringToValue(raw, elemType)
		if err != nil {
			return zero, err
		}
		ptr := reflect.New(elemType)
		ptr.Elem().Set(elemVal)
		result, ok := ptr.Interface().(T)
		if !ok {
			return zero, fmt.Errorf("env: cannot cast to %v", targetType)
		}
		return result, nil
	}

	v, err := convertStringToValue(raw, targetType)
	if err != nil {
		return zero, err
	}
	result, ok := v.Interface().(T)
	if !ok {
		return zero, fmt.Errorf("env: cannot cast to %v", targetType)
	}
	return result, nil
}

// convertStringToValue converts a string to a reflect.Value of type t.
func convertStringToValue(raw string, t reflect.Type) (reflect.Value, error) {
	switch t {
	case reflect.TypeFor[json.RawMessage]():
		if !json.Valid([]byte(raw)) {
			return reflect.Value{}, fmt.Errorf("env: invalid JSON")
		}
		return reflect.ValueOf(json.RawMessage(raw)), nil
	case reflect.TypeFor[time.Time]():
		v, err := parseTime(raw)
		return reflect.ValueOf(v), err
	case reflect.TypeFor[time.Duration]():
		v, err := time.ParseDuration(raw)
		return reflect.ValueOf(v), err
	}

	switch t.Kind() {
	case reflect.String:
		return reflect.ValueOf(raw).Convert(t), nil
	case reflect.Bool:
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(v).Convert(t), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(raw, 10, t.Bits())
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(v).Convert(t), nil
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(raw, t.Bits())
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(v).Convert(t), nil
	}

	return reflect.Value{}, fmt.Errorf("env: unsupported type %v", t)
}

// parseTime parses a time string.
func parseTime(raw string) (time.Time, error) {
	layouts := []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"}
	for _, l := range layouts {
		if t, err := time.Parse(l, raw); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("env: invalid time format %q", raw)
}
