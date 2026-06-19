// schema/number/number.go
package number

import (
	"reflect"
	"strings"

	"github.com/leandroluk/gox/validate/internal/ast"
	"github.com/leandroluk/gox/validate/internal/codec"
	"github.com/leandroluk/gox/validate/internal/defaults"
	"github.com/leandroluk/gox/validate/internal/engine"
	"github.com/leandroluk/gox/validate/internal/reflection"
	"github.com/leandroluk/gox/validate/internal/ruleset"
	"github.com/leandroluk/gox/validate/internal/types"
	"github.com/leandroluk/gox/validate/schema"
	"github.com/leandroluk/gox/validate/schema/number/util"
)

const (
	CodeRequired = "number.required"
	CodeType     = "number.type"
	CodeInvalid  = "number.invalid"
	CodeMin      = "number.min"
	CodeMax      = "number.max"

	CodeEq  = "number.eq"
	CodeNe  = "number.ne"
	CodeGt  = "number.gt"
	CodeGte = "number.gte"
	CodeLt  = "number.lt"
	CodeLte = "number.lte"

	CodeOneOf = "number.oneof"
)

type ruleMap[T any] struct {
	Eq           T
	Gt           T
	Gte          T
	Lt           T
	Lte          T
	Max          T
	Min          T
	Ne           T
	OneOf        T
	Incomparable T
}

var Msg = ruleMap[string]{
	Eq:           "must be equal",
	Gt:           "must be greater",
	Gte:          "must be greater or equal",
	Lt:           "must be lower",
	Lte:          "must be lower or equal",
	Max:          "too large",
	Min:          "too small",
	Ne:           "must not be equal",
	OneOf:        "not allowed",
	Incomparable: "incomparable",
}

