// schema/schema.go
package schema

import "reflect"

type Schema[T any] interface {
	Validate(input any, optionList ...Option) (T, error)
}

type AnySchema interface {
	ValidateAny(input any, options Options) (any, error)
	OutputType() reflect.Type
}
