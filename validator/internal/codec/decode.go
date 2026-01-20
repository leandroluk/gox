// internal/codec/decode.go
package codec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/leandroluk/go/validator/internal/ast"
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
