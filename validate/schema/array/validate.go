// schema/array/validate.go
package array

import (
	"github.com/leandroluk/go/validate/internal/ast"
	"github.com/leandroluk/go/validate/internal/codec"
	"github.com/leandroluk/go/validate/internal/defaults"
	"github.com/leandroluk/go/validate/internal/engine"
	"github.com/leandroluk/go/validate/internal/ruleset"
	"github.com/leandroluk/go/validate/schema"
)

func (s *Schema[E]) validateWithOptions(input any, options schema.Options) ([]E, error) {
	context := engine.NewContext(options)

	value, err := engine.InputToASTWithOptions(input, options)
	if err != nil {
		var zero []E
		return zero, err
	}

	output, _ := s.validateAST(context, value)
	return output, context.Error()
}

func (s *Schema[E]) validateAST(context *engine.Context, value ast.Value) ([]E, bool) {
	if defaultValue, ok := defaults.Apply(value.Presence, context.Options, s.defaultProvider); ok {
		return defaultValue, false
	}

	if value.IsMissing() || value.IsNull() {
		if s.required {
			stop := context.AddIssue(CodeRequired, "required")
			var zero []E
			return zero, stop
		}
		var zero []E
		return zero, false
	}

	if value.Kind != ast.KindArray {
		if context.Options.Coerce {
			coerced := ast.ArrayValue([]ast.Value{value})
			return s.validateAST(context, coerced)
		}
		stop := context.AddIssueWithMeta(CodeType, "expected array", map[string]any{
			"expected": "array",
			"actual":   value.Kind.String(),
		})
		var zero []E
		return zero, stop
	}

	actualLen := len(value.Array)

	_, stopped := s.lengthRules.ApplyAll(actualLen, context)
	if stopped {
		var zero []E
		return zero, true
	}

	output := make([]E, 0, actualLen)

	var seenHash map[string]int
	if s.uniqueEnabled && s.uniqueEqual == nil {
		seenHash = make(map[string]int, actualLen)
	}

	for index := 0; index < actualLen; index++ {
		context.PushIndex(index)

		itemValue := value.Array[index]

		var item E
		var stop bool

		if s.itemValidator != nil {
			item, stop = s.itemValidator(context, itemValue)
		} else {
			item, stop = decodeFallback[E](context, itemValue)
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
				if context.AddIssueWithMeta(CodeUnique, "duplicate", map[string]any{
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

func (s *Schema[E]) hashItem(itemValue ast.Value, item E) string {
	if s.uniqueHash != nil {
		return s.uniqueHash(item)
	}
	return ast.Hash(itemValue)
}

func decodeFallback[E any](context *engine.Context, value ast.Value) (E, bool) {
	var out E

	if value.IsMissing() || value.IsNull() {
		return out, false
	}

	if err := codec.DecodeInto(value, &out); err != nil {
		stop := context.AddIssueWithMeta(CodeItem, "invalid item", map[string]any{
			"error": err.Error(),
		})
		return out, stop
	}

	return out, false
}
