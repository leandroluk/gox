// schema/boolean/validate.go
package boolean

import (
	"github.com/leandroluk/go/validator/internal/ast"
	"github.com/leandroluk/go/validator/internal/defaults"
	"github.com/leandroluk/go/validator/internal/engine"
	"github.com/leandroluk/go/validator/internal/reflection"
	"github.com/leandroluk/go/validator/internal/ruleset"
	"github.com/leandroluk/go/validator/schema"
)

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
		stop := context.AddIssueWithMeta(CodeType, "expected boolean", map[string]any{
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
