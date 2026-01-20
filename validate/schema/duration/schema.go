// schema/duration/schema.go
package duration

import (
	"reflect"
	"time"

	"github.com/leandroluk/gox/validate/internal/defaults"
	"github.com/leandroluk/gox/validate/internal/ruleset"
	"github.com/leandroluk/gox/validate/schema"
	"github.com/leandroluk/gox/validate/schema/duration/rule"
)

type Schema struct {
	required bool

	compareRules *ruleset.Set[time.Duration]
	boundRules   *ruleset.Set[time.Duration]

	defaultProvider defaults.Provider[time.Duration]
	rules           []ruleset.RuleFn[time.Duration]
}

func New() *Schema {
	return &Schema{
		compareRules:    ruleset.NewSet[time.Duration](),
		boundRules:      ruleset.NewSet[time.Duration](),
		defaultProvider: defaults.None[time.Duration](),
		rules:           make([]ruleset.RuleFn[time.Duration], 0),
	}
}

func (s *Schema) putCompare(ruleValue ruleset.Rule[time.Duration]) *Schema {
	s.compareRules.Put(ruleValue)
	return s
}

func (s *Schema) putBound(ruleValue ruleset.Rule[time.Duration]) *Schema {
	s.boundRules.Put(ruleValue)
	return s
}

func (s *Schema) Required() *Schema {
	s.required = true
	return s
}

func (s *Schema) Min(value time.Duration) *Schema {
	return s.putBound(rule.Min(CodeMin, value))
}

func (s *Schema) Max(value time.Duration) *Schema {
	return s.putBound(rule.Max(CodeMax, value))
}

func (s *Schema) Eq(value time.Duration) *Schema {
	return s.putCompare(rule.Eq(CodeEq, value))
}

func (s *Schema) Ne(value time.Duration) *Schema {
	return s.putCompare(rule.Ne(CodeNe, value))
}

func (s *Schema) Gt(value time.Duration) *Schema {
	return s.putCompare(rule.Gt(CodeGt, value))
}

func (s *Schema) Gte(value time.Duration) *Schema {
	return s.putCompare(rule.Gte(CodeGte, value))
}

func (s *Schema) Lt(value time.Duration) *Schema {
	return s.putCompare(rule.Lt(CodeLt, value))
}

func (s *Schema) Lte(value time.Duration) *Schema {
	return s.putCompare(rule.Lte(CodeLte, value))
}

func (s *Schema) Default(value time.Duration) *Schema {
	s.defaultProvider = defaults.Value(value)
	return s
}

func (s *Schema) DefaultFunc(fn func() time.Duration) *Schema {
	s.defaultProvider = defaults.Func(fn)
	return s
}

func (s *Schema) Custom(ruleValue ruleset.RuleFn[time.Duration]) *Schema {
	if ruleValue != nil {
		s.rules = append(s.rules, ruleValue)
	}
	return s
}

func (s *Schema) Validate(input any, optionList ...schema.Option) (time.Duration, error) {
	options := schema.ApplyOptions(optionList...)
	return s.validateWithOptions(input, options)
}

func (s *Schema) ValidateAny(input any, options schema.Options) (any, error) {
	out, err := s.validateWithOptions(input, options)
	return out, err
}

func (s *Schema) OutputType() reflect.Type {
	return reflect.TypeFor[time.Duration]()
}
