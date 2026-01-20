// Package validator provides a fluent, type-safe, and AST-based validation library for Go.
//
// It separates schema definition from domain models, supports distinguishing between "null" and "missing" fields,
// and offers a rich set of composable validation rules.
package validator

import (
	"fmt"
	"reflect"
	"time"

	"github.com/leandroluk/go/validator/internal/ast"
	"github.com/leandroluk/go/validator/internal/engine"
	"github.com/leandroluk/go/validator/internal/issues"
	"github.com/leandroluk/go/validator/internal/registry"
	"github.com/leandroluk/go/validator/internal/types"
	"github.com/leandroluk/go/validator/schema"
	"github.com/leandroluk/go/validator/schema/array"
	"github.com/leandroluk/go/validator/schema/boolean"
	"github.com/leandroluk/go/validator/schema/combinator"
	"github.com/leandroluk/go/validator/schema/date"
	"github.com/leandroluk/go/validator/schema/duration"
	"github.com/leandroluk/go/validator/schema/number"
	"github.com/leandroluk/go/validator/schema/object"
	"github.com/leandroluk/go/validator/schema/record"
	"github.com/leandroluk/go/validator/schema/text"
)

// Value represents an abstract value in the validation AST.
type Value = ast.Value

// Context holds the state of the current validation, including the path and issues found.
type Context = engine.Context

// Issue represents a single validation failure.
type Issue = issues.Issue

// ValidationError is an error type that aggregates multiple validation issues.
type ValidationError = issues.ValidationError

// Options configures the validator behavior.
type Options = schema.Options

// Option is a function that configures Options.
type Option = schema.Option

// Formatter formats validation error messages.
type Formatter = schema.Formatter

// AnySchema represents any type that can validate input.
type AnySchema = schema.AnySchema

// ObjectSchema validates structs or maps against defined fields.
type ObjectSchema[T any] = object.Schema[T]

// ArraySchema validates lists/slices.
type ArraySchema[E any] = array.Schema[E]

// RecordSchema validates maps with homogeneous value types.
type RecordSchema[V any] = record.Schema[V]

// TextSchema validates strings.
type TextSchema = text.Schema

// BooleanSchema validates booleans.
type BooleanSchema = boolean.Schema

// DateSchema validates time.Time values.
type DateSchema = date.Schema

// DurationSchema validates time.Duration values.
type DurationSchema = duration.Schema

// NumberSchema validates numeric values (int, float, etc).
type NumberSchema[N types.Number] = number.Schema[N]

const CodeOneOf = combinator.CodeOneOf

// CombinatorSchema combines multiple schemas.
type CombinatorSchema[T any] = combinator.Schema[T]

// AnyOfSchema succeeds if at least one schema succeeds.
type AnyOfSchema[T any] = combinator.AnyOfSchema[T]

// OneOfSchema succeeds if exactly one schema succeeds.
type OneOfSchema[T any] = combinator.OneOfSchema[T]

// WithFailFast stops validation on the first error.
func WithFailFast(value bool) Option {
	return schema.WithFailFast(value)
}

// WithMaxIssues limits the number of issues reported. Validation stops after reaching this limit.
func WithMaxIssues(value int) Option {
	return schema.WithMaxIssues(value)
}

// WithDefaultOnNull enables setting default values when input is explicit null.
func WithDefaultOnNull(value bool) Option {
	return schema.WithDefaultOnNull(value)
}

// WithCoerce enables automatic type coercion (e.g. string "123" to int 123).
func WithCoerce(value bool) Option {
	return schema.WithCoerce(value)
}

// WithOmitZero treats zero values (e.g. 0, "") as missing/undefined.
func WithOmitZero(value bool) Option {
	return schema.WithOmitZero(value)
}

// WithFormatter sets a custom error message formatter.
func WithFormatter(formatter Formatter) Option {
	return schema.WithFormatter(formatter)
}

// WithCoerceTrimSpace trims spaces from strings before validation/coercion.
func WithCoerceTrimSpace(value bool) Option {
	return schema.WithCoerceTrimSpace(value)
}

// WithCoerceNumberUnderscore allows underscores in number strings (e.g. "1_000").
func WithCoerceNumberUnderscore(value bool) Option {
	return schema.WithCoerceNumberUnderscore(value)
}

// WithCoerceDateUnixSeconds allows coercing unix timestamp (seconds) to Date.
func WithCoerceDateUnixSeconds(value bool) Option {
	return schema.WithCoerceDateUnixSeconds(value)
}

