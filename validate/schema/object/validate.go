// schema/object/validate.go
package object

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/leandroluk/gox/validate/internal/ast"
	"github.com/leandroluk/gox/validate/internal/codec"
	"github.com/leandroluk/gox/validate/internal/defaults"
	"github.com/leandroluk/gox/validate/internal/engine"
	"github.com/leandroluk/gox/validate/internal/ruleset"
	"github.com/leandroluk/gox/validate/schema"
)

func (s *Schema[T]) validateWithOptions(input any, options schema.Options) (T, error) {
	var zero T

	if s == nil {
		return zero, fmt.Errorf("schema is nil")
	}
	if s.buildError != nil {
		return zero, s.buildError
	}

	context := engine.NewContext(options)

	value, err := engine.InputToASTWithOptions(input, options)
	if err != nil {
		return zero, err
	}

	output, _ := s.validateAST(context, value, input)
	return output, context.Error()
}

func (s *Schema[T]) validateAST(context *engine.Context, value ast.Value, originalInput any) (T, bool) {
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

	if value.Kind != ast.KindObject {
		stop := context.AddIssueWithMeta(CodeType, "expected object", map[string]any{
			"expected": "object",
			"actual":   value.Kind.String(),
		})
		var zero T
		return zero, stop
	}

	if s.structOnly {
		var output T
		if err := codec.DecodeInto(value, &output); err != nil {
			stop := context.AddIssueWithMeta(CodeInvalid, "invalid object", map[string]any{
				"error": err.Error(),
			})
			var zero T
			return zero, stop
		}

		if !s.noStructLevel && len(s.rules) > 0 {
			for _, ruleValue := range s.rules {
				if ruleValue(output, context) {
					return output, true
				}
			}
		}

		return output, false
	}

	var output T
	outputPointer := unsafe.Pointer(&output)

	for _, compiledField := range s.fields {
		context.PushField(compiledField.name)

		child, ok := value.Object[compiledField.name]
		if !ok {
			child = ast.MissingValue()
		}

		action, stop := s.applyFieldPlan(context, value, compiledField, child)
		if action == fieldActionSkip {
			compiledField.assign(outputPointer, reflect.Zero(compiledField.fieldType).Interface())
			context.Pop()
			if stop {
				return output, true
			}
			continue
		}
		if stop {
			compiledField.assign(outputPointer, reflect.Zero(compiledField.fieldType).Interface())
			context.Pop()
			return output, true
		}

		fieldValue, fieldError := compiledField.validate(context, child)
		if fieldError != nil {
			stop := context.AddIssueWithMeta(CodeFieldDecode, "invalid field", map[string]any{
				"error": fieldError.Error(),
			})

			compiledField.assign(outputPointer, reflect.Zero(compiledField.fieldType).Interface())
			context.Pop()

			if stop {
				return output, true
			}
			continue
		}

		compiledField.assign(outputPointer, fieldValue)
		context.Pop()
	}

	if !s.noStructLevel && len(s.rules) > 0 {
		if ruleset.Apply(output, context, s.rules...) {
			return output, true
		}
	}

	return output, false
}

type fieldAction uint8

const (
	fieldActionValidate fieldAction = iota
	fieldActionSkip
)

func (s *Schema[T]) applyFieldPlan(context *engine.Context, root ast.Value, fieldValue field[T], child ast.Value) (fieldAction, bool) {
	if len(fieldValue.skipUnlessConditions) == 0 &&
		len(fieldValue.excludedConditions) == 0 &&
		len(fieldValue.requiredConditions) == 0 &&
		len(fieldValue.comparators) == 0 &&
		!fieldValue.required {
		if child.IsMissing() || child.IsNull() {
			return fieldActionSkip, false
		}
		return fieldActionValidate, false
	}

	for _, cond := range fieldValue.skipUnlessConditions {
		if cond.ShouldSkip(context, root) {
			return fieldActionSkip, false
		}
	}

	childPresent := !child.IsMissing() && !child.IsNull()

	for _, cond := range fieldValue.excludedConditions {
		skip, stop := cond.Apply(context, root, child, childPresent)
		if stop {
			return fieldActionSkip, true
		}
		if skip {
			return fieldActionSkip, false
		}
	}

	for _, cond := range fieldValue.requiredConditions {
		if cond.Apply(context, root, child, childPresent) {
			return fieldActionSkip, true
		}
	}

	if !childPresent && !fieldValue.required {
		return fieldActionSkip, false
	}

	for _, comparator := range fieldValue.comparators {
		if comparator.Apply(context, root, child) {
			return fieldActionValidate, true
		}
	}

	return fieldActionValidate, false
}
