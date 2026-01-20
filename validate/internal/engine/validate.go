// internal/engine/validate.go
package engine

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/leandroluk/go/validate/internal/ast"
	"github.com/leandroluk/go/validate/internal/codec"
	"github.com/leandroluk/go/validate/internal/reflection"
	"github.com/leandroluk/go/validate/schema"
)

func Run[T any](walk func(context *Context, value ast.Value) (T, bool), input any, optionList ...schema.Option) (T, error) {
	options := schema.ApplyOptions(optionList...)
	context := NewContext(options)

	value, err := InputToASTWithOptions(input, options)
	if err != nil {
		var zero T
		return zero, err
	}

	context.SetRoot(value)

	output, _ := walk(context, value)
	return output, context.Error()
}

func InputToAST(input any) (ast.Value, error) {
	return InputToASTWithOptions(input, schema.DefaultOptions())
}

func InputToASTWithOptions(input any, options schema.Options) (ast.Value, error) {
	switch typed := input.(type) {
	case nil:
		return ast.NullValue(), nil

	case ast.Value:
		return typed, nil

	case *ast.Value:
		if typed == nil {
			return ast.NullValue(), nil
		}
		return *typed, nil

	case json.RawMessage:
		return codec.Decode([]byte(typed))

	case []byte:
		return codec.Decode(typed)

	default:
		return ReflectToASTWithOptions(input, options)
	}
}

func ReflectToAST(input any) (ast.Value, error) {
	return ReflectToASTWithOptions(input, schema.DefaultOptions())
}

func ReflectToASTWithOptions(input any, options schema.Options) (ast.Value, error) {
	if input == nil {
		return ast.NullValue(), nil
	}
	value := reflect.ValueOf(input)
	return reflectValueToAST(value, options)
}

func reflectValueToAST(value reflect.Value, options schema.Options) (ast.Value, error) {
	if !value.IsValid() {
		return ast.NullValue(), nil
	}

	for value.Kind() == reflect.Interface {
		if value.IsNil() {
			return ast.NullValue(), nil
		}
		value = value.Elem()
	}

	checkValue := value

	if checkValue.Kind() == reflect.Pointer {
		if checkValue.IsNil() {
			return ast.NullValue(), nil
		}
	}

	if options.OmitEmpty {
		if reflection.IsDefaultValue(checkValue) || reflection.IsLengthZero(checkValue) {
			return ast.NullValue(), nil
		}
	}

	if options.OmitZero && reflection.IsDefaultValue(checkValue) {
		return ast.NullValue(), nil
	}

	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return ast.NullValue(), nil
		}
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.String:
		return ast.StringValue(value.String()), nil

	case reflect.Bool:
		return ast.BooleanValue(value.Bool()), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return ast.NumberValue(strconv.FormatInt(value.Int(), 10)), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return ast.NumberValue(strconv.FormatUint(value.Uint(), 10)), nil

	case reflect.Float32, reflect.Float64:
		return ast.NumberValue(strconv.FormatFloat(value.Float(), 'g', -1, 64)), nil

	case reflect.Slice, reflect.Array:
		if value.Kind() == reflect.Slice && value.IsNil() {
			return ast.NullValue(), nil
		}
		if options.OmitEmpty && value.Len() == 0 {
			return ast.NullValue(), nil
		}
		length := value.Len()
		items := make([]ast.Value, 0, length)
		for index := 0; index < length; index++ {
			item, err := reflectValueToAST(value.Index(index), options)
			if err != nil {
				return ast.Value{}, err
			}
			items = append(items, item)
		}
		return ast.ArrayValue(items), nil

	case reflect.Map:
		if value.IsNil() {
			return ast.NullValue(), nil
		}
		if options.OmitEmpty && value.Len() == 0 {
			return ast.NullValue(), nil
		}
		if value.Type().Key().Kind() != reflect.String {
			return ast.Value{}, fmt.Errorf("unsupported map key type: %s", value.Type().Key().String())
		}

		keys := value.MapKeys()
		object := make(map[string]ast.Value, len(keys))
		for _, key := range keys {
			entryValue := value.MapIndex(key)
			if shouldOmitField(entryValue, options, false) {
				continue
			}
			entry, err := reflectValueToAST(entryValue, options)
			if err != nil {
				return ast.Value{}, err
			}
			object[key.String()] = entry
		}
		return ast.ObjectValue(object), nil

	case reflect.Struct:
		if value.Type() == reflect.TypeOf(time.Time{}) {
			timeValue := value.Interface().(time.Time)
			return ast.StringValue(timeValue.Format(time.RFC3339Nano)), nil
		}

		object := make(map[string]ast.Value)
		typeValue := value.Type()
		for index := 0; index < typeValue.NumField(); index++ {
			field := typeValue.Field(index)
			if field.PkgPath != "" {
				continue
			}

			tag := reflection.ParseJSONTag(field.Tag.Get("json"))
			if tag.Ignored {
				continue
			}

			name := field.Name
			if tag.HasTag && tag.Name != "" {
				name = tag.Name
			}
			if name == "" {
				continue
			}

			fieldValue := value.Field(index)
			if shouldOmitField(fieldValue, options, tag.OmitEmpty) {
				continue
			}

			entry, err := reflectValueToAST(fieldValue, options)
			if err != nil {
				return ast.Value{}, err
			}
			object[name] = entry
		}
		return ast.ObjectValue(object), nil

	default:
		return ast.NullValue(), nil
	}
}

func shouldOmitField(value reflect.Value, options schema.Options, localOmitEmpty bool) bool {
	unwrapped, nilLike := reflection.UnwrapInterface(value)

	if options.OmitNil && (nilLike || reflection.IsNilLike(unwrapped)) {
		return true
	}

	effectiveOmitEmpty := options.OmitEmpty || localOmitEmpty
	if effectiveOmitEmpty {
		if nilLike || reflection.IsNilLike(unwrapped) {
			return true
		}
		if reflection.IsDefaultValue(unwrapped) {
			return true
		}
		if reflection.IsLengthZero(unwrapped) {
			return true
		}
	}

	if options.OmitZero && reflection.IsDefaultValue(unwrapped) {
		return true
	}

	return false
}

func jsonFieldName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "" {
		return field.Name
	}

	head, _, _ := strings.Cut(tag, ",")
	if head == "-" {
		return ""
	}
	if head == "" {
		return field.Name
	}
	return head
}
