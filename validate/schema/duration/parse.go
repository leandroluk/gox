// schema/duration/parse.go
package duration

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/leandroluk/go/validate/schema"
)

func parseDurationWithOptions(options schema.Options, input string) (time.Duration, bool) {
	if options.CoerceTrimSpace {
		input = strings.TrimSpace(input)
	}
	if options.CoerceNumberUnderscore {
		input = removeUnderscore(input)
	}

	parsed, err := time.ParseDuration(input)
	if err != nil {
		return 0, false
	}

	return parsed, true
}

func parseNanosecondsTextWithOptions(options schema.Options, text string) (time.Duration, bool) {
	if options.CoerceTrimSpace {
		text = strings.TrimSpace(text)
	}
	if options.CoerceNumberUnderscore {
		text = removeUnderscore(text)
	}

	value, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return 0, false
	}

	return time.Duration(value), true
}

func parseGoNumericToInt64(input any) (int64, bool) {
	switch value := input.(type) {
	case int:
		return int64(value), true
	case int8:
		return int64(value), true
	case int16:
		return int64(value), true
	case int32:
		return int64(value), true
	case int64:
		return value, true

	case uint:
		if uint64(value) > uint64(math.MaxInt64) {
			return 0, false
		}
		return int64(value), true
	case uint8:
		return int64(value), true
	case uint16:
		return int64(value), true
	case uint32:
		return int64(value), true
	case uint64:
		if value > uint64(math.MaxInt64) {
			return 0, false
		}
		return int64(value), true

	default:
		return 0, false
	}
}

func durationFromInt64(options schema.Options, value int64) time.Duration {
	secondsFlag := options.CoerceDurationSeconds
	millisFlag := options.CoerceDurationMilliseconds

	if secondsFlag && !millisFlag {
		return time.Duration(value) * time.Second
	}

	if millisFlag && !secondsFlag {
		return time.Duration(value) * time.Millisecond
	}

	abs := value
	if abs < 0 {
		abs = -abs
	}

	if abs >= 100_000_000_000 {
		return time.Duration(value) * time.Millisecond
	}

	return time.Duration(value) * time.Second
}

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
