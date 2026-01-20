// schema/record/schema.go
package record

import (
	"reflect"

	"github.com/leandroluk/go/validate/internal/ast"
	"github.com/leandroluk/go/validate/internal/defaults"
	"github.com/leandroluk/go/validate/internal/engine"
	"github.com/leandroluk/go/validate/internal/ruleset"
	"github.com/leandroluk/go/validate/schema"
	"github.com/leandroluk/go/validate/schema/record/rule"
)

type KeysFunc func(context *engine.Context, key string) bool

type ValuesFunc[V any] func(context *engine.Context, value ast.Value) (V, bool)

type Schema[V any] struct {
	required  bool
	isDefault bool

	unique bool

	keySchema schema.AnySchema
	keyFunc   KeysFunc

	valueSchema schema.AnySchema
	valueFunc   ValuesFunc[V]

	defaultProvider defaults.Provider[map[string]V]

	lengthRules *ruleset.Set[int]
}

type RecordSchema[V any] = Schema[V]

func New[V any]() *Schema[V] {
	return &Schema[V]{
		defaultProvider: defaults.None[map[string]V](),
		lengthRules:     ruleset.NewSet[int](),
	}
}

func (schemaValue *Schema[V]) Required() *Schema[V] {
	schemaValue.required = true
	return schemaValue
}

func (schemaValue *Schema[V]) IsDefault() *Schema[V] {
	schemaValue.isDefault = true
	return schemaValue
}

func (schemaValue *Schema[V]) Len(value int) *Schema[V] {
	schemaValue.lengthRules.Put(rule.Len(CodeLen, value))
	return schemaValue
}

func (schemaValue *Schema[V]) Min(value int) *Schema[V] {
	schemaValue.lengthRules.Put(rule.Min(CodeMin, value))
	return schemaValue
}

func (schemaValue *Schema[V]) Max(value int) *Schema[V] {
	schemaValue.lengthRules.Put(rule.Max(CodeMax, value))
	return schemaValue
}

func (schemaValue *Schema[V]) Eq(value int) *Schema[V] {
	schemaValue.lengthRules.Put(rule.Eq(CodeEq, value))
	return schemaValue
}

func (schemaValue *Schema[V]) Ne(value int) *Schema[V] {
	schemaValue.lengthRules.Put(rule.Ne(CodeNe, value))
	return schemaValue
}

func (schemaValue *Schema[V]) Gt(value int) *Schema[V] {
	schemaValue.lengthRules.Put(rule.Gt(CodeGt, value))
	return schemaValue
}

func (schemaValue *Schema[V]) Gte(value int) *Schema[V] {
	schemaValue.lengthRules.Put(rule.Gte(CodeGte, value))
	return schemaValue
}

func (schemaValue *Schema[V]) Lt(value int) *Schema[V] {
	schemaValue.lengthRules.Put(rule.Lt(CodeLt, value))
	return schemaValue
}

func (schemaValue *Schema[V]) Lte(value int) *Schema[V] {
	schemaValue.lengthRules.Put(rule.Lte(CodeLte, value))
	return schemaValue
}

func (schemaValue *Schema[V]) Unique() *Schema[V] {
	schemaValue.unique = true
	return schemaValue
}

func (schemaValue *Schema[V]) Keys(schemaValueKey schema.AnySchema) *Schema[V] {
	schemaValue.keySchema = schemaValueKey
	schemaValue.keyFunc = nil
	return schemaValue
}

func (schemaValue *Schema[V]) KeysFunc(fn KeysFunc) *Schema[V] {
	schemaValue.keyFunc = fn
	schemaValue.keySchema = nil
	return schemaValue
}

func (schemaValue *Schema[V]) EndKeys() *Schema[V] {
	return schemaValue
}

func (schemaValue *Schema[V]) Dive(schemaValueItem schema.AnySchema) *Schema[V] {
	schemaValue.valueSchema = schemaValueItem
	schemaValue.valueFunc = nil
	return schemaValue
}

func (schemaValue *Schema[V]) Values(fn func(context *engine.Context, value ast.Value) (V, bool)) *Schema[V] {
	schemaValue.valueFunc = fn
	schemaValue.valueSchema = nil
	return schemaValue
}

func (schemaValue *Schema[V]) Default(value map[string]V) *Schema[V] {
	schemaValue.defaultProvider = defaults.Value(value)
	return schemaValue
}

func (schemaValue *Schema[V]) DefaultFunc(fn func() map[string]V) *Schema[V] {
	schemaValue.defaultProvider = defaults.Func(fn)
	return schemaValue
}

func (schemaValue *Schema[V]) Validate(input any, optionList ...schema.Option) (map[string]V, error) {
	options := schema.ApplyOptions(optionList...)
	return schemaValue.validateWithOptions(input, options)
}

func (schemaValue *Schema[V]) ValidateAny(input any, options schema.Options) (any, error) {
	return schemaValue.validateWithOptions(input, options)
}

func (schemaValue *Schema[V]) OutputType() reflect.Type {
	return reflect.TypeFor[map[string]V]()
}
