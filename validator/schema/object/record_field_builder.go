package object

import (
	"reflect"

	"github.com/leandroluk/go/validator/internal/ast"
	"github.com/leandroluk/go/validator/internal/engine"
	"github.com/leandroluk/go/validator/internal/issues"
	"github.com/leandroluk/go/validator/schema"
)

type RecordFieldBuilder[T any] struct {
	schema    *Schema[T]
	fieldInfo fieldInfo[T]

	required bool
	min      *int
	max      *int
	len      *int

	keySchema   schema.AnySchema
	valueSchema schema.AnySchema

	fieldIndex int
}

func (b *RecordFieldBuilder[T]) Required() *RecordFieldBuilder[T] {
	b.required = true
	return b.build()
}

func (b *RecordFieldBuilder[T]) Min(min int) *RecordFieldBuilder[T] {
	b.min = &min
	return b.build()
}

func (b *RecordFieldBuilder[T]) Max(max int) *RecordFieldBuilder[T] {
	b.max = &max
	return b.build()
}

func (b *RecordFieldBuilder[T]) Len(len int) *RecordFieldBuilder[T] {
	b.len = &len
	return b.build()
}

func (b *RecordFieldBuilder[T]) build() *RecordFieldBuilder[T] {
	mapType := b.fieldInfo.fieldType
	if mapType.Kind() != reflect.Map {
		return b
	}

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

		if val.Kind != ast.KindObject {
			context.AddIssue("record.type", "expected object/map")
			return nil, nil
		}

		obj := val.Object // map[string]Value

		count := len(obj)

		if b.min != nil && count < *b.min {
			context.AddIssueWithMeta("record.min", "too few items", map[string]any{"min": *b.min, "actual": count})
			return nil, nil
		}
		if b.max != nil && count > *b.max {
			context.AddIssueWithMeta("record.max", "too many items", map[string]any{"max": *b.max, "actual": count})
			return nil, nil
		}
		if b.len != nil && count != *b.len {
			context.AddIssueWithMeta("record.len", "invalid length", map[string]any{"len": *b.len, "actual": count})
			return nil, nil
		}

		resultMap := reflect.MakeMapWithSize(mapType, count)
		basePath := context.PathString()

		for key, item := range obj {
			// Validate Key
			if b.keySchema != nil {
				keyVal := ast.StringValue(key)
				_, err := b.keySchema.ValidateAny(keyVal, context.Options)
				if err != nil {
					if vErr, ok := err.(issues.ValidationError); ok {
						for _, issue := range vErr.Issues {
							context.AddIssueWithMeta("record.key", "invalid key", map[string]any{"key": key, "details": issue.Message})
						}
					}
				}
			}

			// Validate Value
			if b.valueSchema != nil {
				itemRes, err := b.valueSchema.ValidateAny(item, context.Options)
				if err != nil {
					if vErr, ok := err.(issues.ValidationError); ok {
						for _, issue := range vErr.Issues {
							// Construct relative path for this item
							// If key is "myKey", path is "myKey" or "myKey.subPath"
							// We use bracket style if it helps, but JSON pointer uses /myKey/subPath
							// V uses standard dot notation usually.

							var itemRelPath string
							if issue.Path != "" {
								if issue.Path[0] == '[' {
									itemRelPath = key + issue.Path
								} else {
									itemRelPath = key + "." + issue.Path
								}
							} else {
								itemRelPath = key
							}

							var fullPath string
							if basePath != "" {
								fullPath = basePath + "." + itemRelPath
							} else {
								fullPath = itemRelPath
							}

							issue.Path = fullPath
							context.Issues.Add(issue)
						}
					} else {
						return nil, err
					}
				}
				if itemRes != nil {
					keyType := mapType.Key()
					var keyVal reflect.Value
					if keyType.Kind() == reflect.String {
						keyVal = reflect.ValueOf(key)
					} else {
						// Simplistic coercion fallback/skip
						continue
					}

					resultMap.SetMapIndex(keyVal, reflect.ValueOf(itemRes))
				}
			}
		}

		return resultMap.Interface(), nil
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

// Transform registers a transformation function for the field.
// It validates the value as a Record first, then applies the transformation.
// The returned value is used as the new value for the field.
func (b *RecordFieldBuilder[T]) Transform(fn func(value any) (any, error)) *Schema[T] {
	// Ensure standard validation is built and registered
	b.build()

	// Grab the registered field
	idx := b.fieldIndex
	if idx < 0 || idx >= len(b.schema.fields) {
		// Should not happen if build works
		return b.schema
	}

	currentField := b.schema.fields[idx]
	originalValidator := currentField.validate

	newValidator := func(ctx *engine.Context, value any) (any, error) {
		out, err := originalValidator(ctx, value)
		if err != nil {
			return nil, err
		}

		return fn(out)
	}

	// Update the field with new validator
	b.schema.fields[idx].validate = newValidator

	return b.schema
}
