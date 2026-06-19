// schema/number/util/util.go
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

func NumberEqual[N types.Number](a N, b N) bool {
	if IsNaN(a) || IsNaN(b) {
		return IsNaN(a) && IsNaN(b)
	}
	return a == b
}
