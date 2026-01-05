package env

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

func convertStringToType[T any](raw string) (T, error) {
	var zero T
	targetType := reflect.TypeFor[T]()

	// Special Types
	switch targetType {
	case reflect.TypeFor[json.RawMessage]():
		if !json.Valid([]byte(raw)) {
			return zero, fmt.Errorf("env: invalid JSON")
		}
		return any(json.RawMessage(raw)).(T), nil
	case reflect.TypeFor[time.Time]():
		t, err := parseTime(raw)
		return any(t).(T), err
	case reflect.TypeFor[time.Duration]():
		d, err := time.ParseDuration(raw)
		return any(d).(T), err
	}

	// Basic Kinds
	switch targetType.Kind() {
	case reflect.String:
		return any(raw).(T), nil
	case reflect.Bool:
		v, err := strconv.ParseBool(raw)
		return any(v).(T), err
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(raw, 10, targetType.Bits())
		return reflect.ValueOf(v).Convert(targetType).Interface().(T), err
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(raw, targetType.Bits())
		return reflect.ValueOf(v).Convert(targetType).Interface().(T), err
	}

	return zero, fmt.Errorf("env: unsupported type %v", targetType)
}

func parseTime(raw string) (time.Time, error) {
	layouts := []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"}
	for _, l := range layouts {
		if t, err := time.Parse(l, raw); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("env: invalid time format %q", raw)
}
