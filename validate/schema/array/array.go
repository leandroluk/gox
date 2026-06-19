// schema/array/array.go
package array

import (
	"reflect"

	"github.com/leandroluk/gox/validate/internal/ast"
	"github.com/leandroluk/gox/validate/internal/codec"
	"github.com/leandroluk/gox/validate/internal/defaults"
	"github.com/leandroluk/gox/validate/internal/engine"
	"github.com/leandroluk/gox/validate/internal/ruleset"
	"github.com/leandroluk/gox/validate/internal/types"
	"github.com/leandroluk/gox/validate/schema"
)

const (
	CodeRequired = "array.required"
	CodeType     = "array.type"
	CodeMin      = "array.min"
	CodeMax      = "array.max"
	CodeItem     = "array.item"

	CodeLen = "array.len"
	CodeEq  = "array.eq"
	CodeNe  = "array.ne"
	CodeGt  = "array.gt"
	CodeGte = "array.gte"
	CodeLt  = "array.lt"
	CodeLte = "array.lte"

	CodeUnique = "array.unique"
)

type ruleMap[T any] struct {
	Eq  T
	Gt  T
	Gte T
	Len T
	Lt  T
	Lte T
	Max T
	Min T
	Ne  T
}

var Msg = ruleMap[string]{
	Eq:  "array.eq.message",
	Gt:  "array.gt.message",
	Gte: "array.gte.message",
	Len: "array.len.message",
	Lt:  "array.lt.message",
	Lte: "array.lte.message",
	Max: "array.max.message",
	Min: "array.min.message",
	Ne:  "array.ne.message",
}
var Rule = ruleMap[func(code string, expected int) ruleset.Rule[int]]{
	Eq: func(code string, expected int) ruleset.Rule[int] {
		return ruleset.New("", func(actual int, context *engine.Context) (int, bool) {
			if actual == expected {
				return actual, false
			}
			stop := context.AddIssue(code, Msg.Eq, types.AnyMap{"expected": expected, "actual": actual})
			return actual, stop
		})
	},
	Gt: func(code string, expected int) ruleset.Rule[int] {
		return ruleset.New("", func(actual int, context *engine.Context) (int, bool) {
			if actual > expected {
				return actual, false
			}
			stop := context.AddIssue(code, Msg.Gt, types.AnyMap{"expected": expected, "actual": actual})
			return actual, stop
		})
	},
	Gte: func(code string, expected int) ruleset.Rule[int] {
		return ruleset.New("", func(actual int, context *engine.Context) (int, bool) {
			if actual >= expected {
				return actual, false
			}
			stop := context.AddIssue(code, Msg.Gte, types.AnyMap{"expected": expected, "actual": actual})
			return actual, stop
		})
	},
	Len: func(code string, expected int) ruleset.Rule[int] {
		return ruleset.New("", func(actual int, context *engine.Context) (int, bool) {
			if actual == expected {
				return actual, false
			}
			stop := context.AddIssue(code, Msg.Len, types.AnyMap{"expected": expected, "actual": actual})
			return actual, stop
		})
	},
	Lt: func(code string, expected int) ruleset.Rule[int] {
		return ruleset.New("", func(actual int, context *engine.Context) (int, bool) {
			if actual < expected {
				return actual, false
			}
			stop := context.AddIssue(code, Msg.Lt, types.AnyMap{"expected": expected, "actual": actual})
			return actual, stop
		})
	},
	Lte: func(code string, expected int) ruleset.Rule[int] {
		return ruleset.New("", func(actual int, context *engine.Context) (int, bool) {
			if actual <= expected {
				return actual, false
			}
			stop := context.AddIssue(code, Msg.Lte, types.AnyMap{"expected": expected, "actual": actual})
			return actual, stop
		})
	},
	Max: func(code string, max int) ruleset.Rule[int] {
		return ruleset.New("max", func(actual int, context *engine.Context) (int, bool) {
			if actual > max {
				stop := context.AddIssue(code, Msg.Max, types.AnyMap{"max": max, "actual": actual})
				return actual, stop
			}
			return actual, false
		})
	},
	Min: func(code string, min int) ruleset.Rule[int] {
		return ruleset.New("min", func(actual int, context *engine.Context) (int, bool) {
			if actual < min {
				stop := context.AddIssue(code, Msg.Min, types.AnyMap{"min": min, "actual": actual})
				return actual, stop
			}
			return actual, false
		})
	},
	Ne: func(code string, expected int) ruleset.Rule[int] {
		return ruleset.New("", func(actual int, context *engine.Context) (int, bool) {
			if actual != expected {
				return actual, false
			}
			stop := context.AddIssue(code, Msg.Ne, types.AnyMap{"expected": expected, "actual": actual})
			return actual, stop
		})
	},
}

