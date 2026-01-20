// internal/ast/value.go
package ast

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
