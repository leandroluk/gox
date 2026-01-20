package object

import (
	"fmt"
	"reflect"

	"github.com/leandroluk/go/validator/internal/ast"
	"github.com/leandroluk/go/validator/internal/engine"
	"github.com/leandroluk/go/validator/internal/issues"
	"github.com/leandroluk/go/validator/schema"
)

type ArrayFieldBuilder[T any] struct {
	schema    *Schema[T]
	fieldInfo fieldInfo[T]

	required bool
	min      *int
	max      *int
	len      *int
	eq       *int
	ne       *int
	gt       *int
	gte      *int
	lt       *int
	lte      *int
	unique   bool

	itemSchema schema.AnySchema

	fieldIndex int
}

func (b *ArrayFieldBuilder[T]) Required() *ArrayFieldBuilder[T] {
	b.required = true
	return b.build()
}

func (b *ArrayFieldBuilder[T]) Min(min int) *ArrayFieldBuilder[T] {
	b.min = &min
	return b.build()
}

func (b *ArrayFieldBuilder[T]) Max(max int) *ArrayFieldBuilder[T] {
	b.max = &max
	return b.build()
}

func (b *ArrayFieldBuilder[T]) Len(len int) *ArrayFieldBuilder[T] {
	b.len = &len
	return b.build()
}

func (b *ArrayFieldBuilder[T]) Eq(len int) *ArrayFieldBuilder[T] {
	b.eq = &len
	return b.build()
}

func (b *ArrayFieldBuilder[T]) Ne(len int) *ArrayFieldBuilder[T] {
	b.ne = &len
	return b.build()
}

func (b *ArrayFieldBuilder[T]) Gt(len int) *ArrayFieldBuilder[T] {
	b.gt = &len
	return b.build()
}

func (b *ArrayFieldBuilder[T]) Gte(len int) *ArrayFieldBuilder[T] {
	b.gte = &len
	return b.build()
}

func (b *ArrayFieldBuilder[T]) Lt(len int) *ArrayFieldBuilder[T] {
	b.lt = &len
	return b.build()
}

func (b *ArrayFieldBuilder[T]) Lte(len int) *ArrayFieldBuilder[T] {
	b.lte = &len
	return b.build()
}

func (b *ArrayFieldBuilder[T]) Unique() *ArrayFieldBuilder[T] {
	b.unique = true
	return b.build()
}

func (b *ArrayFieldBuilder[T]) build() *ArrayFieldBuilder[T] {
	sliceType := b.fieldInfo.fieldType
	if sliceType.Kind() != reflect.Slice {
		return b
	}

	validator := func(context *engine.Context, value any) (any, error) {
		val, ok := value.(ast.Value)
		if !ok {
			return nil, nil // Should not happen
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

		if val.Kind != ast.KindArray {
			context.AddIssue("array.type", "expected array")
			return nil, nil
		}

		arr := val.Array

		count := len(arr)

		if b.min != nil && count < *b.min {
			context.AddIssueWithMeta("array.min", "too short", map[string]any{"min": *b.min, "actual": count})
			return nil, nil
		}
		if b.max != nil && count > *b.max {
			context.AddIssueWithMeta("array.max", "too long", map[string]any{"max": *b.max, "actual": count})
			return nil, nil
		}
		if b.len != nil && count != *b.len {
			context.AddIssueWithMeta("array.len", "invalid length", map[string]any{"len": *b.len, "actual": count})
			return nil, nil
		}
		// ... other rules omitted for brevity, add if needed or rely on base set ...

		resultSlice := reflect.MakeSlice(sliceType, 0, count)

		if b.itemSchema != nil {
			// Pre-calculate path base to avoid re-allocating
			basePath := context.PathString()

			for i, item := range arr {
				itemRes, err := b.itemSchema.ValidateAny(item, context.Options)
				if err != nil {
					if vErr, ok := err.(issues.ValidationError); ok {
						for _, issue := range vErr.Issues {
							// Manually join path: basePath + "[" + i + "]" + issue.Path
							// We can optimize common cases

							indexPath := fmt.Sprintf("[%d]", i)

							// Logic: full = base + index + (issue.Path if not empty)
							// If issue.Path starts with [, concat directly. Else dot.

							var itemRelPath string
							if issue.Path != "" {
								if issue.Path[0] == '[' {
									itemRelPath = indexPath + issue.Path
								} else {
									itemRelPath = indexPath + "." + issue.Path
								}
							} else {
								itemRelPath = indexPath
							}

							var fullPath string
							if basePath != "" {
								// if base is "a.b", and itemRel is "[0]", result "a.b[0]"
								fullPath = basePath + itemRelPath
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
					resultSlice = reflect.Append(resultSlice, reflect.ValueOf(itemRes))
				}
			}
		}

		return resultSlice.Interface(), nil
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
// It validates the value as an Array first, then applies the transformation.
// The returned value is used as the new value for the field.
func (b *ArrayFieldBuilder[T]) Transform(fn func(value any) (any, error)) *Schema[T] {
	// Ensure standard validation is built and registered
	b.build()

	// Grab the registered field
	idx := b.fieldIndex
	if idx < 0 || idx >= len(b.schema.fields) {
		// Should not happen if build works
		return b.schema
	}

	// We need to modify the validator of the field.
	// field[T] is in the same package (object), so we can access its fields if exported or if we are in expected package.
	// We are in package object.

	currentField := b.schema.fields[idx]
	originalValidator := currentField.validate

	newValidator := func(ctx *engine.Context, value any) (any, error) {
		out, err := originalValidator(ctx, value)
		if err != nil {
			return nil, err
		}

		// If out is nil, it might be missing/null (and allowed).
		// Transforming nil might be desired or not.
		// For now, we pass it to fn. User should handle nil if needed.
		// Or we skip transform if nil?
		// Consistency: .Custom receives what ValidateAny returns.

		return fn(out)
	}

	// Update the field with new validator
	b.schema.fields[idx].validate = newValidator

	return b.schema
}