func RuleAs[T types.Number]() ruleMap[func(code string, expected ...any) ruleset.Rule[T]] {
	return ruleMap[func(code string, expected ...any) ruleset.Rule[T]]{
		Eq: func(code string, expectedList ...any) ruleset.Rule[T] {
			expected := expectedList[0].(T)
			return ruleset.New("", func(actual T, context *engine.Context) (T, bool) {
				if util.NumberEqual(actual, expected) {
					return actual, false
				}
				stop := context.AddIssue(code, Msg.Eq, types.AnyMap{"expected": expected, "actual": actual})
				return actual, stop
			})
		},
		Gt: func(code string, expectedList ...any) ruleset.Rule[T] {
			return ruleset.New("", func(actual T, context *engine.Context) (T, bool) {
				expected := expectedList[0].(T)
				if util.IsNaN(actual) || util.IsNaN(expected) {
					stop := context.AddIssue(code, Msg.Incomparable, types.AnyMap{"expected": expected, "actual": actual})
					return actual, stop
				}
				if actual > expected {
					return actual, false
				}
				stop := context.AddIssue(code, Msg.Gt, types.AnyMap{"expected": expectedList, "actual": actual})
				return actual, stop
			})
		},
		Gte: func(code string, expectedList ...any) ruleset.Rule[T] {
			expected := expectedList[0].(T)
			return ruleset.New("", func(actual T, context *engine.Context) (T, bool) {
				if util.IsNaN(actual) || util.IsNaN(expected) {
					stop := context.AddIssue(code, Msg.Incomparable, types.AnyMap{"expected": expected, "actual": actual})
					return actual, stop
				}
				if actual >= expected {
					return actual, false
				}
				stop := context.AddIssue(code, Msg.Gte, types.AnyMap{"expected": expected, "actual": actual})
				return actual, stop
			})
		},
		Lt: func(code string, expectedList ...any) ruleset.Rule[T] {
			expected := expectedList[0].(T)
			return ruleset.New("", func(actual T, context *engine.Context) (T, bool) {
				if util.IsNaN(actual) || util.IsNaN(expected) {
					stop := context.AddIssue(code, Msg.Incomparable, types.AnyMap{"expected": expected, "actual": actual})
					return actual, stop
				}
				if actual < expected {
					return actual, false
				}
				stop := context.AddIssue(code, Msg.Lt, types.AnyMap{"expected": expected, "actual": actual})
				return actual, stop
			})
		},
		Lte: func(code string, expectedList ...any) ruleset.Rule[T] {
			expected := expectedList[0].(T)
			return ruleset.New("", func(actual T, context *engine.Context) (T, bool) {
				if util.IsNaN(actual) || util.IsNaN(expected) {
					stop := context.AddIssue(code, Msg.Incomparable, types.AnyMap{"expected": expected, "actual": actual})
					return actual, stop
				}
				if actual <= expected {
					return actual, false
				}
				stop := context.AddIssue(code, Msg.Lte, types.AnyMap{"expected": expected, "actual": actual})
				return actual, stop
			})
		},
		Max: func(code string, expectedList ...any) ruleset.Rule[T] {
			max := expectedList[0].(T)
			return ruleset.New("max", func(actual T, context *engine.Context) (T, bool) {
				if util.IsNaN(actual) {
					return actual, false
				}
				if actual > max {
					stop := context.AddIssue(code, Msg.Max, types.AnyMap{"max": max, "actual": actual})
					return actual, stop
				}
				return actual, false
			})
		},
		Min: func(code string, expectedList ...any) ruleset.Rule[T] {
			min := expectedList[0].(T)
			return ruleset.New("min", func(actual T, context *engine.Context) (T, bool) {
				if util.IsNaN(actual) {
					return actual, false
				}
				if actual < min {
					stop := context.AddIssue(code, Msg.Min, types.AnyMap{"min": min, "actual": actual})
					return actual, stop
				}
				return actual, false
			})
		},
		Ne: func(code string, expectedList ...any) ruleset.Rule[T] {
			expected := expectedList[0].(T)
			return ruleset.New("", func(actual T, context *engine.Context) (T, bool) {
				if !util.NumberEqual(actual, expected) {
					return actual, false
				}
				stop := context.AddIssue(code, Msg.Ne, types.AnyMap{"expected": expected, "actual": actual})
				return actual, stop
			})
		},
		OneOf: func(code string, expectedList ...any) ruleset.Rule[T] {
			values := make([]T, 0, len(expectedList))
			for _, e := range expectedList {
				values = append(values, e.(T))
			}
			allowed := make([]any, 0, len(values))
			allowedMap := make(map[T]struct{}, len(values))
			allowNaN := false
			for _, value := range values {
				allowed = append(allowed, value)
				if util.IsNaN(value) {
					allowNaN = true
					continue
				}
				allowedMap[value] = struct{}{}
			}
			return ruleset.New("oneof", func(actual T, context *engine.Context) (T, bool) {
				isAllowed := false
				if util.IsNaN(actual) {
					isAllowed = allowNaN
				} else {
					_, isAllowed = allowedMap[actual]
				}
				if isAllowed {
					return actual, false
				}
				stop := context.AddIssue(code, Msg.OneOf, types.AnyMap{"allowed": allowed, "actual": actual})
				return actual, stop
			})
		},
	}
}

func removeUnderscore(input string) string {
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

func parseTo[T types.Number](text string) (T, bool) {
	var zero T

	numberType := reflect.TypeOf(zero)
	if numberType == nil {
		return zero, false
	}

	switch numberType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value, err := codec.ParseInt(text, numberType.Bits())
		if err != nil {
			return zero, false
		}
		return T(value), true

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		value, err := codec.ParseUint(text, numberType.Bits())
		if err != nil {
			return zero, false
		}
		return T(value), true

	case reflect.Float32:
		value, err := codec.ParseFloat(text, numberType.Bits())
		if err != nil {
			return zero, false
		}
		return T(float32(value)), true

	case reflect.Float64:
		value, err := codec.ParseFloat(text, numberType.Bits())
		if err != nil {
			return zero, false
		}
		return T(value), true

	default:
		return zero, false
	}
}

