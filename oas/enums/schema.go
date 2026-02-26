package enums

// SchemaType represents JSON Schema type identifiers used in OpenAPI 3.1.
//
// These correspond to the "type" field in a schema object and define the
// data type of the value.
//
// In OpenAPI 3.1, the type field can be:
//   - A single string value (e.g., "string", "object")
//   - An array of string values (e.g., ["string", "null"] for nullable types)
//
// Example usage:
//
//	schema.Type(enums.SchemaString)        // Simple type
//	schema.Type(enums.SchemaString, enums.SchemaNull)  // Nullable string
//
// Or using type helpers:
//
//	schema.String()   // Equivalent to .Type(enums.SchemaString)
//	schema.Object()   // Equivalent to .Type(enums.SchemaObject)
type SchemaType string

// JSON Schema type constants as defined in OpenAPI 3.1.
const (
	// SchemaOBJECT represents an object/map type.
	// Objects have properties and can have additional constraints like required fields.
	SchemaOBJECT SchemaType = "object"

	// SchemaSTRING represents a string type.
	// Can have constraints like minLength, maxLength, pattern, and format.
	SchemaSTRING SchemaType = "string"

	// SchemaINTEGER represents an integer type (whole numbers).
	// Can have constraints like minimum, maximum, and multipleOf.
	SchemaINTEGER SchemaType = "integer"

	// SchemaNUMBER represents a number type (includes decimals).
	// Can have constraints like minimum, maximum, and multipleOf.
	SchemaNUMBER SchemaType = "number"

	// SchemaBOOLEAN represents a boolean type (true/false).
	SchemaBOOLEAN SchemaType = "boolean"

	// SchemaARRAY represents an array/list type.
	// Must specify items schema and can have minItems, maxItems, uniqueItems.
	SchemaARRAY SchemaType = "array"

	// SchemaNULL represents a null value.
	// Often combined with other types for nullable fields, e.g., ["string", "null"].
	SchemaNULL SchemaType = "null"
)
