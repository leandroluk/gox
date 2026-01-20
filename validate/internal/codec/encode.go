// internal/codec/encode.go
package codec

import (
	"encoding/json"

	"github.com/leandroluk/gox/validate/internal/ast"
)

func Encode(value ast.Value) ([]byte, error) {
	raw := ToRaw(value)
	return json.Marshal(raw)
}

func ToRaw(value ast.Value) any {
	if value.IsMissing() || value.IsNull() || !value.IsPresent() {
		return nil
	}

	switch value.Kind {
	case ast.KindString:
		return value.String

	case ast.KindBoolean:
		return value.Boolean

	case ast.KindNumber:
		return json.Number(value.Number)

	case ast.KindArray:
		if value.Array == nil {
			return []any(nil)
		}
		items := make([]any, 0, len(value.Array))
		for _, entry := range value.Array {
			items = append(items, ToRaw(entry))
		}
		return items

	case ast.KindObject:
		if value.Object == nil {
			return map[string]any(nil)
		}
		object := make(map[string]any, len(value.Object))
		for key, entry := range value.Object {
			object[key] = ToRaw(entry)
		}
		return object

	default:
		return nil
	}
}