// WithCoerceDateUnixMilliseconds allows coercing unix timestamp (ms) to Date.
func WithCoerceDateUnixMilliseconds(value bool) Option {
	return schema.WithCoerceDateUnixMilliseconds(value)
}

// WithCoerceDurationSeconds allows coercing numeric seconds to Duration.
func WithCoerceDurationSeconds(value bool) Option {
	return schema.WithCoerceDurationSeconds(value)
}

// WithCoerceDurationMilliseconds allows coercing numeric milliseconds to Duration.
func WithCoerceDurationMilliseconds(value bool) Option {
	return schema.WithCoerceDurationMilliseconds(value)
}

// WithTimeLocation sets the time location for date parsing/formatting.
func WithTimeLocation(value *time.Location) Option {
	return schema.WithTimeLocation(value)
}

// WithDateLayouts overrides the default date parsing layouts.
func WithDateLayouts(layouts ...string) Option {
	return schema.WithDateLayouts(layouts...)
}

// WithAdditionalDateLayouts adds more date parsing layouts to the defaults.
func WithAdditionalDateLayouts(layouts ...string) Option {
	return schema.WithAdditionalDateLayouts(layouts...)
}

// Register registers a schema in the global registry for its output type.
func Register(schemaValue AnySchema) {
	registry.Register(schemaValue)
}

// ResetRegistry clears all registered schemas. Useful for testing.
func ResetRegistry() {
	registry.Reset()
}

// Object creates a new ObjectSchema for type T.
// The builder function is used to define fields and semantic rules on the schema.
func Object[T any](builder func(target *T, schemaValue *object.Schema[T])) *object.Schema[T] {
	schemaValue := object.New(builder)
	registry.Register(schemaValue)
	return schemaValue
}

// Array creates a new ArraySchema for elements of type E.
func Array[E any]() *array.Schema[E] {
	schemaValue := array.New[E]()
	registry.Register(schemaValue)
	return schemaValue
}

// Record creates a new RecordSchema for maps with values of type V.
func Record[V any]() *record.Schema[V] {
	schemaValue := record.New[V]()
	registry.Register(schemaValue)
	return schemaValue
}

// Text creates a new TextSchema for string validation.
func Text() *text.Schema {
	schemaValue := text.New()
	registry.Register(schemaValue)
	return schemaValue
}

// Boolean creates a new BooleanSchema.
func Boolean() *boolean.Schema {
	schemaValue := boolean.New()
	registry.Register(schemaValue)
	return schemaValue
}

// Date creates a new DateSchema for time.Time validation.
func Date() *date.Schema {
	schemaValue := date.New()
	registry.Register(schemaValue)
	return schemaValue
}

// Duration creates a new DurationSchema for time.Duration validation.
func Duration() *duration.Schema {
	schemaValue := duration.New()
	registry.Register(schemaValue)
	return schemaValue
}

// Number creates a new NumberSchema[N] for numeric validation.
// N can be any integer or float type.
func Number[N types.Number]() *number.Schema[N] {
	schemaValue := number.New[N]()
	registry.Register(schemaValue)
	return schemaValue
}

// AnyOf creates a combinator schema that succeeds if at least one of the provided schemas succeeds.
func AnyOf[T any](schemaList ...combinator.Schema[T]) *combinator.AnyOfSchema[T] {
	schemaValue := combinator.AnyOf(schemaList...)
	registry.Register(schemaValue)
	return schemaValue
}

// OneOf creates a combinator schema that succeeds if exactly one of the provided schemas succeeds.
func OneOf[T any](schemaList ...combinator.Schema[T]) *combinator.OneOfSchema[T] {
	schemaValue := combinator.OneOf(schemaList...)
	registry.Register(schemaValue)
	return schemaValue
}

// Validate validates the input against the registered schema for type T.
// Returns the validated/coerced value of type T or an error.
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

// Fluent API Types

// FieldBuilder is the entry point for defining field rules.
type FieldBuilder[T any] = object.FieldBuilder[T]
type TextFieldBuilder[T any] = object.TextFieldBuilder[T]
type NumberFieldBuilder[T any] = object.NumberFieldBuilder[T]
type BooleanFieldBuilder[T any] = object.BooleanFieldBuilder[T]
type DateFieldBuilder[T any] = object.DateFieldBuilder[T]
type DurationFieldBuilder[T any] = object.DurationFieldBuilder[T]

// Condition Operators

type ConditionOp = object.ConditionOp

const (
	Eq      ConditionOp = object.Eq
	Ne      ConditionOp = object.Ne
	Present ConditionOp = object.Present
	Missing ConditionOp = object.Missing
	Null    ConditionOp = object.Null
	NotNull ConditionOp = object.NotNull
)