func readDirectNumber[T types.Number](input any) (T, bool) {
	var zero T

	switch any(zero).(type) {
	case int:
		value, ok := input.(int)
		if ok {
			return T(value), true
		}
		pointer, ok := input.(*int)
		if ok && pointer != nil {
			return T(*pointer), true
		}

	case int8:
		value, ok := input.(int8)
		if ok {
			return T(value), true
		}
		pointer, ok := input.(*int8)
		if ok && pointer != nil {
			return T(*pointer), true
		}

	case int16:
		value, ok := input.(int16)
		if ok {
			return T(value), true
		}
		pointer, ok := input.(*int16)
		if ok && pointer != nil {
			return T(*pointer), true
		}

	case int32:
		value, ok := input.(int32)
		if ok {
			return T(value), true
		}
		pointer, ok := input.(*int32)
		if ok && pointer != nil {
			return T(*pointer), true
		}

	case int64:
		value, ok := input.(int64)
		if ok {
			return T(value), true
		}
		pointer, ok := input.(*int64)
		if ok && pointer != nil {
			return T(*pointer), true
		}

	case uint:
		value, ok := input.(uint)
		if ok {
			return T(value), true
		}
		pointer, ok := input.(*uint)
		if ok && pointer != nil {
			return T(*pointer), true
		}

	case uint8:
		value, ok := input.(uint8)
		if ok {
			return T(value), true
		}
		pointer, ok := input.(*uint8)
		if ok && pointer != nil {
			return T(*pointer), true
		}

	case uint16:
		value, ok := input.(uint16)
		if ok {
			return T(value), true
		}
		pointer, ok := input.(*uint16)
		if ok && pointer != nil {
			return T(*pointer), true
		}

	case uint32:
		value, ok := input.(uint32)
		if ok {
			return T(value), true
		}
		pointer, ok := input.(*uint32)
		if ok && pointer != nil {
			return T(*pointer), true
		}

	case uint64:
		value, ok := input.(uint64)
		if ok {
			return T(value), true
		}
		pointer, ok := input.(*uint64)
		if ok && pointer != nil {
			return T(*pointer), true
		}

	case float32:
		value, ok := input.(float32)
		if ok {
			return T(value), true
		}
		pointer, ok := input.(*float32)
		if ok && pointer != nil {
			return T(*pointer), true
		}

	case float64:
		value, ok := input.(float64)
		if ok {
			return T(value), true
		}
		pointer, ok := input.(*float64)
		if ok && pointer != nil {
			return T(*pointer), true
		}
	}

	return zero, false
}

type Schema[T types.Number] struct {
	required  bool
	isDefault bool

	valueRules   *ruleset.Set[T]
	boundRules   *ruleset.Set[T]
	compareRules *ruleset.Set[T]

	defaultProvider defaults.Provider[T]
	rules           []ruleset.RuleFn[T]
	RuleMapAs       ruleMap[func(code string, expected ...any) ruleset.Rule[T]]
}

func New[T types.Number]() *Schema[T] {
	return &Schema[T]{
		valueRules:      ruleset.NewSet[T](),
		boundRules:      ruleset.NewSet[T](),
		compareRules:    ruleset.NewSet[T](),
		defaultProvider: defaults.None[T](),
		rules:           make([]ruleset.RuleFn[T], 0),
		RuleMapAs:       RuleAs[T](),
	}
}

func (s *Schema[T]) putValue(ruleValue ruleset.Rule[T]) *Schema[T] {
	s.valueRules.Put(ruleValue)
	return s
}

func (s *Schema[T]) putBound(ruleValue ruleset.Rule[T]) *Schema[T] {
	s.boundRules.Put(ruleValue)
	return s
}

func (s *Schema[T]) putCompare(ruleValue ruleset.Rule[T]) *Schema[T] {
	s.compareRules.Put(ruleValue)
	return s
}

func (s *Schema[T]) Required() *Schema[T] {
	s.required = true
	return s
}

func (s *Schema[T]) IsDefault() *Schema[T] {
	s.isDefault = true
	return s
}

