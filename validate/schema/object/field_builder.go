// schema/object/fieldbuilder.go
package object

import (
	"reflect"

	"github.com/leandroluk/go/validate/internal/engine"
	"github.com/leandroluk/go/validate/schema"
	"github.com/leandroluk/go/validate/schema/boolean"
	"github.com/leandroluk/go/validate/schema/date"
	"github.com/leandroluk/go/validate/schema/duration"
	"github.com/leandroluk/go/validate/schema/text"
)

type FieldBuilder[T any] struct {
	schema    *Schema[T]
	fieldInfo fieldInfo[T]
}

func newFieldBuilder[T any](schema *Schema[T], fieldPointer any) *FieldBuilder[T] {
	if schema == nil || schema.buildTarget == nil {
		schema.buildError = ErrInvalidBuilderUsage
		return &FieldBuilder[T]{schema: schema}
	}

	info, err := resolveFieldInfo(schema.buildTarget, fieldPointer)
	if err != nil {
		schema.buildError = err
		return &FieldBuilder[T]{schema: schema}
	}

	return &FieldBuilder[T]{
		schema:    schema,
		fieldInfo: info,
	}
}

func (fb *FieldBuilder[T]) Text() *TextFieldBuilder[T] {
	if fb.schema.buildError != nil {
		return &TextFieldBuilder[T]{schema: fb.schema, fieldIndex: -1}
	}

	b := &TextFieldBuilder[T]{
		schema:     fb.schema,
		fieldInfo:  fb.fieldInfo,
		textSchema: text.New(),
		fieldIndex: -1,
	}
	b.build()
	return b
}

func (fb *FieldBuilder[T]) Number() *NumberFieldBuilder[T] {
	if fb.schema.buildError != nil {
		return &NumberFieldBuilder[T]{schema: fb.schema, fieldIndex: -1}
	}

	b := &NumberFieldBuilder[T]{
		schema:     fb.schema,
		fieldInfo:  fb.fieldInfo,
		fieldIndex: -1,
	}
	b.build()
	return b
}

func (fb *FieldBuilder[T]) Boolean() *BooleanFieldBuilder[T] {
	if fb.schema.buildError != nil {
		return &BooleanFieldBuilder[T]{schema: fb.schema, fieldIndex: -1}
	}

	b := &BooleanFieldBuilder[T]{
		schema:        fb.schema,
		fieldInfo:     fb.fieldInfo,
		booleanSchema: boolean.New(),
		fieldIndex:    -1,
	}
	b.build()
	return b
}

func (fb *FieldBuilder[T]) Date() *DateFieldBuilder[T] {
	if fb.schema.buildError != nil {
		return &DateFieldBuilder[T]{schema: fb.schema, fieldIndex: -1}
	}

	b := &DateFieldBuilder[T]{
		schema:     fb.schema,
		fieldInfo:  fb.fieldInfo,
		dateSchema: date.New(),
		fieldIndex: -1,
	}
	b.build()
	return b
}

func (fb *FieldBuilder[T]) Duration() *DurationFieldBuilder[T] {
	if fb.schema.buildError != nil {
		return &DurationFieldBuilder[T]{schema: fb.schema, fieldIndex: -1}
	}

	b := &DurationFieldBuilder[T]{
		schema:         fb.schema,
		fieldInfo:      fb.fieldInfo,
		durationSchema: duration.New(),
		fieldIndex:     -1,
	}
	b.build()
	return b
}

// Array starts a builder for array validation.
// You can optionally pass a schema to validate items.
//
//	s.Field(&u.Tags).Array(v.Text().Min(3)).Min(1)
func (fb *FieldBuilder[T]) Array(items ...schema.AnySchema) *ArrayFieldBuilder[T] {
	if fb.schema.buildError != nil {
		return &ArrayFieldBuilder[T]{schema: fb.schema, fieldIndex: -1}
	}

	var itemSchema schema.AnySchema
	if len(items) > 0 {
		itemSchema = items[0]
	}

	b := &ArrayFieldBuilder[T]{
		schema:     fb.schema,
		fieldInfo:  fb.fieldInfo,
		itemSchema: itemSchema,
		fieldIndex: -1,
	}
	b.build()
	return b
}

// Object starts a builder for nested object validation.
// It accepts a builder function with the signature func(target *N, schema *v.ObjectSchema[N]).
//
//	s.Field(&u.Address).Object(func(a *Address, s *v.ObjectSchema[Address]) {
//	    s.Field(&a.City).Text().Required()
//	})
func (fb *FieldBuilder[T]) Object(builderFunc any) *ObjectFieldBuilder[T] {
	if fb.schema.buildError != nil {
		return &ObjectFieldBuilder[T]{schema: fb.schema, fieldIndex: -1}
	}

	b := &ObjectFieldBuilder[T]{
		schema:      fb.schema,
		fieldInfo:   fb.fieldInfo,
		builderFunc: builderFunc,
		fieldIndex:  -1,
	}
	b.build()
	return b
}

// Record starts a builder for map/record validation.
//
//	s.Field(&u.Metadata).Record().Min(1)
func (fb *FieldBuilder[T]) Record(items ...schema.AnySchema) *RecordFieldBuilder[T] {
	if fb.schema.buildError != nil {
		return &RecordFieldBuilder[T]{schema: fb.schema, fieldIndex: -1}
	}

	var valueSchema schema.AnySchema
	if len(items) > 0 {
		valueSchema = items[0]
	}
	// Key schema? Usually arg 2? Or separate method?
	// For simplicity, just Value schema as arg 1.

	b := &RecordFieldBuilder[T]{
		schema:      fb.schema,
		fieldInfo:   fb.fieldInfo,
		valueSchema: valueSchema,
		fieldIndex:  -1,
	}
	b.build()
	return b
}

// Transform registers a transformation function for the field.
// It ignores the context issues unless the function returns an error.
// The returned value is used as the new value for the field.
func (fb *FieldBuilder[T]) Transform(fn func(value any) (any, error)) *Schema[T] {
	return fb.Custom(func(ctx *engine.Context, value any) (any, error) {
		return fn(value)
	})
}

// Custom registers a custom validator/transformer for the field.
// It allows full access to the validation context.
func (fb *FieldBuilder[T]) Custom(validator func(ctx *engine.Context, value any) (any, error)) *Schema[T] {
	if fb.schema.buildError != nil {
		return fb.schema
	}

	compiled, err := newFieldFromInfo(fb.fieldInfo, validator)
	if err != nil {
		fb.schema.buildError = err
		return fb.schema
	}

	fb.schema.fields = append(fb.schema.fields, compiled)
	fb.schema.lastFieldIndex = len(fb.schema.fields) - 1
	return fb.schema
}

type fieldInfo[T any] struct {
	name      string
	offset    uintptr
	fieldType reflect.Type
}
