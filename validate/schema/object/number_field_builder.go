// schema/object/numberfieldbuilder.go
package object

import (
	"reflect"
	"strconv"

	"github.com/leandroluk/gox/validate/internal/ast"
	"github.com/leandroluk/gox/validate/internal/engine"
	"github.com/leandroluk/gox/validate/schema"
)

type NumberFieldBuilder[T any] struct {
	schema     *Schema[T]
	fieldInfo  fieldInfo[T]
	required   bool
	isDefault  bool
	minSet     bool
	maxSet     bool
	minValue   float64
	maxValue   float64
	fieldIndex int
}

func (b *NumberFieldBuilder[T]) build() {
	if b.schema.buildError != nil {
		return
	}

	fieldType := b.fieldInfo.fieldType
	minSet := b.minSet
	maxSet := b.maxSet
	minValue := b.minValue
	maxValue := b.maxValue
	required := b.required

	validator := func(ctx *engine.Context, value any) (any, error) {
		astVal, ok := value.(ast.Value)
		if !ok {
			coerced, err := engine.InputToASTWithOptions(value, ctx.Options)
			if err != nil {
				return nil, err
			}
			astVal = coerced
		}

		if astVal.IsMissing() || astVal.IsNull() {
			if required {
				ctx.AddIssue("number.required", "required")
				return nil, ctx.Error()
			}
			return reflect.Zero(fieldType).Interface(), nil
		}

		if astVal.Kind != ast.KindNumber {
			ctx.AddIssueWithMeta("number.type", "expected number", map[string]any{
				"expected": "number",
				"actual":   astVal.Kind.String(),
			})
			return nil, ctx.Error()
		}

		floatVal, err := strconv.ParseFloat(astVal.Number, 64)
		if err != nil {
			ctx.AddIssueWithMeta("number.parse", "invalid number", map[string]any{"error": err.Error()})
			return nil, ctx.Error()
		}

		if minSet && floatVal < minValue {
			ctx.AddIssueWithMeta("number.min", "too small", map[string]any{"min": minValue, "actual": floatVal})
			return nil, ctx.Error()
		}

		if maxSet && floatVal > maxValue {
			ctx.AddIssueWithMeta("number.max", "too large", map[string]any{"max": maxValue, "actual": floatVal})
			return nil, ctx.Error()
		}

		result := convertToFieldType(floatVal, fieldType)
		return result, nil
	}

	compiled, err := newFieldFromInfo(b.fieldInfo, validator)
	if err != nil {
		b.schema.buildError = err
		return
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
}

func (b *NumberFieldBuilder[T]) Required() *NumberFieldBuilder[T] {
	b.required = true
	b.build()
	return b
}

func (b *NumberFieldBuilder[T]) IsDefault() *NumberFieldBuilder[T] {
	b.isDefault = true
	b.build()
	return b
}

func (b *NumberFieldBuilder[T]) Min(value float64) *NumberFieldBuilder[T] {
	b.minSet = true
	b.minValue = value
	b.build()
	return b
}

func (b *NumberFieldBuilder[T]) Max(value float64) *NumberFieldBuilder[T] {
	b.maxSet = true
	b.maxValue = value
	b.build()
	return b
}

func (b *NumberFieldBuilder[T]) Integer() *NumberFieldBuilder[T] {
	b.build()
	return b
}

func (b *NumberFieldBuilder[T]) RequiredIf(path string, op ConditionOp, expected any) *NumberFieldBuilder[T] {
	b.build()
	b.schema.RequiredIf(path, op, expected)
	return b
}

func (b *NumberFieldBuilder[T]) RequiredWith(paths ...string) *NumberFieldBuilder[T] {
	b.build()
	b.schema.RequiredWith(paths...)
	return b
}

func (b *NumberFieldBuilder[T]) RequiredWithout(paths ...string) *NumberFieldBuilder[T] {
	b.build()
	b.schema.RequiredWithout(paths...)
	return b
}

func (b *NumberFieldBuilder[T]) ExcludedIf(path string, op ConditionOp, expected any) *NumberFieldBuilder[T] {
	b.build()
	b.schema.ExcludedIf(path, op, expected)
	return b
}

func (b *NumberFieldBuilder[T]) SkipUnless(path string, op ConditionOp, expected any) *NumberFieldBuilder[T] {
	b.build()
	b.schema.SkipUnless(path, op, expected)
	return b
}

func (b *NumberFieldBuilder[T]) EqField(other string) *NumberFieldBuilder[T] {
	b.build()
	b.schema.EqField(other)
	return b
}

func (b *NumberFieldBuilder[T]) NeField(other string) *NumberFieldBuilder[T] {
	b.build()
	b.schema.NeField(other)
	return b
}

func (b *NumberFieldBuilder[T]) GtField(other string) *NumberFieldBuilder[T] {
	b.build()
	b.schema.GtField(other)
	return b
}

func (b *NumberFieldBuilder[T]) GteField(other string) *NumberFieldBuilder[T] {
	b.build()
	b.schema.GteField(other)
	return b
}

func (b *NumberFieldBuilder[T]) LtField(other string) *NumberFieldBuilder[T] {
	b.build()
	b.schema.LtField(other)
	return b
}

func (b *NumberFieldBuilder[T]) LteField(other string) *NumberFieldBuilder[T] {
	b.build()
	b.schema.LteField(other)
	return b
}

func (b *NumberFieldBuilder[T]) EqCSField(path string) *NumberFieldBuilder[T] {
	b.build()
	b.schema.EqCSField(path)
	return b
}

func (b *NumberFieldBuilder[T]) NeCSField(path string) *NumberFieldBuilder[T] {
	b.build()
	b.schema.NeCSField(path)
	return b
}

func (b *NumberFieldBuilder[T]) GtCSField(path string) *NumberFieldBuilder[T] {
	b.build()
	b.schema.GtCSField(path)
	return b
}

func (b *NumberFieldBuilder[T]) GteCSField(path string) *NumberFieldBuilder[T] {
	b.build()
	b.schema.GteCSField(path)
	return b
}

func (b *NumberFieldBuilder[T]) LtCSField(path string) *NumberFieldBuilder[T] {
	b.build()
	b.schema.LtCSField(path)
	return b
}

func (b *NumberFieldBuilder[T]) LteCSField(path string) *NumberFieldBuilder[T] {
	b.build()
	b.schema.LteCSField(path)
	return b
}

func (b *NumberFieldBuilder[T]) ValidateAny(value any, options schema.Options) (any, error) {
	return nil, nil
}

func (b *NumberFieldBuilder[T]) Build() *Schema[T] {
	b.build()
	return b.schema
}

// Transform registers a transformation function for the field.
// It validates the value as a Number first, then applies the transformation.
// The returned value is used as the new value for the field.
func (b *NumberFieldBuilder[T]) Transform(fn func(value any) (any, error)) *Schema[T] {
	// Replicates build() logic but wraps final generation.
	// Since build() is complex for NumberFieldBuilder (coercion, ranges, etc.), we should REUSE the logic if possible.
	// But build() logic is embedded in the closure `validator`.

	// We can refactor build() to return the validator closure?
	// Or we just duplicate the logic here for SAFETY and ensure it follows the same steps.
	// Duplication is risky for maintenance.
	// Better: Extract the validation logic into a private method `createValidator()`?

	// However, extracting now might be too much refactoring.
	// Let's implement it by creating a new validator that replicates the checks.

	fieldType := b.fieldInfo.fieldType
	minSet := b.minSet
	maxSet := b.maxSet
	minValue := b.minValue
	maxValue := b.maxValue
	required := b.required

	validator := func(ctx *engine.Context, value any) (any, error) {
		astVal, ok := value.(ast.Value)
		if !ok {
			coerced, err := engine.InputToASTWithOptions(value, ctx.Options)
			if err != nil {
				return nil, err
			}
			astVal = coerced
		}

		if astVal.IsMissing() || astVal.IsNull() {
			if required {
				ctx.AddIssue("number.required", "required")
				return nil, ctx.Error()
			}
			// If missing/null and optional, do we transform?
			// Usually we don't. We return zero value.
			// If user wants to transform zero value, they can use Defaults.
			return reflect.Zero(fieldType).Interface(), nil
		}

		if astVal.Kind != ast.KindNumber {
			ctx.AddIssueWithMeta("number.type", "expected number", map[string]any{
				"expected": "number",
				"actual":   astVal.Kind.String(),
			})
			return nil, ctx.Error()
		}

		floatVal, err := strconv.ParseFloat(astVal.Number, 64)
		if err != nil {
			ctx.AddIssueWithMeta("number.parse", "invalid number", map[string]any{"error": err.Error()})
			return nil, ctx.Error()
		}

		if minSet && floatVal < minValue {
			ctx.AddIssueWithMeta("number.min", "too small", map[string]any{"min": minValue, "actual": floatVal})
			return nil, ctx.Error()
		}

		if maxSet && floatVal > maxValue {
			ctx.AddIssueWithMeta("number.max", "too large", map[string]any{"max": maxValue, "actual": floatVal})
			return nil, ctx.Error()
		}

		result := convertToFieldType(floatVal, fieldType)

		// TRANSFORM HERE
		return fn(result)
	}

	compiled, err := newFieldFromInfo(b.fieldInfo, validator)
	if err != nil {
		b.schema.buildError = err
		return b.schema
	}
	compiled.required = b.required

	if b.fieldIndex == -1 {
		b.schema.fields = append(b.schema.fields, compiled)
		b.schema.lastFieldIndex = len(b.schema.fields) - 1
	} else {
		b.schema.fields[b.fieldIndex] = compiled
		b.schema.lastFieldIndex = b.fieldIndex
	}

	return b.schema
}

func convertToFieldType(floatVal float64, fieldType reflect.Type) any {
	switch fieldType.Kind() {
	case reflect.Int:
		return int(floatVal)
	case reflect.Int8:
		return int8(floatVal)
	case reflect.Int16:
		return int16(floatVal)
	case reflect.Int32:
		return int32(floatVal)
	case reflect.Int64:
		return int64(floatVal)
	case reflect.Uint:
		return uint(floatVal)
	case reflect.Uint8:
		return uint8(floatVal)
	case reflect.Uint16:
		return uint16(floatVal)
	case reflect.Uint32:
		return uint32(floatVal)
	case reflect.Uint64:
		return uint64(floatVal)
	case reflect.Float32:
		return float32(floatVal)
	case reflect.Float64:
		return floatVal
	default:
		return floatVal
	}
}
