// internal/codec/normalize.go
package codec

import "strings"

func NormalizeNumberString(input string, trimSpace bool, allowUnderscore bool) string {
	if trimSpace {
		input = strings.TrimSpace(input)
	}
	if !allowUnderscore || strings.IndexByte(input, '_') < 0 {
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
