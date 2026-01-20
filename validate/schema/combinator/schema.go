// schema/combinator/schema.go
package combinator

import (
	"reflect"

	"github.com/leandroluk/gox/validate/schema"
)

type Schema[T any] interface {
	Validate(input any, optionList ...schema.Option) (T, error)
	ValidateAny(input any, options schema.Options) (any, error)
	OutputType() reflect.Type
}
