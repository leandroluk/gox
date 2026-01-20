// schema/object/durationfieldbuilder.go
package object

import (
	"time"

	"github.com/leandroluk/go/validate/internal/ast"
	"github.com/leandroluk/go/validate/internal/engine"
	"github.com/leandroluk/go/validate/schema"
	"github.com/leandroluk/go/validate/schema/duration"
)

type DurationFieldBuilder[T any] struct {
	schema         *Schema[T]
	fieldInfo      fieldInfo[T]
	durationSchema *duration.Schema
	fieldIndex     int
	required       bool
}

func (b *DurationFieldBuilder[T]) build() {
	if b.schema.buildError != nil {
		return
	}

	validator := func(ctx *engine.Context, value any) (any, error) {
		if b.required {
			astVal, ok := value.(ast.Value)
			if ok && (astVal.IsMissing() || astVal.IsNull()) {
				ctx.AddIssue("duration.required", "required")
				return nil, ctx.Error()
			}
		}
		return b.durationSchema.ValidateAny(value, ctx.Options)
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

func (b *DurationFieldBuilder[T]) Required() *DurationFieldBuilder[T] {
	b.durationSchema.Required()
	b.required = true
	b.build()
	return b
}

func (b *DurationFieldBuilder[T]) Default(value time.Duration) *DurationFieldBuilder[T] {
	b.durationSchema.Default(value)
	b.build()
	return b
}

func (b *DurationFieldBuilder[T]) DefaultFunc(fn func() time.Duration) *DurationFieldBuilder[T] {
	b.durationSchema.DefaultFunc(fn)
	b.build()
	return b
}

func (b *DurationFieldBuilder[T]) Min(value time.Duration) *DurationFieldBuilder[T] {
	b.durationSchema.Min(value)
	b.build()
	return b
}

func (b *DurationFieldBuilder[T]) Max(value time.Duration) *DurationFieldBuilder[T] {
	b.durationSchema.Max(value)
	b.build()
	return b
}

func (b *DurationFieldBuilder[T]) RequiredIf(path string, op ConditionOp, expected any) *DurationFieldBuilder[T] {
	b.build()
	b.schema.RequiredIf(path, op, expected)
	return b
}

func (b *DurationFieldBuilder[T]) RequiredWith(paths ...string) *DurationFieldBuilder[T] {
	b.build()
	b.schema.RequiredWith(paths...)
	return b
}

func (b *DurationFieldBuilder[T]) RequiredWithout(paths ...string) *DurationFieldBuilder[T] {
	b.build()
	b.schema.RequiredWithout(paths...)
	return b
}

func (b *DurationFieldBuilder[T]) ExcludedIf(path string, op ConditionOp, expected any) *DurationFieldBuilder[T] {
	b.build()
	b.schema.ExcludedIf(path, op, expected)
	return b
}

func (b *DurationFieldBuilder[T]) SkipUnless(path string, op ConditionOp, expected any) *DurationFieldBuilder[T] {
	b.build()
	b.schema.SkipUnless(path, op, expected)
	return b
}

func (b *DurationFieldBuilder[T]) EqField(other string) *DurationFieldBuilder[T] {
	b.build()
	b.schema.EqField(other)
	return b
}

func (b *DurationFieldBuilder[T]) NeField(other string) *DurationFieldBuilder[T] {
	b.build()
	b.schema.NeField(other)
	return b
}

func (b *DurationFieldBuilder[T]) GtField(other string) *DurationFieldBuilder[T] {
	b.build()
	b.schema.GtField(other)
	return b
}

func (b *DurationFieldBuilder[T]) GteField(other string) *DurationFieldBuilder[T] {
	b.build()
	b.schema.GteField(other)
	return b
}

func (b *DurationFieldBuilder[T]) LtField(other string) *DurationFieldBuilder[T] {
	b.build()
	b.schema.LtField(other)
	return b
}

func (b *DurationFieldBuilder[T]) LteField(other string) *DurationFieldBuilder[T] {
	b.build()
	b.schema.LteField(other)
	return b
}

func (b *DurationFieldBuilder[T]) EqCSField(path string) *DurationFieldBuilder[T] {
	b.build()
	b.schema.EqCSField(path)
	return b
}

func (b *DurationFieldBuilder[T]) NeCSField(path string) *DurationFieldBuilder[T] {
	b.build()
	b.schema.NeCSField(path)
	return b
}

func (b *DurationFieldBuilder[T]) GtCSField(path string) *DurationFieldBuilder[T] {
	b.build()
	b.schema.GtCSField(path)
	return b
}

func (b *DurationFieldBuilder[T]) GteCSField(path string) *DurationFieldBuilder[T] {
	b.build()
	b.schema.GteCSField(path)
	return b
}

func (b *DurationFieldBuilder[T]) LtCSField(path string) *DurationFieldBuilder[T] {
	b.build()
	b.schema.LtCSField(path)
	return b
}

func (b *DurationFieldBuilder[T]) LteCSField(path string) *DurationFieldBuilder[T] {
	b.build()
	b.schema.LteCSField(path)
	return b
}

func (b *DurationFieldBuilder[T]) ValidateAny(value any, options schema.Options) (any, error) {
	return b.durationSchema.ValidateAny(value, options)
}

func (b *DurationFieldBuilder[T]) Build() *Schema[T] {
	b.build()
	return b.schema
}

// Transform registers a transformation function for the field.
// It validates the value as a Duration first, then applies the transformation.
// The returned value is used as the new value for the field.
func (b *DurationFieldBuilder[T]) Transform(fn func(value any) (any, error)) *Schema[T] {
	validator := func(ctx *engine.Context, value any) (any, error) {
		if b.required {
			astVal, ok := value.(ast.Value)
			if ok && (astVal.IsMissing() || astVal.IsNull()) {
				ctx.AddIssue("duration.required", "required")
				return nil, ctx.Error()
			}
		}

		out, err := b.durationSchema.ValidateAny(value, ctx.Options)
		if err != nil {
			return nil, err
		}

		return fn(out)
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
