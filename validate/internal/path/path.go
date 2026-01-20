// internal/path/path.go
package path

import (
	"strconv"
	"strings"
	"unicode"
)

type PartKind uint8

const (
	PartField PartKind = iota + 1
	PartIndex
	PartKey
)

type Part struct {
	Kind  PartKind
	Field string
	Index int
	Key   string
}

func IsIdentifier(value string) bool {
	if value == "" {
		return false
	}

	for index, runeValue := range value {
		if index == 0 {
			if runeValue != '_' && !unicode.IsLetter(runeValue) {
				return false
			}
			continue
		}
		if runeValue != '_' && !unicode.IsLetter(runeValue) && !unicode.IsDigit(runeValue) {
			return false
		}
	}

	return true
}

func QuoteKey(value string) string {
	var builder strings.Builder
	builder.WriteByte('"')
	for _, runeValue := range value {
		switch runeValue {
		case '\\':
			builder.WriteString(`\\`)
		case '"':
			builder.WriteString(`\"`)
		case '\n':
			builder.WriteString(`\n`)
		case '\r':
			builder.WriteString(`\r`)
		case '\t':
			builder.WriteString(`\t`)
		default:
			builder.WriteRune(runeValue)
		}
	}
	builder.WriteByte('"')
	return builder.String()
}

func AppendTo(builder *strings.Builder, parts []Part) {
	for index, part := range parts {
		switch part.Kind {
		case PartField:
			if index > 0 {
				builder.WriteByte('.')
			}
			builder.WriteString(part.Field)

		case PartIndex:
			builder.WriteByte('[')
			builder.WriteString(strconv.Itoa(part.Index))
			builder.WriteByte(']')

		case PartKey:
			if IsIdentifier(part.Key) {
				if index > 0 {
					builder.WriteByte('.')
				}
				builder.WriteString(part.Key)
				continue
			}
			builder.WriteByte('[')
			builder.WriteString(QuoteKey(part.Key))
			builder.WriteByte(']')
		}
	}
}
