// schema/boolean/schema.go
package boolean

import (
	"reflect"

	"github.com/leandroluk/go/validate/internal/defaults"
	"github.com/leandroluk/go/validate/internal/ruleset"
	"github.com/leandroluk/go/validate/schema"
)

type Schema struct {
	required  bool
	isDefault bool

	defaultProvider defaults.Provider[bool]
	rules           []ruleset.RuleFn[bool]
}

func New() *Schema {
	return &Schema{
		defaultProvider: defaults.None[bool](),
		rules:           make([]ruleset.RuleFn[bool], 0),
	}
}

func (s *Schema) Required() *Schema {
	s.required = true
	return s
}

func (s *Schema) IsDefault() *Schema {
	s.isDefault = true
	return s
}

func (s *Schema) Default(value bool) *Schema {
	s.defaultProvider = defaults.Value(value)
	return s
}

func (s *Schema) DefaultFunc(fn func() bool) *Schema {
	s.defaultProvider = defaults.Func(fn)
	return s
}

func (s *Schema) Custom(ruleValue ruleset.RuleFn[bool]) *Schema {
	if ruleValue != nil {
		s.rules = append(s.rules, ruleValue)
	}
	return s
}

func (s *Schema) Validate(input any, optionList ...schema.Option) (bool, error) {
	options := schema.ApplyOptions(optionList...)
	return s.validateWithOptions(input, options)
}

func (s *Schema) ValidateAny(input any, options schema.Options) (any, error) {
	return s.validateWithOptions(input, options)
}

func (s *Schema) OutputType() reflect.Type {
	return reflect.TypeFor[bool]()
}
