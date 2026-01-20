// schema/number/schema.go
package number

import (
	"reflect"

	"github.com/leandroluk/go/validate/internal/defaults"
	"github.com/leandroluk/go/validate/internal/ruleset"
	"github.com/leandroluk/go/validate/internal/types"
	"github.com/leandroluk/go/validate/schema"
	"github.com/leandroluk/go/validate/schema/number/rule"
)

type Schema[N types.Number] struct {
	required  bool
	isDefault bool

	valueRules   *ruleset.Set[N]
	boundRules   *ruleset.Set[N]
	compareRules *ruleset.Set[N]

	defaultProvider defaults.Provider[N]
	rules           []ruleset.RuleFn[N]
}

func New[N types.Number]() *Schema[N] {
	return &Schema[N]{
		valueRules:      ruleset.NewSet[N](),
		boundRules:      ruleset.NewSet[N](),
		compareRules:    ruleset.NewSet[N](),
		defaultProvider: defaults.None[N](),
		rules:           make([]ruleset.RuleFn[N], 0),
	}
}

func (s *Schema[N]) putValue(ruleValue ruleset.Rule[N]) *Schema[N] {
	s.valueRules.Put(ruleValue)
	return s
}

func (s *Schema[N]) putBound(ruleValue ruleset.Rule[N]) *Schema[N] {
	s.boundRules.Put(ruleValue)
	return s
}

func (s *Schema[N]) putCompare(ruleValue ruleset.Rule[N]) *Schema[N] {
	s.compareRules.Put(ruleValue)
	return s
}

func (s *Schema[N]) Required() *Schema[N] {
	s.required = true
	return s
}

func (s *Schema[N]) IsDefault() *Schema[N] {
	s.isDefault = true
	return s
}

func (s *Schema[N]) Min(value N) *Schema[N] {
	return s.putBound(rule.Min(CodeMin, value))
}

func (s *Schema[N]) Max(value N) *Schema[N] {
	return s.putBound(rule.Max(CodeMax, value))
}

func (s *Schema[N]) Eq(value N) *Schema[N] {
	return s.putCompare(rule.Eq(CodeEq, value))
}

func (s *Schema[N]) Ne(value N) *Schema[N] {
	return s.putCompare(rule.Ne(CodeNe, value))
}

func (s *Schema[N]) Gt(value N) *Schema[N] {
	return s.putCompare(rule.Gt(CodeGt, value))
}

func (s *Schema[N]) Gte(value N) *Schema[N] {
	return s.putCompare(rule.Gte(CodeGte, value))
}

func (s *Schema[N]) Lt(value N) *Schema[N] {
	return s.putCompare(rule.Lt(CodeLt, value))
}

func (s *Schema[N]) Lte(value N) *Schema[N] {
	return s.putCompare(rule.Lte(CodeLte, value))
}

func (s *Schema[N]) OneOf(values ...N) *Schema[N] {
	if len(values) == 0 {
		s.valueRules.Remove("oneof")
		return s
	}
	return s.putValue(rule.OneOf(CodeOneOf, values...))
}

func (s *Schema[N]) Default(value N) *Schema[N] {
	s.defaultProvider = defaults.Value(value)
	return s
}

func (s *Schema[N]) DefaultFunc(fn func() N) *Schema[N] {
	s.defaultProvider = defaults.Func(fn)
	return s
}

func (s *Schema[N]) Custom(ruleValue ruleset.RuleFn[N]) *Schema[N] {
	if ruleValue != nil {
		s.rules = append(s.rules, ruleValue)
	}
	return s
}

func (s *Schema[N]) Validate(input any, optionList ...schema.Option) (N, error) {
	options := schema.ApplyOptions(optionList...)
	return s.validateWithOptions(input, options)
}

func (s *Schema[N]) ValidateAny(input any, options schema.Options) (any, error) {
	return s.validateWithOptions(input, options)
}

func (s *Schema[N]) OutputType() reflect.Type {
	return reflect.TypeFor[N]()
}
