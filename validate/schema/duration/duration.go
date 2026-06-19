// schema/duration/duration.go
package duration

import (
	"math"
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
	CodeRequired = "duration.required"
	CodeType     = "duration.type"
	CodeInvalid  = "duration.invalid"
	CodeMin      = "duration.min"
	CodeMax      = "duration.max"

	CodeEq  = "duration.eq"
	CodeNe  = "duration.ne"
	CodeGt  = "duration.gt"
	CodeGte = "duration.gte"
	CodeLt  = "duration.lt"
	CodeLte = "duration.lte"
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
	Max: "too large",
	Min: "too small",
	Ne:  "must not be equal",
}

var Rule = ruleMap[func(code string, expected time.Duration) ruleset.Rule[time.Duration]]{
	Eq: func(code string, expected time.Duration) ruleset.Rule[time.Duration] {
		return ruleset.New("", func(actual time.Duration, context *engine.Context) (time.Duration, bool) {
			if actual == expected {
				return actual, false
			}
			stop := context.AddIssue(code, Msg.Eq, types.AnyMap{"expected": expected.String(), "actual": actual.String()})
			return actual, stop
		})
	},
	Gt: func(code string, expected time.Duration) ruleset.Rule[time.Duration] {
		return ruleset.New("", func(actual time.Duration, context *engine.Context) (time.Duration, bool) {
			if actual > expected {
				return actual, false
			}
			stop := context.AddIssue(code, Msg.Gt, types.AnyMap{"expected": expected.String(), "actual": actual.String()})
			return actual, stop
		})
	},
	Gte: func(code string, expected time.Duration) ruleset.Rule[time.Duration] {
		return ruleset.New("", func(actual time.Duration, context *engine.Context) (time.Duration, bool) {
			if actual >= expected {
				return actual, false
			}
			stop := context.AddIssue(code, Msg.Gte, types.AnyMap{"expected": expected.String(), "actual": actual.String()})
			return actual, stop
		})
	},
	Lt: func(code string, expected time.Duration) ruleset.Rule[time.Duration] {
		return ruleset.New("", func(actual time.Duration, context *engine.Context) (time.Duration, bool) {
			if actual < expected {
				return actual, false
			}
			stop := context.AddIssue(code, Msg.Lt, types.AnyMap{"expected": expected.String(), "actual": actual.String()})
			return actual, stop
		})
	},
	Lte: func(code string, expected time.Duration) ruleset.Rule[time.Duration] {
		return ruleset.New("", func(actual time.Duration, context *engine.Context) (time.Duration, bool) {
			if actual <= expected {
				return actual, false
			}
			stop := context.AddIssue(code, Msg.Lte, types.AnyMap{"expected": expected.String(), "actual": actual.String()})
			return actual, stop
		})
	},
	Max: func(code string, maximum time.Duration) ruleset.Rule[time.Duration] {
		return ruleset.New("max", func(actual time.Duration, context *engine.Context) (time.Duration, bool) {
			if actual > maximum {
				stop := context.AddIssue(code, Msg.Max, types.AnyMap{"max": maximum.String(), "actual": actual.String()})
				return actual, stop
			}
			return actual, false
		})
	},
	Min: func(code string, minimum time.Duration) ruleset.Rule[time.Duration] {
		return ruleset.New("min", func(actual time.Duration, context *engine.Context) (time.Duration, bool) {
			if actual < minimum {
				stop := context.AddIssue(code, Msg.Min, types.AnyMap{"min": minimum.String(), "actual": actual.String()})
				return actual, stop
			}
			return actual, false
		})
	},
	Ne: func(code string, expected time.Duration) ruleset.Rule[time.Duration] {
		return ruleset.New("", func(actual time.Duration, context *engine.Context) (time.Duration, bool) {
			if actual != expected {
				return actual, false
			}
			stop := context.AddIssue(code, Msg.Ne, types.AnyMap{"expected": expected.String(), "actual": actual.String()})
			return actual, stop
		})
	},
}

func parseDurationWithOptions(options schema.Options, input string) (time.Duration, bool) {
	if options.CoerceTrimSpace {
		input = strings.TrimSpace(input)
	}
	if options.CoerceNumberUnderscore {
		input = removeDurationUnderscore(input)
	}

	parsed, err := time.ParseDuration(input)
	if err != nil {
		return 0, false
	}

	return parsed, true
}

func parseNanosecondsTextWithOptions(options schema.Options, text string) (time.Duration, bool) {
	if options.CoerceTrimSpace {
		text = strings.TrimSpace(text)
	}
	if options.CoerceNumberUnderscore {
		text = removeDurationUnderscore(text)
	}

	value, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return 0, false
	}

	return time.Duration(value), true
}

func parseGoNumericToInt64(input any) (int64, bool) {
	switch value := input.(type) {
	case int:
		return int64(value), true
	case int8:
		return int64(value), true
	case int16:
		return int64(value), true
	case int32:
		return int64(value), true
	case int64:
		return value, true

	case uint:
		if uint64(value) > uint64(math.MaxInt64) {
			return 0, false
		}
		return int64(value), true
	case uint8:
		return int64(value), true
	case uint16:
		return int64(value), true
	case uint32:
		return int64(value), true
	case uint64:
		if value > uint64(math.MaxInt64) {
			return 0, false
		}
		return int64(value), true

	default:
		return 0, false
	}
}

func durationFromInt64(options schema.Options, value int64) time.Duration {
	secondsFlag := options.CoerceDurationSeconds
	millisFlag := options.CoerceDurationMilliseconds

	if secondsFlag && !millisFlag {
		return time.Duration(value) * time.Second
	}

	if millisFlag && !secondsFlag {
		return time.Duration(value) * time.Millisecond
	}

	abs := value
	if abs < 0 {
		abs = -abs
	}

	if abs >= 100_000_000_000 {
		return time.Duration(value) * time.Millisecond
	}

	return time.Duration(value) * time.Second
}

func removeDurationUnderscore(input string) string {
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

type Schema struct {
	required  bool
	isDefault bool

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

func (s *Schema) IsDefault() *Schema {
	s.isDefault = true
	return s
}

func (s *Schema) Min(value time.Duration) *Schema {
	return s.putBound(Rule.Min(CodeMin, value))
}

func (s *Schema) Max(value time.Duration) *Schema {
	return s.putBound(Rule.Max(CodeMax, value))
}

func (s *Schema) Eq(value time.Duration) *Schema {
	return s.putCompare(Rule.Eq(CodeEq, value))
}

func (s *Schema) Ne(value time.Duration) *Schema {
	return s.putCompare(Rule.Ne(CodeNe, value))
}

func (s *Schema) Gt(value time.Duration) *Schema {
	return s.putCompare(Rule.Gt(CodeGt, value))
}

func (s *Schema) Gte(value time.Duration) *Schema {
	return s.putCompare(Rule.Gte(CodeGte, value))
}

func (s *Schema) Lt(value time.Duration) *Schema {
	return s.putCompare(Rule.Lt(CodeLt, value))
}

func (s *Schema) Lte(value time.Duration) *Schema {
	return s.putCompare(Rule.Lte(CodeLte, value))
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

		context.AddIssue(CodeInvalid, "invalid duration", map[string]any{
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
			stop := context.AddIssue(CodeType, "expected duration", map[string]any{
				"expected": "duration",
				"actual":   value.Kind.String(),
			})
			var zero time.Duration
			return zero, stop
		}

		stop := context.AddIssue(CodeInvalid, "invalid duration", map[string]any{
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
