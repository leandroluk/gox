// schema/date/schema.go
package date

import (
	"reflect"
	"time"

	"github.com/leandroluk/go/validate/internal/defaults"
	"github.com/leandroluk/go/validate/internal/ruleset"
	"github.com/leandroluk/go/validate/schema"
	"github.com/leandroluk/go/validate/schema/date/rule"
)

type Schema struct {
	required bool

	dateTimeOnly bool

	compareRules *ruleset.Set[time.Time]
	boundRules   *ruleset.Set[time.Time]

	defaultProvider defaults.Provider[time.Time]

	rules []ruleset.RuleFn[time.Time]
}

func New() *Schema {
	return &Schema{
		compareRules:    ruleset.NewSet[time.Time](),
		boundRules:      ruleset.NewSet[time.Time](),
		defaultProvider: defaults.None[time.Time](),
		rules:           make([]ruleset.RuleFn[time.Time], 0),
	}
}

func (s *Schema) Required() *Schema {
	s.required = true
	return s
}

func (s *Schema) DateTime() *Schema {
	s.dateTimeOnly = true
	return s
}

func (s *Schema) Datetime() *Schema {
	return s.DateTime()
}

func (s *Schema) Min(value time.Time) *Schema {
	s.boundRules.Put(rule.Min(CodeMin, value))
	return s
}

func (s *Schema) Max(value time.Time) *Schema {
	s.boundRules.Put(rule.Max(CodeMax, value))
	return s
}

func (s *Schema) Eq(value time.Time) *Schema {
	s.compareRules.Put(rule.Eq(CodeEq, value))
	return s
}

func (s *Schema) Ne(value time.Time) *Schema {
	s.compareRules.Put(rule.Ne(CodeNe, value))
	return s
}

func (s *Schema) Gt(value time.Time) *Schema {
	s.compareRules.Put(rule.Gt(CodeGt, value))
	return s
}

func (s *Schema) Gte(value time.Time) *Schema {
	s.compareRules.Put(rule.Gte(CodeGte, value))
	return s
}

func (s *Schema) Lt(value time.Time) *Schema {
	s.compareRules.Put(rule.Lt(CodeLt, value))
	return s
}

func (s *Schema) Lte(value time.Time) *Schema {
	s.compareRules.Put(rule.Lte(CodeLte, value))
	return s
}

func (s *Schema) Default(value time.Time) *Schema {
	s.defaultProvider = defaults.Value(value)
	return s
}

func (s *Schema) DefaultFunc(fn func() time.Time) *Schema {
	s.defaultProvider = defaults.Func(fn)
	return s
}

func (s *Schema) Custom(ruleValue ruleset.RuleFn[time.Time]) *Schema {
	if ruleValue != nil {
		s.rules = append(s.rules, ruleValue)
	}
	return s
}

func (s *Schema) Validate(input any, optionList ...schema.Option) (time.Time, error) {
	options := schema.ApplyOptions(optionList...)
	return s.validateWithOptions(input, options)
}

func (s *Schema) ValidateAny(input any, options schema.Options) (any, error) {
	return s.validateWithOptions(input, options)
}

func (s *Schema) OutputType() reflect.Type {
	return reflect.TypeFor[time.Time]()
}
