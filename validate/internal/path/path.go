// internal/path/path.go
package path

import (
	"strconv"
	"strings"
	"unicode"
)

type Builder struct {
	parts []Part
}

func NewBuilder() Builder {
	return Builder{parts: make([]Part, 0, 8)}
}

func (builder *Builder) Reset() {
	builder.parts = builder.parts[:0]
}

func (builder *Builder) Len() int {
	return len(builder.parts)
}

func (builder *Builder) PushField(name string) {
	builder.parts = append(builder.parts, Part{
		Kind:  PartField,
		Field: name,
	})
}

func (builder *Builder) PushIndex(index int) {
	builder.parts = append(builder.parts, Part{
		Kind:  PartIndex,
		Index: index,
	})
}

func (builder *Builder) PushKey(key string) {
	builder.parts = append(builder.parts, Part{
		Kind: PartKey,
		Key:  key,
	})
}

func (builder *Builder) Pop() {
	if len(builder.parts) == 0 {
		return
	}
	builder.parts = builder.parts[:len(builder.parts)-1]
}

func (builder *Builder) String() string {
	if len(builder.parts) == 0 {
		return ""
	}
	var stringBuilder strings.Builder
	AppendTo(&stringBuilder, builder.parts)
	return stringBuilder.String()
}

func (builder *Builder) Snapshot() []Part {
	return append([]Part(nil), builder.parts...)
}

func (builder *Builder) Restore(parts []Part) {
	builder.parts = append(builder.parts[:0], parts...)
}

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
