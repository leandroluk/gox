// schema/boolean/boolean.go
package boolean

import (
	"reflect"
	"strconv"

	"github.com/leandroluk/gox/validate/internal/ast"
	"github.com/leandroluk/gox/validate/internal/defaults"
	"github.com/leandroluk/gox/validate/internal/engine"
	"github.com/leandroluk/gox/validate/internal/reflection"
	"github.com/leandroluk/gox/validate/internal/ruleset"
	"github.com/leandroluk/gox/validate/schema"
)

const (
	CodeRequired = "boolean.required"
	CodeType     = "boolean.type"
)

func parseBooleanWithOptions(options schema.Options, value ast.Value) (bool, bool) {
	if value.Kind == ast.KindBoolean {
		return value.Boolean, true
	}

	if !options.Coerce {
		return false, false
	}

	switch value.Kind {
	case ast.KindString:
		parsed, err := strconv.ParseBool(value.String)
		if err != nil {
			return false, false
		}
		return parsed, true

	case ast.KindNumber:
		if value.Number == "0" {
			return false, true
		}
		if value.Number == "1" {
			return true, true
		}
		return false, false

	default:
		return false, false
	}
}

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

func (s *Schema) validateWithOptions(input any, options schema.Options) (bool, error) {
	context := engine.NewContext(options)

	value, err := engine.InputToASTWithOptions(input, options)
	if err != nil {
		return false, err
	}

	output, _ := s.validateAST(context, value)
	return output, context.Error()
}

func (s *Schema) validateAST(context *engine.Context, value ast.Value) (bool, bool) {
	if defaultValue, ok := defaults.Apply(value.Presence, context.Options, s.defaultProvider); ok {
		return defaultValue, false
	}

	if value.IsMissing() || value.IsNull() {
		if s.required {
			stop := context.AddIssue(CodeRequired, "required")
			return false, stop
		}
		return false, false
	}

	output, ok := parseBooleanWithOptions(context.Options, value)
	if !ok {
		stop := context.AddIssue(CodeType, "expected boolean", map[string]any{
			"expected": "boolean",
			"actual":   value.Kind.String(),
		})
		return false, stop
	}

	if s.isDefault && reflection.IsDefault(output) {
		return output, false
	}

	if len(s.rules) > 0 {
		if ruleset.Apply(output, context, s.rules...) {
			return output, true
		}
	}

	return output, false
}
