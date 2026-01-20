// schema/object/datefieldbuilder.go
package object

import (
	"time"

	"github.com/leandroluk/go/validator/internal/ast"
	"github.com/leandroluk/go/validator/internal/engine"
	"github.com/leandroluk/go/validator/schema"
	"github.com/leandroluk/go/validator/schema/date"
)

type DateFieldBuilder[T any] struct {
	schema     *Schema[T]
	fieldInfo  fieldInfo[T]
	dateSchema *date.Schema
	fieldIndex int
	required   bool
}

func (b *DateFieldBuilder[T]) build() {
	if b.schema.buildError != nil {
		return
	}

	validator := func(ctx *engine.Context, value any) (any, error) {
		if b.required {
			astVal, ok := value.(ast.Value)
			if ok && (astVal.IsMissing() || astVal.IsNull()) {
				ctx.AddIssue("date.required", "required")
				return nil, ctx.Error()
			}
		}
		return b.dateSchema.ValidateAny(value, ctx.Options)
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

func (b *DateFieldBuilder[T]) Required() *DateFieldBuilder[T] {
	b.dateSchema.Required()
	b.required = true
	b.build()
	return b
}

func (b *DateFieldBuilder[T]) Default(value time.Time) *DateFieldBuilder[T] {
	b.dateSchema.Default(value)
	b.build()
	return b
}

func (b *DateFieldBuilder[T]) DefaultFunc(fn func() time.Time) *DateFieldBuilder[T] {
	b.dateSchema.DefaultFunc(fn)
	b.build()
	return b
}

func (b *DateFieldBuilder[T]) Min(value time.Time) *DateFieldBuilder[T] {
	b.dateSchema.Min(value)
	b.build()
	return b
}

func (b *DateFieldBuilder[T]) Max(value time.Time) *DateFieldBuilder[T] {
	b.dateSchema.Max(value)
	b.build()
	return b
}

func (b *DateFieldBuilder[T]) RequiredIf(path string, op ConditionOp, expected any) *DateFieldBuilder[T] {
	b.build()
	b.schema.RequiredIf(path, op, expected)
	return b
}

func (b *DateFieldBuilder[T]) RequiredWith(paths ...string) *DateFieldBuilder[T] {
	b.build()
	b.schema.RequiredWith(paths...)
	return b
}

func (b *DateFieldBuilder[T]) RequiredWithout(paths ...string) *DateFieldBuilder[T] {
	b.build()
	b.schema.RequiredWithout(paths...)
	return b
}

func (b *DateFieldBuilder[T]) ExcludedIf(path string, op ConditionOp, expected any) *DateFieldBuilder[T] {
	b.build()
	b.schema.ExcludedIf(path, op, expected)
	return b
}

func (b *DateFieldBuilder[T]) SkipUnless(path string, op ConditionOp, expected any) *DateFieldBuilder[T] {
	b.build()
	b.schema.SkipUnless(path, op, expected)
	return b
}

func (b *DateFieldBuilder[T]) EqField(other string) *DateFieldBuilder[T] {
	b.build()
	b.schema.EqField(other)
	return b
}

func (b *DateFieldBuilder[T]) NeField(other string) *DateFieldBuilder[T] {
	b.build()
	b.schema.NeField(other)
	return b
}

func (b *DateFieldBuilder[T]) GtField(other string) *DateFieldBuilder[T] {
	b.build()
	b.schema.GtField(other)
	return b
}

func (b *DateFieldBuilder[T]) GteField(other string) *DateFieldBuilder[T] {
	b.build()
	b.schema.GteField(other)
	return b
}

func (b *DateFieldBuilder[T]) LtField(other string) *DateFieldBuilder[T] {
	b.build()
	b.schema.LtField(other)
	return b
}

func (b *DateFieldBuilder[T]) LteField(other string) *DateFieldBuilder[T] {
	b.build()
	b.schema.LteField(other)
	return b
}

func (b *DateFieldBuilder[T]) EqCSField(path string) *DateFieldBuilder[T] {
	b.build()
	b.schema.EqCSField(path)
	return b
}

func (b *DateFieldBuilder[T]) NeCSField(path string) *DateFieldBuilder[T] {
	b.build()
	b.schema.NeCSField(path)
	return b
}

func (b *DateFieldBuilder[T]) GtCSField(path string) *DateFieldBuilder[T] {
	b.build()
	b.schema.GtCSField(path)
	return b
}

func (b *DateFieldBuilder[T]) GteCSField(path string) *DateFieldBuilder[T] {
	b.build()
	b.schema.GteCSField(path)
	return b
}

func (b *DateFieldBuilder[T]) LtCSField(path string) *DateFieldBuilder[T] {
	b.build()
	b.schema.LtCSField(path)
	return b
}

func (b *DateFieldBuilder[T]) LteCSField(path string) *DateFieldBuilder[T] {
	b.build()
	b.schema.LteCSField(path)
	return b
}

func (b *DateFieldBuilder[T]) ValidateAny(value any, options schema.Options) (any, error) {
	return b.dateSchema.ValidateAny(value, options)
}

func (b *DateFieldBuilder[T]) Build() *Schema[T] {
	b.build()
	return b.schema
}

// Transform registers a transformation function for the field.
// It validates the value as a Date first, then applies the transformation.
// The returned value is used as the new value for the field.
func (b *DateFieldBuilder[T]) Transform(fn func(value any) (any, error)) *Schema[T] {
	validator := func(ctx *engine.Context, value any) (any, error) {
		if b.required {
			astVal, ok := value.(ast.Value)
			if ok && (astVal.IsMissing() || astVal.IsNull()) {
				ctx.AddIssue("date.required", "required")
				return nil, ctx.Error()
			}
		}

		out, err := b.dateSchema.ValidateAny(value, ctx.Options)
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
