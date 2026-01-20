// schema/object/booleanfieldbuilder.go
package object

import (
	"github.com/leandroluk/go/validate/internal/ast"
	"github.com/leandroluk/go/validate/internal/engine"
	"github.com/leandroluk/go/validate/schema"
	"github.com/leandroluk/go/validate/schema/boolean"
)

type BooleanFieldBuilder[T any] struct {
	schema        *Schema[T]
	fieldInfo     fieldInfo[T]
	booleanSchema *boolean.Schema
	fieldIndex    int
	required      bool
}

func (b *BooleanFieldBuilder[T]) build() {
	if b.schema.buildError != nil {
		return
	}

	validator := func(ctx *engine.Context, value any) (any, error) {
		if b.required {
			astVal, ok := value.(ast.Value)
			if ok && (astVal.IsMissing() || astVal.IsNull()) {
				ctx.AddIssue("boolean.required", "required")
				return nil, ctx.Error()
			}
		}
		return b.booleanSchema.ValidateAny(value, ctx.Options)
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

func (b *BooleanFieldBuilder[T]) Required() *BooleanFieldBuilder[T] {
	b.booleanSchema.Required()
	b.required = true
	b.build()
	return b
}

func (b *BooleanFieldBuilder[T]) IsDefault() *BooleanFieldBuilder[T] {
	b.booleanSchema.IsDefault()
	b.build()
	return b
}

func (b *BooleanFieldBuilder[T]) Default(value bool) *BooleanFieldBuilder[T] {
	b.booleanSchema.Default(value)
	b.build()
	return b
}

func (b *BooleanFieldBuilder[T]) DefaultFunc(fn func() bool) *BooleanFieldBuilder[T] {
	b.booleanSchema.DefaultFunc(fn)
	b.build()
	return b
}

func (b *BooleanFieldBuilder[T]) RequiredIf(path string, op ConditionOp, expected any) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.RequiredIf(path, op, expected)
	return b
}

func (b *BooleanFieldBuilder[T]) RequiredWith(paths ...string) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.RequiredWith(paths...)
	return b
}

func (b *BooleanFieldBuilder[T]) RequiredWithout(paths ...string) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.RequiredWithout(paths...)
	return b
}

func (b *BooleanFieldBuilder[T]) ExcludedIf(path string, op ConditionOp, expected any) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.ExcludedIf(path, op, expected)
	return b
}

func (b *BooleanFieldBuilder[T]) SkipUnless(path string, op ConditionOp, expected any) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.SkipUnless(path, op, expected)
	return b
}

func (b *BooleanFieldBuilder[T]) EqField(other string) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.EqField(other)
	return b
}

func (b *BooleanFieldBuilder[T]) NeField(other string) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.NeField(other)
	return b
}

func (b *BooleanFieldBuilder[T]) GtField(other string) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.GtField(other)
	return b
}

func (b *BooleanFieldBuilder[T]) GteField(other string) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.GteField(other)
	return b
}

func (b *BooleanFieldBuilder[T]) LtField(other string) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.LtField(other)
	return b
}

func (b *BooleanFieldBuilder[T]) LteField(other string) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.LteField(other)
	return b
}

func (b *BooleanFieldBuilder[T]) EqCSField(path string) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.EqCSField(path)
	return b
}

func (b *BooleanFieldBuilder[T]) NeCSField(path string) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.NeCSField(path)
	return b
}

func (b *BooleanFieldBuilder[T]) GtCSField(path string) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.GtCSField(path)
	return b
}

func (b *BooleanFieldBuilder[T]) GteCSField(path string) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.GteCSField(path)
	return b
}

func (b *BooleanFieldBuilder[T]) LtCSField(path string) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.LtCSField(path)
	return b
}

func (b *BooleanFieldBuilder[T]) LteCSField(path string) *BooleanFieldBuilder[T] {
	b.build()
	b.schema.LteCSField(path)
	return b
}

func (b *BooleanFieldBuilder[T]) ValidateAny(value any, options schema.Options) (any, error) {
	return b.booleanSchema.ValidateAny(value, options)
}

// Transform registers a transformation function for the field.
// It validates the value as a Boolean first, then applies the transformation.
// The returned value is used as the new value for the field.
func (b *BooleanFieldBuilder[T]) Transform(fn func(value any) (any, error)) *Schema[T] {
	validator := func(ctx *engine.Context, value any) (any, error) {
		if b.required {
			astVal, ok := value.(ast.Value)
			if ok && (astVal.IsMissing() || astVal.IsNull()) {
				ctx.AddIssue("boolean.required", "required")
				return nil, ctx.Error()
			}
		}

		out, err := b.booleanSchema.ValidateAny(value, ctx.Options)
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

func (b *BooleanFieldBuilder[T]) Build() *Schema[T] {
	b.build()
	return b.schema
}
