// schema/combinator/anyof.go
package combinator

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/leandroluk/go/validator/internal/issues"
	"github.com/leandroluk/go/validator/schema"
)

type AnyOfSchema[T any] struct {
	schemaList []Schema[T]
}

func AnyOf[T any](schemaList ...Schema[T]) *AnyOfSchema[T] {
	if len(schemaList) == 0 {
		panic("combinator.AnyOf: at least one schema is required")
	}
	return &AnyOfSchema[T]{schemaList: append([]Schema[T](nil), schemaList...)}
}

func Or[T any](schemaList ...Schema[T]) *AnyOfSchema[T] {
	return AnyOf(schemaList...)
}

func (s *AnyOfSchema[T]) Validate(input any, optionList ...schema.Option) (T, error) {
	options := schema.ApplyOptions(optionList...)
	return s.validateWithOptions(input, options)
}

func (s *AnyOfSchema[T]) ValidateAny(input any, options schema.Options) (any, error) {
	output, err := s.validateWithOptions(input, options)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (s *AnyOfSchema[T]) OutputType() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

func (s *AnyOfSchema[T]) validateWithOptions(input any, options schema.Options) (T, error) {
	var zero T

	var bestError error
	bestCount := int(^uint(0) >> 1)

	for _, candidate := range s.schemaList {
		outputAny, err := candidate.ValidateAny(input, options)
		if err == nil {
			output, ok := outputAny.(T)
			if !ok {
				return zero, fmt.Errorf("combinator.AnyOf: schema returned %T, expected %T", outputAny, zero)
			}
			return output, nil
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

	if bestError == nil {
		return zero, nil
	}

	return zero, bestError
}