func (s *Schema[T]) Min(value T) *Schema[T] {
	return s.putBound(s.RuleMapAs.Min(CodeMin, value))
}

func (s *Schema[T]) Max(value T) *Schema[T] {
	return s.putBound(s.RuleMapAs.Max(CodeMax, value))
}

func (s *Schema[T]) Eq(value T) *Schema[T] {
	return s.putCompare(s.RuleMapAs.Eq(CodeEq, value))
}

func (s *Schema[T]) Ne(value T) *Schema[T] {
	return s.putCompare(s.RuleMapAs.Ne(CodeNe, value))
}

func (s *Schema[T]) Gt(value T) *Schema[T] {
	return s.putCompare(s.RuleMapAs.Gt(CodeGt, value))
}

func (s *Schema[T]) Gte(value T) *Schema[T] {
	return s.putCompare(s.RuleMapAs.Gte(CodeGte, value))
}

func (s *Schema[T]) Lt(value T) *Schema[T] {
	return s.putCompare(s.RuleMapAs.Lt(CodeLt, value))
}

func (s *Schema[T]) Lte(value T) *Schema[T] {
	return s.putCompare(s.RuleMapAs.Lte(CodeLte, value))
}

func (s *Schema[T]) OneOf(values ...T) *Schema[T] {
	if len(values) == 0 {
		s.valueRules.Remove("oneof")
		return s
	}
	return s.putValue(s.RuleMapAs.OneOf(CodeOneOf, values))
}

func (s *Schema[T]) Default(value T) *Schema[T] {
	s.defaultProvider = defaults.Value(value)
	return s
}

func (s *Schema[T]) DefaultFunc(fn func() T) *Schema[T] {
	s.defaultProvider = defaults.Func(fn)
	return s
}

func (s *Schema[T]) Custom(ruleValue ruleset.RuleFn[T]) *Schema[T] {
	if ruleValue != nil {
		s.rules = append(s.rules, ruleValue)
	}
	return s
}

func (s *Schema[T]) Validate(input any, optionList ...schema.Option) (T, error) {
	options := schema.ApplyOptions(optionList...)
	return s.validateWithOptions(input, options)
}

func (s *Schema[T]) ValidateAny(input any, options schema.Options) (any, error) {
	return s.validateWithOptions(input, options)
}

func (s *Schema[T]) OutputType() reflect.Type {
	return reflect.TypeFor[T]()
}

func (s *Schema[T]) validateWithOptions(input any, options schema.Options) (T, error) {
	context := engine.NewContext(options)

	if parsed, ok := readDirectNumber[T](input); ok {
		output, _ := s.validateNumberValue(context, parsed)
		return output, context.Error()
	}

	value, err := engine.InputToASTWithOptions(input, options)
	if err != nil {
		var zero T
		return zero, err
	}

	output, _ := s.validateAST(context, value)
	return output, context.Error()
}

func (s *Schema[T]) validateAST(context *engine.Context, value ast.Value) (T, bool) {
	if defaultValue, ok := defaults.Apply(value.Presence, context.Options, s.defaultProvider); ok {
		return defaultValue, false
	}

	if value.IsMissing() || value.IsNull() {
		if s.required {
			stop := context.AddIssue(CodeRequired, "required")
			var zero T
			return zero, stop
		}
		var zero T
		return zero, false
	}

	var parsed T
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

		parsed, ok = parseTo[T](text)

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

			parsed, ok = parseTo[T](text)
		}

	default:
		ok = false
	}

	if !ok {
		if value.Kind != ast.KindNumber && value.Kind != ast.KindString {
			stop := context.AddIssue(CodeType, "expected number", map[string]any{
				"expected": "number",
				"actual":   value.Kind.String(),
			})
			var zero T
			return zero, stop
		}

		stop := context.AddIssue(CodeInvalid, "invalid number", map[string]any{
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
		var zero T
		return zero, stop
	}

	return s.validateNumberValue(context, parsed)
}

func (s *Schema[T]) validateNumberValue(context *engine.Context, parsed T) (T, bool) {
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
