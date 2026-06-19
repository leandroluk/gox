// internal/codec/codec.go
package codec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/leandroluk/gox/validate/internal/ast"
)

func Decode(data []byte) (ast.Value, error) {
	if len(data) == 0 {
		return ast.NullValue(), nil
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	var raw any
	if err := decoder.Decode(&raw); err != nil {
		return ast.Value{}, err
	}

	return fromRaw(raw), nil
}

func fromRaw(raw any) ast.Value {
	switch typed := raw.(type) {
	case nil:
		return ast.NullValue()

	case ast.Value:
		return typed

	case bool:
		return ast.BooleanValue(typed)

	case string:
		return ast.StringValue(typed)

	case json.Number:
		return ast.NumberValue(typed.String())

	case float64:
		return ast.NumberValue(strconv.FormatFloat(typed, 'g', -1, 64))

	case []any:
		items := make([]ast.Value, 0, len(typed))
		for _, entry := range typed {
			items = append(items, fromRaw(entry))
		}
		return ast.ArrayValue(items)

	case map[string]any:
		object := make(map[string]ast.Value, len(typed))
		for key, entry := range typed {
			object[key] = fromRaw(entry)
		}
		return ast.ObjectValue(object)

	default:
		return ast.NullValue()
	}
}

func DecodeInto(value ast.Value, target any) error {
	if target == nil {
		return fmt.Errorf("target is nil")
	}

	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Pointer || targetValue.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer")
	}

	handled, err := decodeDirect(value, targetValue.Elem())
	if handled {
		return err
	}

	data, err := Encode(value)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, target)
}

func decodeDirect(value ast.Value, target reflect.Value) (bool, error) {
	if !target.CanSet() {
		return true, fmt.Errorf("target cannot be set")
	}

	targetType := target.Type()

	if value.IsMissing() || value.IsNull() {
		target.Set(reflect.Zero(targetType))
		return true, nil
	}

	if targetType == reflect.TypeOf(time.Time{}) {
		if value.Kind != ast.KindString {
			return true, fmt.Errorf("expected string")
		}

		parsed, err := time.Parse(time.RFC3339Nano, value.String)
		if err == nil {
			target.Set(reflect.ValueOf(parsed))
			return true, nil
		}

		parsed, err = time.Parse(time.RFC3339, value.String)
		if err == nil {
			target.Set(reflect.ValueOf(parsed))
			return true, nil
		}

		return true, err
	}

	switch target.Kind() {
	case reflect.Pointer:
		holder := reflect.New(targetType.Elem())

		handled, err := decodeDirect(value, holder.Elem())
		if handled {
			if err != nil {
				return true, err
			}
			target.Set(holder)
			return true, nil
		}

		return false, nil

	case reflect.String:
		if value.Kind != ast.KindString {
			return true, fmt.Errorf("expected string")
		}
		target.SetString(value.String)
		return true, nil

	case reflect.Bool:
		if value.Kind != ast.KindBoolean {
			return true, fmt.Errorf("expected boolean")
		}
		target.SetBool(value.Boolean)
		return true, nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if value.Kind != ast.KindNumber {
			return true, fmt.Errorf("expected number")
		}
		parsed, err := ParseInt(value.Number, targetType.Bits())
		if err != nil {
			return true, err
		}
		target.SetInt(parsed)
		return true, nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if value.Kind != ast.KindNumber {
			return true, fmt.Errorf("expected number")
		}
		parsed, err := ParseUint(value.Number, targetType.Bits())
		if err != nil {
			return true, err
		}
		target.SetUint(parsed)
		return true, nil

	case reflect.Float32, reflect.Float64:
		if value.Kind != ast.KindNumber {
			return true, fmt.Errorf("expected number")
		}
		parsed, err := ParseFloat(value.Number, targetType.Bits())
		if err != nil {
			return true, err
		}
		target.SetFloat(parsed)
		return true, nil

	case reflect.Slice:
		if value.Kind != ast.KindArray {
			return true, fmt.Errorf("expected array")
		}

		elementType := targetType.Elem()
		length := len(value.Array)

		sliceValue := reflect.MakeSlice(targetType, length, length)
		for index := 0; index < length; index++ {
			item := value.Array[index]

			handled, err := decodeDirect(item, sliceValue.Index(index))
			if handled {
				if err != nil {
					return true, err
				}
				continue
			}

			holder := reflect.New(elementType)
			data, err := Encode(item)
			if err != nil {
				return true, err
			}
			if err := json.Unmarshal(data, holder.Interface()); err != nil {
				return true, err
			}
			sliceValue.Index(index).Set(holder.Elem())
		}

		target.Set(sliceValue)
		return true, nil

	case reflect.Map:
		if value.Kind != ast.KindObject {
			return true, fmt.Errorf("expected object")
		}

		if targetType.Key().Kind() != reflect.String {
			return false, nil
		}

		elementType := targetType.Elem()
		mapValue := reflect.MakeMapWithSize(targetType, len(value.Object))

		for key, item := range value.Object {
			entry := reflect.New(elementType).Elem()

			handled, err := decodeDirect(item, entry)
			if handled {
				if err != nil {
					return true, err
				}
			} else {
				holder := reflect.New(elementType)
				data, err := Encode(item)
				if err != nil {
					return true, err
				}
				if err := json.Unmarshal(data, holder.Interface()); err != nil {
					return true, err
				}
				entry.Set(holder.Elem())
			}

			mapValue.SetMapIndex(reflect.ValueOf(key), entry)
		}

		target.Set(mapValue)
		return true, nil

	default:
		return false, nil
	}
}

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

func ParseInt(text string, bitSize int) (int64, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0, fmt.Errorf("invalid number")
	}
	value, err := strconv.ParseInt(text, 10, bitSize)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func ParseUint(text string, bitSize int) (uint64, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0, fmt.Errorf("invalid number")
	}
	value, err := strconv.ParseUint(text, 10, bitSize)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func ParseFloat(text string, bitSize int) (float64, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0, fmt.Errorf("invalid number")
	}
	value, err := strconv.ParseFloat(text, bitSize)
	if err != nil {
		return 0, err
	}
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, fmt.Errorf("invalid number")
	}
	return value, nil
}
