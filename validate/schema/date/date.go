// schema/date/date.go
package date

import (
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/leandroluk/gox/validate/internal/ast"
	"github.com/leandroluk/gox/validate/internal/defaults"
	"github.com/leandroluk/gox/validate/internal/engine"
	"github.com/leandroluk/gox/validate/internal/ruleset"
	"github.com/leandroluk/gox/validate/internal/types"
	"github.com/leandroluk/gox/validate/schema"
)

const (
	CodeRequired = "date.required"
	CodeType     = "date.type"
	CodeInvalid  = "date.invalid"
	CodeMin      = "date.min"
	CodeMax      = "date.max"

	CodeDateTime = "date.datetime"

	CodeEq  = "date.eq"
	CodeNe  = "date.ne"
	CodeGt  = "date.gt"
	CodeGte = "date.gte"
	CodeLt  = "date.lt"
	CodeLte = "date.lte"
)

type ruleMap[T any] struct {
	Eq  T
	Gt  T
	Gte T
	Lt  T
	Lte T
	Max T
	Min T
	Ne  T
}

var Msg = ruleMap[string]{
	Eq:  "must be equal",
	Gt:  "must be greater",
	Gte: "must be greater or equal",
	Lt:  "must be lower",
	Lte: "must be lower or equal",
	Max: "too late",
	Min: "too early",
	Ne:  "must not be equal",
}

var Rule = ruleMap[func(code string, expected time.Time) ruleset.Rule[time.Time]]{
	Eq: func(code string, expected time.Time) ruleset.Rule[time.Time] {
		return ruleset.New("", func(actual time.Time, context *engine.Context) (time.Time, bool) {
			if actual.Equal(expected) {
				return actual, false
			}
			stop := context.AddIssue(code, Msg.Eq, types.AnyMap{"expected": fmt(expected), "actual": fmt(actual)})
			return actual, stop
		})
	},
	Gt: func(code string, expected time.Time) ruleset.Rule[time.Time] {
		return ruleset.New("", func(actual time.Time, context *engine.Context) (time.Time, bool) {
			if actual.After(expected) {
				return actual, false
			}
			stop := context.AddIssue(code, Msg.Gt, types.AnyMap{"expected": fmt(expected), "actual": fmt(actual)})
			return actual, stop
		})
	},
	Gte: func(code string, expected time.Time) ruleset.Rule[time.Time] {
		return ruleset.New("", func(actual time.Time, context *engine.Context) (time.Time, bool) {
			if actual.After(expected) || actual.Equal(expected) {
				return actual, false
			}
			stop := context.AddIssue(code, Msg.Gte, types.AnyMap{"expected": fmt(expected), "actual": fmt(actual)})
			return actual, stop
		})
	},
	Lt: func(code string, expected time.Time) ruleset.Rule[time.Time] {
		return ruleset.New("", func(actual time.Time, context *engine.Context) (time.Time, bool) {
			if actual.Before(expected) {
				return actual, false
			}
			stop := context.AddIssue(code, Msg.Lt, types.AnyMap{"expected": fmt(expected), "actual": fmt(actual)})
			return actual, stop
		})
	},
	Lte: func(code string, expected time.Time) ruleset.Rule[time.Time] {
		return ruleset.New("", func(actual time.Time, context *engine.Context) (time.Time, bool) {
			if actual.Before(expected) || actual.Equal(expected) {
				return actual, false
			}
			stop := context.AddIssue(code, Msg.Lte, types.AnyMap{"expected": fmt(expected), "actual": fmt(actual)})
			return actual, stop
		})
	},
	Max: func(code string, max time.Time) ruleset.Rule[time.Time] {
		return ruleset.New("max", func(actual time.Time, context *engine.Context) (time.Time, bool) {
			if actual.After(max) {
				stop := context.AddIssue(code, Msg.Max, types.AnyMap{"max": fmt(max), "actual": fmt(actual)})
				return actual, stop
			}
			return actual, false
		})
	},
	Min: func(code string, min time.Time) ruleset.Rule[time.Time] {
		return ruleset.New("min", func(actual time.Time, context *engine.Context) (time.Time, bool) {
			if actual.Before(min) {
				stop := context.AddIssue(code, Msg.Min, types.AnyMap{"min": fmt(min), "actual": fmt(actual)})
				return actual, stop
			}
			return actual, false
		})
	},
	Ne: func(code string, expected time.Time) ruleset.Rule[time.Time] {
		return ruleset.New("", func(actual time.Time, context *engine.Context) (time.Time, bool) {
			if !actual.Equal(expected) {
				return actual, false
			}
			stop := context.AddIssue(code, Msg.Ne, types.AnyMap{"expected": fmt(expected), "actual": fmt(actual)})
			return actual, stop
		})
	},
}

