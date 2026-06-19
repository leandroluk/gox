// internal/ast/ast.go
package ast

import (
	"encoding/binary"
	"hash"
	"hash/fnv"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

func Hash(value Value) string {
	hasher := fnv.New64a()
	writeValue(hasher, value)
	return strconv.FormatUint(hasher.Sum64(), 16)
}

func writeValue(hasher hash.Hash64, value Value) {
	if value.IsMissing() {
		_, _ = hasher.Write([]byte{0})
		return
	}

	if value.IsNull() || !value.IsPresent() {
		_, _ = hasher.Write([]byte{1})
		return
	}

	_, _ = hasher.Write([]byte{2, byte(value.Kind)})

	switch value.Kind {
	case KindString:
		writeString(hasher, value.String)

	case KindNumber:
		writeString(hasher, value.Number)

	case KindBoolean:
		if value.Boolean {
			_, _ = hasher.Write([]byte{1})
		} else {
			_, _ = hasher.Write([]byte{0})
		}

	case KindArray:
		writeUvarint(hasher, uint64(len(value.Array)))
		for index := 0; index < len(value.Array); index++ {
			writeValue(hasher, value.Array[index])
		}

	case KindObject:
		keys := make([]string, 0, len(value.Object))
		for key := range value.Object {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		writeUvarint(hasher, uint64(len(keys)))
		for _, key := range keys {
			writeString(hasher, key)
			writeValue(hasher, value.Object[key])
		}

	default:
	}
}

func writeString(hasher hash.Hash64, text string) {
	writeUvarint(hasher, uint64(len(text)))
	if text != "" {
		_, _ = hasher.Write([]byte(text))
	}
}

func writeUvarint(hasher hash.Hash64, value uint64) {
	var buffer [10]byte
	n := binary.PutUvarint(buffer[:], value)
	_, _ = hasher.Write(buffer[:n])
}

type Presence uint8

const (
	Missing Presence = iota
	Null
	Present
)

func (presence Presence) String() string {
	switch presence {
	case Missing:
		return "missing"
	case Null:
		return "null"
	case Present:
		return "present"
	default:
		return "unknown"
	}
}

func Query(root Value, path string) Value {
	path = strings.TrimSpace(path)
	if path == "" || path == "." || path == "$" {
		return root
	}

	if strings.HasPrefix(path, "$.") {
		path = path[2:]
	} else if strings.HasPrefix(path, ".") {
		path = path[1:]
	}

	current := root
	index := 0

	for index < len(path) {
		if path[index] == '.' {
			index++
			continue
		}

		if path[index] == '[' {
			bracketStart := index

			next, token, kind, ok := parseBracket(path, index)
			if !ok {
				return MissingValue()
			}
			index = next

			if kind == bracketIndex {
				if current.Kind != KindArray {
					return MissingValue()
				}
				if token < 0 || token >= len(current.Array) {
					return MissingValue()
				}
				current = current.Array[token]
				continue
			}

			if current.Kind != KindObject {
				return MissingValue()
			}
			key := parseKeyToken(path, bracketStart, token)
			child, ok := current.Object[key]
			if !ok {
				return MissingValue()
			}
			current = child
			continue
		}

		next, name, ok := parseIdentifier(path, index)
		if !ok {
			return MissingValue()
		}
		index = next

		if current.Kind != KindObject {
			return MissingValue()
		}
		child, ok := current.Object[name]
		if !ok {
			return MissingValue()
		}
		current = child
	}

	return current
}

type bracketKind uint8

const (
	bracketKey bracketKind = iota
	bracketIndex
)

func parseIdentifier(input string, start int) (int, string, bool) {
	index := start
	for index < len(input) {
		ch := input[index]
		if ch == '.' || ch == '[' {
			break
		}
		index++
	}
	if index == start {
		return start, "", false
	}
	return index, input[start:index], true
}

func parseBracket(input string, start int) (int, int, bracketKind, bool) {
	index := start
	if index >= len(input) || input[index] != '[' {
		return start, 0, bracketKey, false
	}
	index++

	for index < len(input) && input[index] == ' ' {
		index++
	}

	if index >= len(input) {
		return start, 0, bracketKey, false
	}

	if input[index] == '"' {
		end, ok := scanQuotedKey(input, index)
		if !ok {
			return start, 0, bracketKey, false
		}
		index = end
		for index < len(input) && input[index] == ' ' {
			index++
		}
		if index >= len(input) || input[index] != ']' {
			return start, 0, bracketKey, false
		}
		return index + 1, start, bracketKey, true
	}

	sign := 1
	if input[index] == '-' {
		sign = -1
		index++
	}

	digitStart := index
	for index < len(input) && input[index] >= '0' && input[index] <= '9' {
		index++
	}
	digitEnd := index

	for index < len(input) && input[index] == ' ' {
		index++
	}

	if digitStart != digitEnd && index < len(input) && input[index] == ']' {
		number, err := strconv.Atoi(input[digitStart:digitEnd])
		if err != nil {
			return start, 0, bracketKey, false
		}
		return index + 1, sign * number, bracketIndex, true
	}

	for index < len(input) && input[index] != ']' {
		index++
	}
	if index >= len(input) || input[index] != ']' {
		return start, 0, bracketKey, false
	}
	return index + 1, start, bracketKey, true
}

func scanQuotedKey(input string, quoteIndex int) (int, bool) {
	if quoteIndex >= len(input) || input[quoteIndex] != '"' {
		return quoteIndex, false
	}
	index := quoteIndex + 1
	for index < len(input) {
		ch := input[index]
		if ch == '\\' {
			index++
			if index >= len(input) {
				return index, false
			}
			_, size := utf8.DecodeRuneInString(input[index:])
			if size <= 0 {
				return index, false
			}
			index += size
			continue
		}
		if ch == '"' {
			return index + 1, true
		}
		_, size := utf8.DecodeRuneInString(input[index:])
		if size <= 0 {
			return index, false
		}
		index += size
	}
	return index, false
}

func parseKeyToken(input string, bracketStart int, _ int) string {
	index := bracketStart + 1
	for index < len(input) && input[index] == ' ' {
		index++
	}
	if index < len(input) && input[index] == '"' {
		index++
		var builder strings.Builder
		for index < len(input) {
			ch := input[index]
			if ch == '\\' {
				index++
				if index >= len(input) {
					break
				}
				builder.WriteByte(input[index])
				index++
				continue
			}
			if ch == '"' {
				break
			}
			builder.WriteByte(ch)
			index++
		}
		return builder.String()
	}

	end := index
	for end < len(input) && input[end] != ']' {
		end++
	}
	raw := strings.TrimSpace(input[index:end])
	return raw
}

type Kind uint8

const (
	KindInvalid Kind = iota
	KindString
	KindNumber
	KindBoolean
	KindArray
	KindObject
)

func (kind Kind) String() string {
	switch kind {
	case KindString:
		return "string"
	case KindNumber:
		return "number"
	case KindBoolean:
		return "boolean"
	case KindArray:
		return "array"
	case KindObject:
		return "object"
	default:
		return "invalid"
	}
}

type Value struct {
	Presence Presence
	Kind     Kind

	String  string
	Number  string
	Boolean bool

	Array  []Value
	Object map[string]Value
}

func MissingValue() Value {
	return Value{Presence: Missing, Kind: KindInvalid}
}

func NullValue() Value {
	return Value{Presence: Null, Kind: KindInvalid}
}

func StringValue(value string) Value {
	return Value{Presence: Present, Kind: KindString, String: value}
}

func NumberValue(value string) Value {
	return Value{Presence: Present, Kind: KindNumber, Number: value}
}

func BooleanValue(value bool) Value {
	return Value{Presence: Present, Kind: KindBoolean, Boolean: value}
}

func ArrayValue(value []Value) Value {
	return Value{Presence: Present, Kind: KindArray, Array: value}
}

func ObjectValue(value map[string]Value) Value {
	return Value{Presence: Present, Kind: KindObject, Object: value}
}

func (value Value) IsMissing() bool {
	return value.Presence == Missing
}

func (value Value) IsNull() bool {
	return value.Presence == Null
}

func (value Value) IsPresent() bool {
	return value.Presence == Present
}

func (value Value) Is(kind Kind) bool {
	return value.Kind == kind && value.Presence == Present
}

func (value Value) CloneShallow() Value {
	cloned := value
	if value.Kind == KindArray && value.Array != nil {
		cloned.Array = append([]Value(nil), value.Array...)
	}
	if value.Kind == KindObject && value.Object != nil {
		cloned.Object = make(map[string]Value, len(value.Object))
		for key, entry := range value.Object {
			cloned.Object[key] = entry
		}
	}
	return cloned
}
