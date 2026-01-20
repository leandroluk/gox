// internal/ast/hash.go
package ast

import (
	"encoding/binary"
	"hash"
	"hash/fnv"
	"sort"
	"strconv"
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
