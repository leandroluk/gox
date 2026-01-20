// schema/object/textfieldbuilder.go
package object

import (
	"github.com/leandroluk/go/validator/internal/ast"
	"github.com/leandroluk/go/validator/internal/engine"
	"github.com/leandroluk/go/validator/schema"
	"github.com/leandroluk/go/validator/schema/text"
)

type TextFieldBuilder[T any] struct {
	schema     *Schema[T]
	fieldInfo  fieldInfo[T]
	textSchema *text.Schema
	fieldIndex int
	required   bool
}

func (b *TextFieldBuilder[T]) build() {
	if b.schema.buildError != nil {
		return
	}

	validator := func(ctx *engine.Context, value any) (any, error) {
		// Manually check required to ensure it works even if applyFieldPlan logic changes
		if b.required {
			astVal, ok := value.(ast.Value)
			if ok && (astVal.IsMissing() || astVal.IsNull()) {
				ctx.AddIssue("text.required", "required")
				return nil, ctx.Error()
			}
			// If not ast.Value, we assume it's present or handled by ValidateAny
		}
		return b.textSchema.ValidateAny(value, ctx.Options)
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

func (b *TextFieldBuilder[T]) Required() *TextFieldBuilder[T] {
	b.textSchema.Required()
	b.required = true
	b.build()
	return b
}

func (b *TextFieldBuilder[T]) Default(value string) *TextFieldBuilder[T] {
	b.textSchema.Default(value)
	b.build()
	return b
}

func (b *TextFieldBuilder[T]) DefaultFunc(fn func() string) *TextFieldBuilder[T] {
	b.textSchema.DefaultFunc(fn)
	b.build()
	return b
}

func (b *TextFieldBuilder[T]) Len(n int) *TextFieldBuilder[T] {
	b.textSchema.Len(n)
	b.build()
	return b
}

func (b *TextFieldBuilder[T]) Min(n int) *TextFieldBuilder[T] {
	b.textSchema.Min(n)
	b.build()
	return b
}

func (b *TextFieldBuilder[T]) Max(n int) *TextFieldBuilder[T] {
	b.textSchema.Max(n)
	b.build()
	return b
}

func (b *TextFieldBuilder[T]) Email() *TextFieldBuilder[T] {
	b.textSchema.Email()
	b.build()
	return b
}

func (b *TextFieldBuilder[T]) URL() *TextFieldBuilder[T] {
	b.textSchema.URL()
	b.build()
	return b
}

func (b *TextFieldBuilder[T]) URI() *TextFieldBuilder[T] {
	b.textSchema.URI()
	b.build()
	return b
}

func (b *TextFieldBuilder[T]) UUID() *TextFieldBuilder[T] {
	b.textSchema.UUID()
	b.build()
	return b
}

func (b *TextFieldBuilder[T]) Pattern(pattern string) *TextFieldBuilder[T] {
	b.textSchema.Pattern(pattern)
	b.build()
	return b
}

func (b *TextFieldBuilder[T]) OneOf(values ...string) *TextFieldBuilder[T] {
	b.textSchema.OneOf(values...)
	b.build()
	return b
}

func (b *TextFieldBuilder[T]) RequiredIf(path string, op ConditionOp, expected any) *TextFieldBuilder[T] {
	b.build()
	b.schema.RequiredIf(path, op, expected)
	return b
}

func (b *TextFieldBuilder[T]) RequiredWith(paths ...string) *TextFieldBuilder[T] {
	b.build()
	b.schema.RequiredWith(paths...)
	return b
}

func (b *TextFieldBuilder[T]) RequiredWithout(paths ...string) *TextFieldBuilder[T] {
	b.build()
	b.schema.RequiredWithout(paths...)
	return b
}

func (b *TextFieldBuilder[T]) ExcludedIf(path string, op ConditionOp, expected any) *TextFieldBuilder[T] {
	b.build()
	b.schema.ExcludedIf(path, op, expected)
	return b
}

func (b *TextFieldBuilder[T]) SkipUnless(path string, op ConditionOp, expected any) *TextFieldBuilder[T] {
	b.build()
	b.schema.SkipUnless(path, op, expected)
	return b
}

func (b *TextFieldBuilder[T]) EqField(other string) *TextFieldBuilder[T] {
	b.build()
	b.schema.EqField(other)
	return b
}

func (b *TextFieldBuilder[T]) NeField(other string) *TextFieldBuilder[T] {
	b.build()
	b.schema.NeField(other)
	return b
}

func (b *TextFieldBuilder[T]) GtField(other string) *TextFieldBuilder[T] {
	b.build()
	b.schema.GtField(other)
	return b
}

func (b *TextFieldBuilder[T]) GteField(other string) *TextFieldBuilder[T] {
	b.build()
	b.schema.GteField(other)
	return b
}

func (b *TextFieldBuilder[T]) LtField(other string) *TextFieldBuilder[T] {
	b.build()
	b.schema.LtField(other)
	return b
}

func (b *TextFieldBuilder[T]) LteField(other string) *TextFieldBuilder[T] {
	b.build()
	b.schema.LteField(other)
	return b
}

func (b *TextFieldBuilder[T]) EqCSField(path string) *TextFieldBuilder[T] {
	b.build()
	b.schema.EqCSField(path)
	return b
}

func (b *TextFieldBuilder[T]) NeCSField(path string) *TextFieldBuilder[T] {
	b.build()
	b.schema.NeCSField(path)
	return b
}

func (b *TextFieldBuilder[T]) GtCSField(path string) *TextFieldBuilder[T] {
	b.build()
	b.schema.GtCSField(path)
	return b
}

func (b *TextFieldBuilder[T]) GteCSField(path string) *TextFieldBuilder[T] {
	b.build()
	b.schema.GteCSField(path)
	return b
}

func (b *TextFieldBuilder[T]) LtCSField(path string) *TextFieldBuilder[T] {
	b.build()
	b.schema.LtCSField(path)
	return b
}

func (b *TextFieldBuilder[T]) LteCSField(path string) *TextFieldBuilder[T] {
	b.build()
	b.schema.LteCSField(path)
	return b
}

func (b *TextFieldBuilder[T]) FieldContains(other string) *TextFieldBuilder[T] {
	b.build()
	b.schema.FieldContains(other)
	return b
}

func (b *TextFieldBuilder[T]) FieldExcludes(other string) *TextFieldBuilder[T] {
	b.build()
	b.schema.FieldExcludes(other)
	return b
}

func (b *TextFieldBuilder[T]) ValidateAny(value any, options schema.Options) (any, error) {
	return b.textSchema.ValidateAny(value, options)
}

func (b *TextFieldBuilder[T]) Build() *Schema[T] {
	b.build()
	return b.schema
}

// Transform registers a transformation function for the field.
// It validates the value as Text first, then applies the transformation.
// The returned value is used as the new value for the field.
func (b *TextFieldBuilder[T]) Transform(fn func(value any) (any, error)) *Schema[T] {
	b.build() // Ensure any previous mutations on textSchema are applied?
	// Actually build() registers the field. If we call Transform, we are overtaking the registration.
	// We need to NOT call build() separately if we use Custom logic, OR we need to use a custom validator that uses textSchema.

	// The current b.build() creates a validator and appends to fields.
	// We want to create a NEW validator that uses b.textSchema + fn, and register THAT.

	validator := func(ctx *engine.Context, value any) (any, error) {
		if b.required {
			astVal, ok := value.(ast.Value)
			if ok && (astVal.IsMissing() || astVal.IsNull()) {
				ctx.AddIssue("text.required", "required")
				return nil, ctx.Error()
			}
		}

		// 1. Validate as Text
		// We use b.textSchema directly.
		out, err := b.textSchema.ValidateAny(value, ctx.Options)
		if err != nil {
			return nil, err
		}
		// If textSchema returned nil (soft failure or optional), we might skip transform?
		// But ValidateAny usually returns value if success.

		// 2. Transform
		return fn(out)
	}

	compiled, err := newFieldFromInfo(b.fieldInfo, validator)
	if err != nil {
		b.schema.buildError = err
		return b.schema
	}
	compiled.required = b.required

	// Overwrite or Append?
	// If b.fieldIndex == -1, we append.
	// BUT b.build() might have been called by previous methods if they were fluent?
	// In the current design, every method like Min() calls build().
	// So the field is ALREADY registered.
	// We need to UPDATE it.

	if b.fieldIndex == -1 {
		// Should not happen if we are chained from Text() which calls build().
		// But safe to handle.
		b.schema.fields = append(b.schema.fields, compiled)
		b.schema.lastFieldIndex = len(b.schema.fields) - 1
	} else {
		b.schema.fields[b.fieldIndex] = compiled
		b.schema.lastFieldIndex = b.fieldIndex
	}

	return b.schema
}

var _ ast.Value
