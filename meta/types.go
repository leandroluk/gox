package meta

import "reflect"

// Enumerable can be implemented by enum types to expose their valid values.
// The meta package detects this automatically when registering a field.
type Enumerable interface {
	Values() []string
}

// ExternalDocs represents external documentation for an object or field.
type ExternalDocs struct {
	Description string
	URL         string
}

// ObjectMetadata holds all documentation and constraints for a specific struct type.
type ObjectMetadata struct {
	Title        string
	Description  string
	Deprecated   bool
	ExternalDocs *ExternalDocs
	Throws       []ThrowsMetadata
	Fields       map[string]*FieldMetadata
	Required     []string // Automatically populated from fields
	Example      any
	Type         reflect.Type
}

// FieldMetadata holds documentation for a specific field within a struct.
type FieldMetadata struct {
	Description  string
	Example      any
	Type         reflect.Type
	JSONName     string
	Format       string
	Nullable     bool
	Required     bool
	ReadOnly     bool
	WriteOnly    bool
	Deprecated   bool
	Enum         []string
	ExternalDocs *ExternalDocs

	// Constraints
	Min        *float64
	Max        *float64
	MultipleOf *float64
	MinLength  *int
	MaxLength  *int
	Pattern    string
	MinItems   *int
	MaxItems   *int
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
