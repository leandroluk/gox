// schema/combinator/oneof.go
package combinator

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/leandroluk/gox/validate/internal/issues"
	"github.com/leandroluk/gox/validate/schema"
)

type OneOfSchema[T any] struct {
	schemaList []Schema[T]
}

func OneOf[T any](schemaList ...Schema[T]) *OneOfSchema[T] {
	if len(schemaList) == 0 {
		panic("combinator.OneOf: at least one schema is required")
	}
	return &OneOfSchema[T]{schemaList: append([]Schema[T](nil), schemaList...)}
}

func (s *OneOfSchema[T]) Validate(input any, optionList ...schema.Option) (T, error) {
	options := schema.ApplyOptions(optionList...)
	return s.validateWithOptions(input, options)
}

func (s *OneOfSchema[T]) ValidateAny(input any, options schema.Options) (any, error) {
	output, err := s.validateWithOptions(input, options)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (s *OneOfSchema[T]) OutputType() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

func (s *OneOfSchema[T]) validateWithOptions(input any, options schema.Options) (T, error) {
	var zero T

	var bestError error
	bestCount := int(^uint(0) >> 1)

	matchCount := 0
	var matchedOutput T

	for _, candidate := range s.schemaList {
		outputAny, err := candidate.ValidateAny(input, options)
		if err == nil {
			output, ok := outputAny.(T)
			if !ok {
				return zero, fmt.Errorf("combinator.OneOf: schema returned %T, expected %T", outputAny, zero)
			}
			matchCount++
			matchedOutput = output
			if matchCount > 1 {
				break
			}
			continue
		}

		var validationError issues.ValidationError
		if !errors.As(err, &validationError) {
			return zero, err
		}

		count := len(validationError.Issues)
		if count < bestCount {
			bestCount = count
			bestError = err
		}
	}

	if matchCount == 1 {
		return matchedOutput, nil
	}

	if matchCount == 0 {
		if bestError == nil {
			return zero, nil
		}
		return zero, bestError
	}

	issue := issues.NewIssue("", CodeOneOf, "expected exactly one match").WithMetaMap(map[string]any{
		"matches": matchCount,
	})
	return zero, issues.NewValidationError([]issues.Issue{issue})
}
