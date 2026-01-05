package meta

import "reflect"

// ObjectMetadata holds all documentation and constraints for a specific struct type.
type ObjectMetadata struct {
	Description string
	Throws      []ThrowsMetadata
	Fields      map[string]*FieldMetadata
	Example     any
	Type        reflect.Type
}

// FieldMetadata holds documentation for a specific field within a struct.
type FieldMetadata struct {
	Description string
	Example     any
	Type        reflect.Type
	Nullable    bool
}

// ThrowsMetadata represents a potential error that an object or method might return.
type ThrowsMetadata struct {
	ErrorType   reflect.Type
	Description string
}

// ObjectOption defines the interface for decorators that apply to the whole struct.
type ObjectOption interface {
	applyToObject(structPointer any, metadata *ObjectMetadata)
}

// FieldOption defines the interface for decorators that apply to a specific field.
type FieldOption interface {
	applyToField(fieldMetadata *FieldMetadata)
}
