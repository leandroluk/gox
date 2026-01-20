// schema/number/util/is_nan.go
package util

import (
	"math"
	"reflect"

	"github.com/leandroluk/gox/validate/internal/types"
)

func IsNaN[N types.Number](value N) bool {
	kind := reflect.TypeOf(value).Kind()

	switch kind {
	case reflect.Float32:
		return math.IsNaN(float64(float32(value)))
	case reflect.Float64:
		return math.IsNaN(float64(value))
	default:
		return false
	}
}
