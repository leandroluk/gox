// schema/array/schema.go
package array

import (
	"reflect"

	"github.com/leandroluk/gox/validate/internal/ast"
	"github.com/leandroluk/gox/validate/internal/defaults"
	"github.com/leandroluk/gox/validate/internal/engine"
	"github.com/leandroluk/gox/validate/internal/ruleset"
	"github.com/leandroluk/gox/validate/schema"
	"github.com/leandroluk/gox/validate/schema/array/rule"
)

type ItemValidator[E any] func(context *engine.Context, value ast.Value) (E, bool)

type Schema[E any] struct {
	required bool

	lengthRules *ruleset.Set[int]

	defaultProvider defaults.Provider[[]E]

	uniqueEnabled bool
	uniqueHash    func(value E) string
	uniqueEqual   func(left E, right E) bool

	itemValidator ItemValidator[E]
	rules         []ruleset.RuleFn[[]E]
}

func New[E any]() *Schema[E] {
	return &Schema[E]{
		lengthRules:     ruleset.NewSet[int](),
		defaultProvider: defaults.None[[]E](),
		rules:           make([]ruleset.RuleFn[[]E], 0),
	}
}

func (s *Schema[E]) putLength(ruleValue ruleset.Rule[int]) *Schema[E] {
	s.lengthRules.Put(ruleValue)
	return s
}

func (s *Schema[E]) Required() *Schema[E] {
	s.required = true
	return s
}

func (s *Schema[E]) Min(length int) *Schema[E] {
	if normalized, ok := normalizeLimit(length); ok {
		return s.putLength(rule.Min(CodeMin, normalized))
	}
	return s
}

func (s *Schema[E]) Max(length int) *Schema[E] {
	if normalized, ok := normalizeLimit(length); ok {
		return s.putLength(rule.Max(CodeMax, normalized))
	}
	return s
}

func (s *Schema[E]) Len(length int) *Schema[E] {
	if normalized, ok := normalizeLimit(length); ok {
		return s.putLength(rule.Len(CodeLen, normalized))
	}
	return s
}

func (s *Schema[E]) Eq(length int) *Schema[E] {
	if normalized, ok := normalizeLimit(length); ok {
		return s.putLength(rule.Eq(CodeEq, normalized))
	}
	return s
}

func (s *Schema[E]) Ne(length int) *Schema[E] {
	if normalized, ok := normalizeLimit(length); ok {
		return s.putLength(rule.Ne(CodeNe, normalized))
	}
	return s
}

func (s *Schema[E]) Gt(length int) *Schema[E] {
	if normalized, ok := normalizeLimit(length); ok {
		return s.putLength(rule.Gt(CodeGt, normalized))
	}
	return s
}

func (s *Schema[E]) Gte(length int) *Schema[E] {
	if normalized, ok := normalizeLimit(length); ok {
		return s.putLength(rule.Gte(CodeGte, normalized))
	}
	return s
}

func (s *Schema[E]) Lt(length int) *Schema[E] {
	if normalized, ok := normalizeLimit(length); ok {
		return s.putLength(rule.Lt(CodeLt, normalized))
	}
	return s
}

func (s *Schema[E]) Lte(length int) *Schema[E] {
	if normalized, ok := normalizeLimit(length); ok {
		return s.putLength(rule.Lte(CodeLte, normalized))
	}
	return s
}

func (s *Schema[E]) Unique() *Schema[E] {
	s.uniqueEnabled = true
	s.uniqueHash = nil
	s.uniqueEqual = nil
	return s
}

func (s *Schema[E]) UniqueByHash(hash func(value E) string) *Schema[E] {
	s.uniqueEnabled = true
	s.uniqueHash = hash
	s.uniqueEqual = nil
	return s
}

func (s *Schema[E]) UniqueByEqual(equal func(left E, right E) bool) *Schema[E] {
	s.uniqueEnabled = true
	s.uniqueEqual = equal
	s.uniqueHash = nil
	return s
}

func (s *Schema[E]) Default(value []E) *Schema[E] {
	s.defaultProvider = defaults.Value(value)
	return s
}

func (s *Schema[E]) DefaultFunc(fn func() []E) *Schema[E] {
	s.defaultProvider = defaults.Func(fn)
	return s
}

func (s *Schema[E]) Items(validator ItemValidator[E]) *Schema[E] {
	s.itemValidator = validator
	return s
}

func (s *Schema[E]) Custom(ruleValue ruleset.RuleFn[[]E]) *Schema[E] {
	if ruleValue != nil {
		s.rules = append(s.rules, ruleValue)
	}
	return s
}

func (s *Schema[E]) Validate(input any, optionList ...schema.Option) ([]E, error) {
	options := schema.ApplyOptions(optionList...)
	return s.validateWithOptions(input, options)
}

func (s *Schema[E]) ValidateAny(input any, options schema.Options) (any, error) {
	return s.validateWithOptions(input, options)
}

func (s *Schema[E]) OutputType() reflect.Type {
	return reflect.TypeOf((*[]E)(nil)).Elem()
}
