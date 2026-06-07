package util

import (
	"encoding/json"
	"maps"

	"github.com/google/uuid"
)

// ID returns a new UUIDv7 as a string.
func ID() string {
	return Must(uuid.NewV7()).String()
}

// Ptr returns the value asserted as type T.
func Ptr[T any](v T) *T {
	return &v
}

// Must panics if err is non-nil, otherwise returns the value asserted as type T.
func Must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

// Try executes the function fn and catches any panic that implements the error interface,
// returning it as an error. If the panic value is not an error, it is re-panicked.
func Try(fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				panic(r)
			}
		}
	}()
	fn()
	return nil
}

// Check panics if err is non-nil.
func Check(err error) {
	if err != nil {
		panic(err)
	}
}

// MapMerge merges one or more maps into a new map. The last map in the slice wins in case of key collisions.
func MapMerge[K comparable, V any](base map[K]V, items ...map[K]V) map[K]V {
	result := make(map[K]V, len(base))
	maps.Copy(result, base)
	for _, m := range items {
		maps.Copy(result, m)
	}
	return result
}

// SetDefault returns value if it is not the zero value of type T, otherwise it returns defaultValue.
func SetDefault[T comparable](value T, defaultValue T) T {
	if value == *new(T) {
		return defaultValue
	}
	return value
}

// SetNil checks if the value pointed by the pointer is nil, sets it to defaultValue if it is, and returns the final value.
func SetNil[T any](value *T, defaultValue T) T {
	if any(*value) == nil {
		*value = defaultValue
	}
	return *value
}

// StructFromMap converts a map of strings to any into a struct of type T.
func StructFromMap[T any](input map[string]any) (T, error) {
	var result T

	bytes, err := json.Marshal(input)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(bytes, &result)
	return result, err
}
