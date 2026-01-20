// internal/ast/query.go
package ast

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

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
