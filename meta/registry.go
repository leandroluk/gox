package meta

import (
	"fmt"
	"reflect"
	"strings"
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

// ResolveFieldName compares pointers and types to find the string name of a struct field.
func ResolveFieldName(structPointer any, fieldPointer any) (name string, jsonTag string, fieldType reflect.Type) {
	structValue := reflect.ValueOf(structPointer)
	if structValue.Kind() == reflect.Pointer {
		structValue = structValue.Elem()
	}

	fVal := reflect.ValueOf(fieldPointer)
	if fVal.Kind() != reflect.Pointer {
		return "", "", nil
	}
	targetAddr := fVal.Pointer()
	targetType := fVal.Type().Elem()

	return resolveFieldNameRecursive(structValue, structValue.Type(), targetAddr, targetType, "")
}

func resolveFieldNameRecursive(
	v reflect.Value,
	t reflect.Type,
	targetAddr uintptr,
	targetType reflect.Type,
	jsonPrefix string,
) (name string, jsonTag string, fieldType reflect.Type) {
	if v.Kind() != reflect.Struct {
		return "", "", nil
	}

	for i := 0; i < v.NumField(); i++ {
		fVal := v.Field(i)
		fType := t.Field(i)

		rawJSON := fType.Tag.Get("json")
		jsonName := strings.Split(rawJSON, ",")[0]
		if jsonName == "" || jsonName == "-" {
			jsonName = fType.Name
		}
		fullJSON := jsonName
		if jsonPrefix != "" {
			fullJSON = jsonPrefix + "." + jsonName
		}

		if fVal.CanAddr() && fVal.Addr().Pointer() == targetAddr {
			if fVal.Type() == targetType {
				return fType.Name, fullJSON, fVal.Type()
			}
		}

		checkVal := fVal
		if checkVal.Kind() == reflect.Pointer && !checkVal.IsNil() {
			checkVal = checkVal.Elem()
		}

		if checkVal.Kind() == reflect.Struct {
			subName, subJSON, subType := resolveFieldNameRecursive(checkVal, checkVal.Type(), targetAddr, targetType, fullJSON)
			if subName != "" {
				if fType.Anonymous {
					return subName, subJSON, subType
				}
				return fType.Name + "." + subName, subJSON, subType
			}
		}
	}
	return "", "", nil
}
