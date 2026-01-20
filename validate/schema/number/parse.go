// schema/number/parse.go
package number

import (
	"reflect"
	"strings"

	"github.com/leandroluk/go/validate/internal/codec"
	"github.com/leandroluk/go/validate/internal/types"
)

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

func parseTo[N types.Number](text string) (N, bool) {
	var zero N

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
		return N(value), true

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		value, err := codec.ParseUint(text, numberType.Bits())
		if err != nil {
			return zero, false
		}
		return N(value), true

	case reflect.Float32:
		value, err := codec.ParseFloat(text, numberType.Bits())
		if err != nil {
			return zero, false
		}
		return N(float32(value)), true

	case reflect.Float64:
		value, err := codec.ParseFloat(text, numberType.Bits())
		if err != nil {
			return zero, false
		}
		return N(value), true

	default:
		return zero, false
	}
}

func readDirectNumber[N types.Number](input any) (N, bool) {
	var zero N

	switch any(zero).(type) {
	case int:
		value, ok := input.(int)
		if ok {
			return N(value), true
		}
		pointer, ok := input.(*int)
		if ok && pointer != nil {
			return N(*pointer), true
		}

	case int8:
		value, ok := input.(int8)
		if ok {
			return N(value), true
		}
		pointer, ok := input.(*int8)
		if ok && pointer != nil {
			return N(*pointer), true
		}

	case int16:
		value, ok := input.(int16)
		if ok {
			return N(value), true
		}
		pointer, ok := input.(*int16)
		if ok && pointer != nil {
			return N(*pointer), true
		}

	case int32:
		value, ok := input.(int32)
		if ok {
			return N(value), true
		}
		pointer, ok := input.(*int32)
		if ok && pointer != nil {
			return N(*pointer), true
		}

	case int64:
		value, ok := input.(int64)
		if ok {
			return N(value), true
		}
		pointer, ok := input.(*int64)
		if ok && pointer != nil {
			return N(*pointer), true
		}

	case uint:
		value, ok := input.(uint)
		if ok {
			return N(value), true
		}
		pointer, ok := input.(*uint)
		if ok && pointer != nil {
			return N(*pointer), true
		}

	case uint8:
		value, ok := input.(uint8)
		if ok {
			return N(value), true
		}
		pointer, ok := input.(*uint8)
		if ok && pointer != nil {
			return N(*pointer), true
		}

	case uint16:
		value, ok := input.(uint16)
		if ok {
			return N(value), true
		}
		pointer, ok := input.(*uint16)
		if ok && pointer != nil {
			return N(*pointer), true
		}

	case uint32:
		value, ok := input.(uint32)
		if ok {
			return N(value), true
		}
		pointer, ok := input.(*uint32)
		if ok && pointer != nil {
			return N(*pointer), true
		}

	case uint64:
		value, ok := input.(uint64)
		if ok {
			return N(value), true
		}
		pointer, ok := input.(*uint64)
		if ok && pointer != nil {
			return N(*pointer), true
		}

	case float32:
		value, ok := input.(float32)
		if ok {
			return N(value), true
		}
		pointer, ok := input.(*float32)
		if ok && pointer != nil {
			return N(*pointer), true
		}

	case float64:
		value, ok := input.(float64)
		if ok {
			return N(value), true
		}
		pointer, ok := input.(*float64)
		if ok && pointer != nil {
			return N(*pointer), true
		}
	}

	return zero, false
}
