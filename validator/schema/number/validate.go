// schema/number/validate.go
package number

import (
	"strings"

	"github.com/leandroluk/go/validator/internal/ast"
	"github.com/leandroluk/go/validator/internal/defaults"
	"github.com/leandroluk/go/validator/internal/engine"
	"github.com/leandroluk/go/validator/internal/reflection"
	"github.com/leandroluk/go/validator/internal/ruleset"
	"github.com/leandroluk/go/validator/schema"
)

func (s *Schema[N]) validateWithOptions(input any, options schema.Options) (N, error) {
	context := engine.NewContext(options)

	if parsed, ok := readDirectNumber[N](input); ok {
		output, _ := s.validateNumberValue(context, parsed)
		return output, context.Error()
	}

	value, err := engine.InputToASTWithOptions(input, options)
	if err != nil {
		var zero N
		return zero, err
	}

	output, _ := s.validateAST(context, value)
	return output, context.Error()
}

func (s *Schema[N]) validateAST(context *engine.Context, value ast.Value) (N, bool) {
	if defaultValue, ok := defaults.Apply(value.Presence, context.Options, s.defaultProvider); ok {
		return defaultValue, false
	}

	if value.IsMissing() || value.IsNull() {
		if s.required {
			stop := context.AddIssue(CodeRequired, "required")
			var zero N
			return zero, stop
		}
		var zero N
		return zero, false
	}

	var parsed N
	var ok bool

	switch value.Kind {
	case ast.KindNumber:
		text := value.Number

		if !context.Options.CoerceNumberUnderscore && strings.IndexByte(text, '_') >= 0 {
			ok = false
			break
		}
		if context.Options.CoerceNumberUnderscore {
			text = removeUnderscore(text)
		}

		parsed, ok = parseTo[N](text)

	case ast.KindString:
		if context.Options.Coerce {
			text := value.String

			if !context.Options.CoerceTrimSpace && strings.TrimSpace(text) != text {
				ok = false
				break
			}
			if context.Options.CoerceTrimSpace {
				text = strings.TrimSpace(text)
			}

			if !context.Options.CoerceNumberUnderscore && strings.IndexByte(text, '_') >= 0 {
				ok = false
				break
			}
			if context.Options.CoerceNumberUnderscore {
				text = removeUnderscore(text)
			}

			parsed, ok = parseTo[N](text)
		}

	default:
		ok = false
	}

	if !ok {
		if value.Kind != ast.KindNumber && value.Kind != ast.KindString {
			stop := context.AddIssueWithMeta(CodeType, "expected number", map[string]any{
				"expected": "number",
				"actual":   value.Kind.String(),
			})
			var zero N
			return zero, stop
		}

		stop := context.AddIssueWithMeta(CodeInvalid, "invalid number", map[string]any{
			"value": func() any {
				if value.Kind == ast.KindNumber {
					return value.Number
				}
				if value.Kind == ast.KindString {
					return value.String
				}
				return nil
			}(),
		})
		var zero N
		return zero, stop
	}

	return s.validateNumberValue(context, parsed)
}

func (s *Schema[N]) validateNumberValue(context *engine.Context, parsed N) (N, bool) {
	if s.isDefault && reflection.IsDefault(parsed) {
		return parsed, false
	}

	var stopped bool

	parsed, stopped = s.valueRules.ApplyAll(parsed, context)
	if stopped {
		return parsed, true
	}

	parsed, stopped = s.boundRules.ApplyAll(parsed, context)
	if stopped {
		return parsed, true
	}

	parsed, stopped = s.compareRules.ApplyAll(parsed, context)
	if stopped {
		return parsed, true
	}

	if len(s.rules) > 0 {
		if ruleset.Apply(parsed, context, s.rules...) {
			return parsed, true
		}
	}

	return parsed, false
}
