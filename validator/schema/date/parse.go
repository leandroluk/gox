// schema/date/parse.go
package date

import (
	"strconv"
	"strings"
	"time"

	"github.com/leandroluk/go/validator/schema"
)

func parseDate(options schema.Options, input string) (time.Time, string, bool) {
	location := options.TimeLocation
	if location == nil {
		location = time.UTC
	}

	for _, layout := range options.DateLayouts {
		if layout == "" {
			continue
		}
		if parsed, err := time.ParseInLocation(layout, input, location); err == nil {
			return parsed, layout, true
		}
	}

	return time.Time{}, "", false
}

func parseUnixTextWithOptions(options schema.Options, text string) (time.Time, bool) {
	if options.CoerceTrimSpace {
		text = strings.TrimSpace(text)
	}
	if options.CoerceNumberUnderscore {
		text = removeUnderscore(text)
	}

	value, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return time.Time{}, false
	}

	return unixFromInt64(options, value), true
}

func parseUnixNumberWithOptions(options schema.Options, text string) (time.Time, bool) {
	value, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return time.Time{}, false
	}
	return unixFromInt64(options, value), true
}

func unixFromInt64(options schema.Options, value int64) time.Time {
	secondsFlag := options.CoerceDateUnixSeconds
	millisFlag := options.CoerceDateUnixMilliseconds

	if !secondsFlag && !millisFlag {
		return time.Unix(value, 0).UTC()
	}

	if secondsFlag && !millisFlag {
		return time.Unix(value, 0).UTC()
	}

	if millisFlag && !secondsFlag {
		return unixMillis(value)
	}

	abs := value
	if abs < 0 {
		abs = -abs
	}

	if abs >= 100_000_000_000 {
		return unixMillis(value)
	}

	return time.Unix(value, 0).UTC()
}

func unixMillis(value int64) time.Time {
	seconds := value / 1000
	millisRemainder := value % 1000
	if millisRemainder < 0 {
		millisRemainder = -millisRemainder
	}
	return time.Unix(seconds, millisRemainder*1_000_000).UTC()
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

func layoutHasClock(layout string) bool {
	if strings.Contains(layout, "15") {
		return true
	}
	if strings.Contains(layout, "03") {
		return true
	}
	if strings.Contains(layout, "04") {
		return true
	}
	if strings.Contains(layout, "05") {
		return true
	}
	if strings.Contains(layout, "PM") || strings.Contains(layout, "pm") {
		return true
	}
	return false
}