func normalizeLimit(value int) (int, bool) {
	if value < 0 {
		return 0, false
	}
	return value, true
}

type ItemValidator[T any] func(context *engine.Context, value ast.Value) (T, bool)

type Schema[T any] struct {
	required  bool
	isDefault bool

	lengthRules *ruleset.Set[int]

	defaultProvider defaults.Provider[[]T]

	uniqueEnabled bool
	uniqueHash    func(value T) string
	uniqueEqual   func(left T, right T) bool

	itemValidator ItemValidator[T]
	rules         []ruleset.RuleFn[[]T]
}

func New[T any]() *Schema[T] {
	return &Schema[T]{
		lengthRules:     ruleset.NewSet[int](),
		defaultProvider: defaults.None[[]T](),
		rules:           make([]ruleset.RuleFn[[]T], 0),
	}
}

func (s *Schema[T]) putLength(ruleValue ruleset.Rule[int]) *Schema[T] {
	s.lengthRules.Put(ruleValue)
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

func (s *Schema[T]) Min(length int) *Schema[T] {
	if normalized, ok := normalizeLimit(length); ok {
		return s.putLength(Rule.Min(CodeMin, normalized))
	}
	return s
}

func (s *Schema[T]) Max(length int) *Schema[T] {
	if normalized, ok := normalizeLimit(length); ok {
		return s.putLength(Rule.Max(CodeMax, normalized))
	}
	return s
}

func (s *Schema[T]) Len(length int) *Schema[T] {
	if normalized, ok := normalizeLimit(length); ok {
		return s.putLength(Rule.Len(CodeLen, normalized))
	}
	return s
}

func (s *Schema[T]) Eq(length int) *Schema[T] {
	if normalized, ok := normalizeLimit(length); ok {
		return s.putLength(Rule.Eq(CodeEq, normalized))
	}
	return s
}

func (s *Schema[T]) Ne(length int) *Schema[T] {
	if normalized, ok := normalizeLimit(length); ok {
		return s.putLength(Rule.Ne(CodeNe, normalized))
	}
	return s
}

func (s *Schema[T]) Gt(length int) *Schema[T] {
	if normalized, ok := normalizeLimit(length); ok {
		return s.putLength(Rule.Gt(CodeGt, normalized))
	}
	return s
}

func (s *Schema[T]) Gte(length int) *Schema[T] {
	if normalized, ok := normalizeLimit(length); ok {
		return s.putLength(Rule.Gte(CodeGte, normalized))
	}
	return s
}

func (s *Schema[T]) Lt(length int) *Schema[T] {
	if normalized, ok := normalizeLimit(length); ok {
		return s.putLength(Rule.Lt(CodeLt, normalized))
	}
	return s
}

func (s *Schema[T]) Lte(length int) *Schema[T] {
	if normalized, ok := normalizeLimit(length); ok {
		return s.putLength(Rule.Lte(CodeLte, normalized))
	}
	return s
}

func (s *Schema[T]) Unique() *Schema[T] {
	s.uniqueEnabled = true
	s.uniqueHash = nil
	s.uniqueEqual = nil
	return s
}

func (s *Schema[T]) UniqueByHash(hash func(value T) string) *Schema[T] {
	s.uniqueEnabled = true
	s.uniqueHash = hash
	s.uniqueEqual = nil
	return s
}

func (s *Schema[T]) UniqueByEqual(equal func(left T, right T) bool) *Schema[T] {
	s.uniqueEnabled = true
	s.uniqueEqual = equal
	s.uniqueHash = nil
	return s
}

func (s *Schema[T]) Default(value []T) *Schema[T] {
	s.defaultProvider = defaults.Value(value)
	return s
}

func (s *Schema[T]) DefaultFunc(fn func() []T) *Schema[T] {
	s.defaultProvider = defaults.Func(fn)
	return s
}

func (s *Schema[T]) Items(validator ItemValidator[T]) *Schema[T] {
	s.itemValidator = validator
	return s
}

func (s *Schema[T]) Custom(ruleValue ruleset.RuleFn[[]T]) *Schema[T] {
	if ruleValue != nil {
		s.rules = append(s.rules, ruleValue)
	}
	return s
}

func (s *Schema[T]) Validate(input any, optionList ...schema.Option) ([]T, error) {
	options := schema.ApplyOptions(optionList...)
	return s.validateWithOptions(input, options)
}

func (s *Schema[T]) ValidateAny(input any, options schema.Options) (any, error) {
	return s.validateWithOptions(input, options)
}

func (s *Schema[T]) OutputType() reflect.Type {
	return reflect.TypeFor[[]T]()
}

func (s *Schema[T]) validateWithOptions(input any, options schema.Options) ([]T, error) {
	context := engine.NewContext(options)

	value, err := engine.InputToASTWithOptions(input, options)
	if err != nil {
		var zero []T
		return zero, err
	}

	output, _ := s.validateAST(context, value)
	return output, context.Error()
}

func (s *Schema[T]) validateAST(context *engine.Context, value ast.Value) ([]T, bool) {
	if defaultValue, ok := defaults.Apply(value.Presence, context.Options, s.defaultProvider); ok {
		return defaultValue, false
	}

	if value.IsMissing() || value.IsNull() {
		if s.required {
			stop := context.AddIssue(CodeRequired, "required")
			var zero []T
			return zero, stop
		}
		var zero []T
		return zero, false
	}

	if value.Kind != ast.KindArray {
		if context.Options.Coerce {
			coerced := ast.ArrayValue([]ast.Value{value})
			return s.validateAST(context, coerced)
		}
		stop := context.AddIssue(CodeType, "expected array", map[string]any{
			"expected": "array",
			"actual":   value.Kind.String(),
		})
		var zero []T
		return zero, stop
	}

	actualLen := len(value.Array)

	_, stopped := s.lengthRules.ApplyAll(actualLen, context)
	if stopped {
		var zero []T
		return zero, true
	}

	output := make([]T, 0, actualLen)

	var seenHash map[string]int
	if s.uniqueEnabled && s.uniqueEqual == nil {
		seenHash = make(map[string]int, actualLen)
	}

	for index := 0; index < actualLen; index++ {
		context.PushIndex(index)

		itemValue := value.Array[index]

		var item T
		var stop bool

		if s.itemValidator != nil {
			item, stop = s.itemValidator(context, itemValue)
		} else {
			item, stop = decodeFallback[T](context, itemValue)
		}

		if stop {
			context.Pop()
			output = append(output, item)
			return output, true
		}

		if s.uniqueEnabled && !itemValue.IsMissing() && !itemValue.IsNull() {
			firstIndex := -1

			if s.uniqueEqual != nil {
				for previousIndex := 0; previousIndex < len(output); previousIndex++ {
					if s.uniqueEqual(output[previousIndex], item) {
						firstIndex = previousIndex
						break
					}
				}
			} else {
				hash := s.hashItem(itemValue, item)
				if previousIndex, ok := seenHash[hash]; ok {
					firstIndex = previousIndex
				} else {
					seenHash[hash] = index
				}
			}

			if firstIndex >= 0 {
				if context.AddIssue(CodeUnique, "duplicate", map[string]any{
					"first": firstIndex,
				}) {
					context.Pop()
					output = append(output, item)
					return output, true
				}
			}
		}

		output = append(output, item)
		context.Pop()
	}

	if len(s.rules) > 0 {
		if ruleset.Apply(output, context, s.rules...) {
			return output, true
		}
	}

	return output, false
}

func (s *Schema[T]) hashItem(itemValue ast.Value, item T) string {
	if s.uniqueHash != nil {
		return s.uniqueHash(item)
	}
	return ast.Hash(itemValue)
}

func decodeFallback[T any](context *engine.Context, value ast.Value) (T, bool) {
	var out T

	if value.IsMissing() || value.IsNull() {
		return out, false
	}

	if err := codec.DecodeInto(value, &out); err != nil {
		stop := context.AddIssue(CodeItem, "invalid item", map[string]any{"error": err.Error()})
		return out, stop
	}

	return out, false
}
