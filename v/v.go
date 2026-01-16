// v.go
package v

import (
	"fmt"
	"reflect"
	"time"

	"github.com/leandroluk/go/v/internal/ast"
	"github.com/leandroluk/go/v/internal/engine"
	"github.com/leandroluk/go/v/internal/issues"
	"github.com/leandroluk/go/v/internal/registry"
	"github.com/leandroluk/go/v/internal/types"
	"github.com/leandroluk/go/v/schema"
	"github.com/leandroluk/go/v/schema/array"
	"github.com/leandroluk/go/v/schema/boolean"
	"github.com/leandroluk/go/v/schema/combinator"
	"github.com/leandroluk/go/v/schema/date"
	"github.com/leandroluk/go/v/schema/duration"
	"github.com/leandroluk/go/v/schema/number"
	"github.com/leandroluk/go/v/schema/object"
	"github.com/leandroluk/go/v/schema/record"
	"github.com/leandroluk/go/v/schema/text"
)

type Value = ast.Value
type Context = engine.Context
type Issue = issues.Issue
type ValidationError = issues.ValidationError

type Options = schema.Options
type Option = schema.Option
type Formatter = schema.Formatter
type AnySchema = schema.AnySchema

type ObjectSchema[T any] = object.Schema[T]

type ArraySchema[E any] = array.Schema[E]

type RecordSchema[V any] = record.Schema[V]

type TextSchema = text.Schema

type BooleanSchema = boolean.Schema

type DateSchema = date.Schema

type DurationSchema = duration.Schema

type NumberSchema[N types.Number] = number.Schema[N]

const CodeOneOf = combinator.CodeOneOf

type CombinatorSchema[T any] = combinator.Schema[T]
type AnyOfSchema[T any] = combinator.AnyOfSchema[T]
type OneOfSchema[T any] = combinator.OneOfSchema[T]

func WithFailFast(value bool) Option {
	return schema.WithFailFast(value)
}

func WithMaxIssues(value int) Option {
	return schema.WithMaxIssues(value)
}

func WithDefaultOnNull(value bool) Option {
	return schema.WithDefaultOnNull(value)
}

func WithCoerce(value bool) Option {
	return schema.WithCoerce(value)
}

func WithOmitZero(value bool) Option {
	return schema.WithOmitZero(value)
}

func WithFormatter(formatter Formatter) Option {
	return schema.WithFormatter(formatter)
}

func WithCoerceTrimSpace(value bool) Option {
	return schema.WithCoerceTrimSpace(value)
}

func WithCoerceNumberUnderscore(value bool) Option {
	return schema.WithCoerceNumberUnderscore(value)
}

func WithCoerceDateUnixSeconds(value bool) Option {
	return schema.WithCoerceDateUnixSeconds(value)
}

func WithCoerceDateUnixMilliseconds(value bool) Option {
	return schema.WithCoerceDateUnixMilliseconds(value)
}

func WithCoerceDurationSeconds(value bool) Option {
	return schema.WithCoerceDurationSeconds(value)
}

func WithCoerceDurationMilliseconds(value bool) Option {
	return schema.WithCoerceDurationMilliseconds(value)
}

func WithTimeLocation(value *time.Location) Option {
	return schema.WithTimeLocation(value)
}

func WithDateLayouts(layouts ...string) Option {
	return schema.WithDateLayouts(layouts...)
}

func WithAdditionalDateLayouts(layouts ...string) Option {
	return schema.WithAdditionalDateLayouts(layouts...)
}

func Register(schemaValue AnySchema) {
	registry.Register(schemaValue)
}

func ResetRegistry() {
	registry.Reset()
}

func Object[T any](builder func(target *T, schemaValue *object.Schema[T])) *object.Schema[T] {
	schemaValue := object.New(builder)
	registry.Register(schemaValue)
	return schemaValue
}

func Array[E any]() *array.Schema[E] {
	schemaValue := array.New[E]()
	registry.Register(schemaValue)
	return schemaValue
}

func Record[V any]() *record.Schema[V] {
	schemaValue := record.New[V]()
	registry.Register(schemaValue)
	return schemaValue
}

func Text() *text.Schema {
	schemaValue := text.New()
	registry.Register(schemaValue)
	return schemaValue
}

func Boolean() *boolean.Schema {
	schemaValue := boolean.New()
	registry.Register(schemaValue)
	return schemaValue
}

func Date() *date.Schema {
	schemaValue := date.New()
	registry.Register(schemaValue)
	return schemaValue
}

func Duration() *duration.Schema {
	schemaValue := duration.New()
	registry.Register(schemaValue)
	return schemaValue
}

func NumberSchemaOf[N types.Number]() *number.Schema[N] {
	schemaValue := number.New[N]()
	registry.Register(schemaValue)
	return schemaValue
}

func AnyOf[T any](schemaList ...combinator.Schema[T]) *combinator.AnyOfSchema[T] {
	schemaValue := combinator.AnyOf(schemaList...)
	registry.Register(schemaValue)
	return schemaValue
}

func OneOf[T any](schemaList ...combinator.Schema[T]) *combinator.OneOfSchema[T] {
	schemaValue := combinator.OneOf(schemaList...)
	registry.Register(schemaValue)
	return schemaValue
}

func Validate[T any](input any, optionList ...Option) (T, error) {
	var zero T

	schemaValue, ok := registry.LookupTyped[T]()
	if !ok {
		outputType := reflect.TypeFor[T]()
		return zero, fmt.Errorf("no schema registered for %s", outputType.String())
	}

	options := schema.ApplyOptions(optionList...)
	output, err := schemaValue.ValidateAny(input, options)
	if err != nil {
		return zero, err
	}

	typed, ok := output.(T)
	if !ok {
		expectedType := reflect.TypeFor[T]()
		actualType := reflect.TypeOf(output)
		return zero, fmt.Errorf("schema returned incompatible type (expected %s, got %v)", expectedType.String(), actualType)
	}

	return typed, nil
}
