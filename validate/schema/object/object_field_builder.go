package object

import (
	"fmt"
	"reflect"

	"github.com/leandroluk/gox/validate/internal/ast"
	"github.com/leandroluk/gox/validate/internal/engine"
	"github.com/leandroluk/gox/validate/internal/issues"
)

type ObjectFieldBuilder[T any] struct {
	schema    *Schema[T]
	fieldInfo fieldInfo[T]

	required      bool
	structOnly    bool
	noStructLevel bool
	builderFunc   any

	fieldIndex int
}

func (b *ObjectFieldBuilder[T]) Required() *ObjectFieldBuilder[T] {
	b.required = true
	return b.build()
}

func (b *ObjectFieldBuilder[T]) StructOnly() *ObjectFieldBuilder[T] {
	b.structOnly = true
	b.noStructLevel = false
	return b.build()
}

func (b *ObjectFieldBuilder[T]) NoStructLevel() *ObjectFieldBuilder[T] {
	b.noStructLevel = true
	b.structOnly = false
	return b.build()
}

// build updates the schema with the current configuration.
func (b *ObjectFieldBuilder[T]) build() *ObjectFieldBuilder[T] {
	targetType := b.fieldInfo.fieldType

	validator := func(context *engine.Context, value any) (any, error) {
		val, ok := value.(ast.Value)
		if !ok {
			return nil, nil
		}

		if val.IsMissing() {
			if b.required {
				context.AddIssue("object.required", "required")
				return nil, nil
			}
			return nil, nil
		}
		if val.IsNull() {
			return nil, nil
		}

		if b.builderFunc == nil {
			// No validation if no builder provided? Or assume type check?
			// Let's assume just type check passed implicitly by conversion.
			return nil, nil
		}

		fnVal := reflect.ValueOf(b.builderFunc)
		// Expected: func(*N, *Schema[N])

		// Create *N
		nestedType := targetType
		if nestedType.Kind() == reflect.Ptr {
			nestedType = nestedType.Elem()
		}
		nestedPtr := reflect.New(nestedType)

		// Inspect builder func to find Schema type
		funcType := fnVal.Type()
		if funcType.NumIn() != 2 {
			return nil, fmt.Errorf("invalid builder signature: expected 2 arguments")
		}

		schemaTypePtr := funcType.In(1)    // *Schema[N]
		schemaType := schemaTypePtr.Elem() // Schema[N]

		// Instantiate Schema[N]
		nestedSchemaVal := reflect.New(schemaType) // *Schema[N]

		// Initialize fields of Schema[N] manually via reflection
		// defaults: lastFieldIndex = -1
		fLastFieldIndex := nestedSchemaVal.Elem().FieldByName("lastFieldIndex")
		if fLastFieldIndex.IsValid() && fLastFieldIndex.CanSet() {
			fLastFieldIndex.SetInt(-1)
		}

		// Set buildTarget
		fBuildTarget := nestedSchemaVal.Elem().FieldByName("buildTarget")
		if fBuildTarget.IsValid() && fBuildTarget.CanSet() {
			fBuildTarget.Set(nestedPtr)
		}

		// Initialize slices if they are nil?
		// Go handles appending to nil slices fine. No action needed.

		// Call builder: func(target, schema)
		args := []reflect.Value{nestedPtr, nestedSchemaVal}
		fnVal.Call(args)

		// Clean up buildTarget
		if fBuildTarget.IsValid() && fBuildTarget.CanSet() {
			fBuildTarget.Set(reflect.Zero(fBuildTarget.Type()))
		}

		// Call ValidateAny
		mValidateAny := nestedSchemaVal.MethodByName("ValidateAny")
		if !mValidateAny.IsValid() {
			return nil, fmt.Errorf("ValidateAny not found on schema")
		}

		res := mValidateAny.Call([]reflect.Value{
			reflect.ValueOf(val),             // input AST
			reflect.ValueOf(context.Options), // options
		})

		resVal := res[0].Interface()
		resErr := res[1].Interface()

		if resErr != nil {
			errCompat, _ := resErr.(error)

			if vErr, ok := errCompat.(issues.ValidationError); ok {
				// Rebase paths
				basePath := context.PathString()
				for _, issue := range vErr.Issues {
					// Logic: base + issue.Path (dot separator unless bracket)
					// If base is empty, just issue.Path.

					var fullPath string
					if basePath == "" {
						fullPath = issue.Path
					} else {
						if issue.Path != "" {
							if issue.Path[0] == '[' {
								fullPath = basePath + issue.Path
							} else {
								fullPath = basePath + "." + issue.Path
							}
						} else {
							fullPath = basePath
						}
					}

					issue.Path = fullPath
					context.Issues.Add(issue)
				}
				return nil, nil // issues added
			}

			// Non-validation error (e.g. internal)
			return nil, errCompat
		}

		return resVal, nil
	}

	compiled, err := newFieldFromInfo(b.fieldInfo, validator)
	if err != nil {
		b.schema.buildError = err
		return b
	}
	compiled.required = b.required

	if b.fieldIndex == -1 {
		b.schema.fields = append(b.schema.fields, compiled)
		b.schema.lastFieldIndex = len(b.schema.fields) - 1
		b.fieldIndex = b.schema.lastFieldIndex
	} else {
		b.schema.fields[b.fieldIndex] = compiled
		b.schema.lastFieldIndex = b.fieldIndex
	}

	return b
}