func fmt(t time.Time) string { return t.Format(time.RFC3339Nano) }

func parseDate(options schema.Options, input string) (time.Time, string, bool) {
	location := options.TimeLocation
	if location == nil {
		location = time.UTC
	}

	for _, layout := range options.DateLayouts {
		if layout == "" {
			continue
		}
		if parsed, err := time.ParseInLocation(layout, input, location); err == nil {
			return parsed, layout, true
		}
	}

	return time.Time{}, "", false
}

func parseUnixTextWithOptions(options schema.Options, text string) (time.Time, bool) {
	if options.CoerceTrimSpace {
		text = strings.TrimSpace(text)
	}
	if options.CoerceNumberUnderscore {
		text = removeDateUnderscore(text)
	}

	value, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return time.Time{}, false
	}

	return unixFromInt64(options, value), true
}

func parseUnixNumberWithOptions(options schema.Options, text string) (time.Time, bool) {
	value, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return time.Time{}, false
	}
	return unixFromInt64(options, value), true
}

func unixFromInt64(options schema.Options, value int64) time.Time {
	secondsFlag := options.CoerceDateUnixSeconds
	millisFlag := options.CoerceDateUnixMilliseconds

	if !secondsFlag && !millisFlag {
		return time.Unix(value, 0).UTC()
	}

	if secondsFlag && !millisFlag {
		return time.Unix(value, 0).UTC()
	}

	if millisFlag && !secondsFlag {
		return unixMillis(value)
	}

	abs := value
	if abs < 0 {
		abs = -abs
	}

	if abs >= 100_000_000_000 {
		return unixMillis(value)
	}

	return time.Unix(value, 0).UTC()
}

func unixMillis(value int64) time.Time {
	seconds := value / 1000
	millisRemainder := value % 1000
	if millisRemainder < 0 {
		millisRemainder = -millisRemainder
	}
	return time.Unix(seconds, millisRemainder*1_000_000).UTC()
}

func removeDateUnderscore(input string) string {
	if strings.IndexByte(input, '_') < 0 {
		return input
	}

	var builder strings.Builder
	builder.Grow(len(input))

	for index := 0; index < len(input); index++ {
		ch := input[index]
		if ch == '_' {
			continue
		}
		builder.WriteByte(ch)
	}

	return builder.String()
}

func layoutHasClock(layout string) bool {
	if strings.Contains(layout, "15") {
		return true
	}
	if strings.Contains(layout, "03") {
		return true
	}
	if strings.Contains(layout, "04") {
		return true
	}
	if strings.Contains(layout, "05") {
		return true
	}
	if strings.Contains(layout, "PM") || strings.Contains(layout, "pm") {
		return true
	}
	return false
}

type Schema struct {
	required  bool
	isDefault bool

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

func (s *Schema) IsDefault() *Schema {
	s.isDefault = true
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
	s.boundRules.Put(Rule.Min(CodeMin, value))
	return s
}

func (s *Schema) Max(value time.Time) *Schema {
	s.boundRules.Put(Rule.Max(CodeMax, value))
	return s
}

func (s *Schema) Eq(value time.Time) *Schema {
	s.compareRules.Put(Rule.Eq(CodeEq, value))
	return s
}

func (s *Schema) Ne(value time.Time) *Schema {
	s.compareRules.Put(Rule.Ne(CodeNe, value))
	return s
}

func (s *Schema) Gt(value time.Time) *Schema {
	s.compareRules.Put(Rule.Gt(CodeGt, value))
	return s
}

func (s *Schema) Gte(value time.Time) *Schema {
	s.compareRules.Put(Rule.Gte(CodeGte, value))
	return s
}

func (s *Schema) Lt(value time.Time) *Schema {
	s.compareRules.Put(Rule.Lt(CodeLt, value))
	return s
}

func (s *Schema) Lte(value time.Time) *Schema {
	s.compareRules.Put(Rule.Lte(CodeLte, value))
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
			stop := context.AddIssue(CodeType, "expected date", map[string]any{
				"expected": "date",
				"actual":   value.Kind.String(),
			})
			var zero time.Time
			return zero, stop
		}

		stop := context.AddIssue(CodeInvalid, "invalid date", map[string]any{
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
		stop := context.AddIssue(CodeDateTime, "expected datetime", map[string]any{
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
