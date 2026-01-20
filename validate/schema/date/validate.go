// schema/date/validate.go
package date

import (
	"strings"
	"time"

	"github.com/leandroluk/go/validate/internal/ast"
	"github.com/leandroluk/go/validate/internal/defaults"
	"github.com/leandroluk/go/validate/internal/engine"
	"github.com/leandroluk/go/validate/internal/ruleset"
	"github.com/leandroluk/go/validate/schema"
)

func (s *Schema) validateWithOptions(input any, options schema.Options) (time.Time, error) {
	context := engine.NewContext(options)

	value, err := engine.InputToASTWithOptions(input, options)
	if err != nil {
		var zero time.Time
		return zero, err
	}

	output, _ := s.validateAST(context, value)
	return output, context.Error()
}

func (s *Schema) validateAST(context *engine.Context, value ast.Value) (time.Time, bool) {
	if defaultValue, ok := defaults.Apply(value.Presence, context.Options, s.defaultProvider); ok {
		return defaultValue, false
	}

	if value.IsMissing() || value.IsNull() {
		if s.required {
			stop := context.AddIssue(CodeRequired, "required")
			var zero time.Time
			return zero, stop
		}
		var zero time.Time
		return zero, false
	}

	var parsed time.Time
	var matchedLayout string
	var ok bool

	switch value.Kind {
	case ast.KindString:
		text := value.String
		if context.Options.CoerceTrimSpace {
			text = strings.TrimSpace(text)
		}

		parsed, matchedLayout, ok = parseDate(context.Options, text)

		if !ok && context.Options.Coerce && (context.Options.CoerceDateUnixSeconds || context.Options.CoerceDateUnixMilliseconds) {
			parsed, ok = parseUnixTextWithOptions(context.Options, text)
			if ok {
				matchedLayout = "unix"
			}
		}

	case ast.KindNumber:
		if context.Options.Coerce {
			parsed, ok = parseUnixNumberWithOptions(context.Options, value.Number)
			if ok {
				matchedLayout = "unix"
			}
		}

	default:
		ok = false
	}

	if !ok {
		if value.Kind != ast.KindString && value.Kind != ast.KindNumber {
			stop := context.AddIssueWithMeta(CodeType, "expected date", map[string]any{
				"expected": "date",
				"actual":   value.Kind.String(),
			})
			var zero time.Time
			return zero, stop
		}

		stop := context.AddIssueWithMeta(CodeInvalid, "invalid date", map[string]any{
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
		var zero time.Time
		return zero, stop
	}

	if s.dateTimeOnly && matchedLayout != "unix" && !layoutHasClock(matchedLayout) {
		stop := context.AddIssueWithMeta(CodeDateTime, "expected datetime", map[string]any{
			"value":  value.String,
			"layout": matchedLayout,
		})
		return parsed, stop
	}

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
