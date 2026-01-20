// schema/duration/validate.go
package duration

import (
	"time"

	"github.com/leandroluk/gox/validate/internal/ast"
	"github.com/leandroluk/gox/validate/internal/defaults"
	"github.com/leandroluk/gox/validate/internal/engine"
	"github.com/leandroluk/gox/validate/internal/ruleset"
	"github.com/leandroluk/gox/validate/schema"
)

func (s *Schema) validateWithOptions(input any, options schema.Options) (time.Duration, error) {
	context := engine.NewContext(options)

	if durationValue, ok := input.(time.Duration); ok {
		output, _ := s.validateDurationValue(context, durationValue)
		return output, context.Error()
	}

	if durationPointer, ok := input.(*time.Duration); ok {
		if durationPointer == nil {
			output, _ := s.validateAST(context, ast.NullValue())
			return output, context.Error()
		}
		output, _ := s.validateDurationValue(context, *durationPointer)
		return output, context.Error()
	}

	if numeric, ok := parseGoNumericToInt64(input); ok {
		if options.Coerce && (options.CoerceDurationSeconds || options.CoerceDurationMilliseconds) {
			output := durationFromInt64(options, numeric)
			_, _ = s.validateDurationValue(context, output)
			return output, context.Error()
		}

		context.AddIssueWithMeta(CodeInvalid, "invalid duration", map[string]any{
			"value": input,
		})
		var zero time.Duration
		return zero, context.Error()
	}

	value, err := engine.InputToASTWithOptions(input, options)
	if err != nil {
		var zero time.Duration
		return zero, err
	}

	output, _ := s.validateAST(context, value)
	return output, context.Error()
}

func (s *Schema) validateAST(context *engine.Context, value ast.Value) (time.Duration, bool) {
	if defaultValue, ok := defaults.Apply(value.Presence, context.Options, s.defaultProvider); ok {
		return defaultValue, false
	}

	if value.IsMissing() || value.IsNull() {
		if s.required {
			stop := context.AddIssue(CodeRequired, "required")
			var zero time.Duration
			return zero, stop
		}
		var zero time.Duration
		return zero, false
	}

	var parsed time.Duration
	var ok bool

	switch value.Kind {
	case ast.KindString:
		parsed, ok = parseDurationWithOptions(context.Options, value.String)

	case ast.KindNumber:
		parsed, ok = parseNanosecondsTextWithOptions(context.Options, value.Number)

	default:
		ok = false
	}

	if !ok {
		if value.Kind != ast.KindString && value.Kind != ast.KindNumber {
			stop := context.AddIssueWithMeta(CodeType, "expected duration", map[string]any{
				"expected": "duration",
				"actual":   value.Kind.String(),
			})
			var zero time.Duration
			return zero, stop
		}

		stop := context.AddIssueWithMeta(CodeInvalid, "invalid duration", map[string]any{
			"value": func() any {
				if value.Kind == ast.KindString {
					return value.String
				}
				if value.Kind == ast.KindNumber {
					return value.Number
				}
				return nil
			}(),
		})

		var zero time.Duration
		return zero, stop
	}

	return s.validateDurationValue(context, parsed)
}

func (s *Schema) validateDurationValue(context *engine.Context, parsed time.Duration) (time.Duration, bool) {
	var stopped bool

	parsed, stopped = s.compareRules.ApplyAll(parsed, context)
	if stopped {
		return parsed, true
	}

	parsed, stopped = s.boundRules.ApplyAll(parsed, context)
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
