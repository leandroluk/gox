package meta

import (
	"fmt"
	"reflect"
	"sync"
)

var (
	registryMutex  sync.RWMutex
	structRegistry = make(map[reflect.Type]*ObjectMetadata)
)

// GetObjectMetadataAs retrieves metadata for the type T using Generics.
func GetObjectMetadataAs[T any]() *ObjectMetadata {
	structType := reflect.TypeFor[T]()
	if structType.Kind() == reflect.Pointer {
		structType = structType.Elem()
	}

	registryMutex.RLock()
	defer registryMutex.RUnlock()

	return structRegistry[structType]
}

// GetObjectMetadataOf retrieves metadata based on the instance's type.
func GetObjectMetadataOf(structInstance any) *ObjectMetadata {
	if structInstance == nil {
		return nil
	}
	structType := reflect.TypeOf(structInstance)
	if structType.Kind() == reflect.Pointer {
		structType = structType.Elem()
	}
	if structType.Kind() != reflect.Struct {
		return nil
	}
	registryMutex.RLock()
	defer registryMutex.RUnlock()
	return structRegistry[structType]
}

// GetObjectMetadataByType retrieves metadata for a specific reflect.Type.
func GetObjectMetadataByType(structType reflect.Type) *ObjectMetadata {
	if structType == nil {
		return nil
	}
	if structType.Kind() == reflect.Pointer {
		structType = structType.Elem()
	}
	registryMutex.RLock()
	defer registryMutex.RUnlock()
	return structRegistry[structType]
}

// Describe initializes or updates metadata for a struct pointer.
func Describe(target any, options ...ObjectOption) {
	if target == nil {
		panic("meta: target is nil in Describe")
	}

	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Pointer || targetValue.Elem().Kind() != reflect.Struct {
		panic(fmt.Sprintf("meta: Describe target must be pointer to struct, got %T", target))
	}

	structType := targetValue.Elem().Type()

	registryMutex.Lock()
	metadata, exists := structRegistry[structType]
	if !exists {
		metadata = &ObjectMetadata{
			Fields: make(map[string]*FieldMetadata),
			Type:   structType,
		}
		structRegistry[structType] = metadata
	}
	registryMutex.Unlock()

	for _, option := range options {
		if option != nil {
			option.applyToObject(target, metadata)
		}
	}
}

// resolveFieldName compares pointers and types to find the string name of a struct field.
func resolveFieldName(structPointer any, fieldPointer any) string {
	structValue := reflect.ValueOf(structPointer)
	if structValue.Kind() == reflect.Pointer {
		structValue = structValue.Elem()
	}

	fVal := reflect.ValueOf(fieldPointer)
	if fVal.Kind() != reflect.Pointer {
		return ""
	}
	targetAddr := fVal.Pointer()

	// We get the specific element type being pointed to (e.g., string, int, Struct)
	// This allows us to distinguish between &MyStruct and &MyStruct.FirstField
	// which share the same memory address but have different types.
	targetType := fVal.Type().Elem()

	return resolveFieldNameRecursive(structValue, targetAddr, targetType)
}

func resolveFieldNameRecursive(v reflect.Value, targetAddr uintptr, targetType reflect.Type) string {
	if v.Kind() != reflect.Struct {
		return ""
	}
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fVal := v.Field(i)
		fType := t.Field(i)

		// 1. Check if the address matches
		if fVal.CanAddr() && fVal.Addr().Pointer() == targetAddr {
			// 2. Strict Type Check:
			// If the address is the same, we must also check if the type is the same.
			// This handles the edge case where a Struct and its first Field share the same address.
			if fVal.Type() == targetType {
				return fType.Name
			}
		}

		// 3. Recursion
		checkVal := fVal
		if checkVal.Kind() == reflect.Pointer && !checkVal.IsNil() {
			checkVal = checkVal.Elem()
		}

		if checkVal.Kind() == reflect.Struct {
			if sub := resolveFieldNameRecursive(checkVal, targetAddr, targetType); sub != "" {
				if fType.Anonymous {
					return sub
				}
				return fType.Name + "." + sub
			}
		}
	}
	return ""
}
