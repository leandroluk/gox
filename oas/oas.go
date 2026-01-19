// Package oas provides type-safe OpenAPI 3.1 document construction with a fluent API.
//
// This package offers a comprehensive, ergonomic way to build OpenAPI specifications
// programmatically in Go, featuring:
//
//   - 100% fluent API with void callbacks for hierarchical construction
//   - Type-safe enumerations for common values (content types, schema types)
//   - Automatic validation of required fields during JSON marshaling
//   - Full support for both marshaling (to JSON) and unmarshaling (from JSON)
//   - NestJS-style security helpers for common authentication patterns
//   - Schema type helpers for reduced verbosity (.String(), .Object(), etc.)
//   - Content type helpers for common media types (.Json(), .Xml(), etc.)
//
// # Basic Usage
//
// Create a simple API document:
//
//	doc := types.New().
//	    Info(func(i *types.Info) {
//	        i.Title("My API").Version("1.0.0")
//	    }).
//	    Path("/users", func(p *types.Path) {
//	        p.Get(func(o *types.Operation) {
//	            o.Summary("List users").
//	              Response("200", func(r *types.Response) {
//	                  r.Description("Success")
//	              })
//	        })
//	    })
//
// # Security Helpers
//
// Register security schemes and apply them to operations:
//
//	doc.WithBearerToken("bearer").
//	    WithApiKey("api_key", "header")
//
//	operation.UseBearerToken("bearer", "read", "write")
//
// # Schema Type Helpers
//
// Use type helpers for concise schema definitions:
//
//	schema.String().MinLength(3).MaxLength(100)
//	schema.Object().Required("id", "name")
//	schema.Array().Items(func(s *Schema) { s.String() })
//
// # Validation
//
// All types automatically validate required fields during marshaling:
//
//	data, err := json.Marshal(doc)
//	// Returns error if required fields are missing
//
// # Architecture
//
// The package is organized into sub-packages:
//
//   - types: Core OpenAPI types with fluent builders
//   - enums: Type-safe enumerations (ContentType, SchemaType)
//   - oas (root): Type aliases for backward compatibility
//
// For detailed examples and API reference, see the README.md file.
package oas

import "github.com/leandroluk/go/oas/types"

// Type aliases for backward compatibility and cleaner imports.
// All core functionality is implemented in the types sub-package.

// Document represents an OpenAPI 3.1 document.
type Document = types.Document

// Path represents an OAS path item (endpoints at a specific path).
type Path = types.Path

// Operation represents an API operation (GET, POST, etc.).
type Operation = types.Operation

// Components holds reusable objects for different aspects of the OAS.
type Components = types.Components

// SecurityRequirement maps security scheme names to required scopes.
type SecurityRequirement = types.SecurityRequirement

// SecurityRequirements is a list of security requirement alternatives.
type SecurityRequirements = types.SecurityRequirements

// Callback represents a map of possible out-of-band callbacks related to the parent operation.
type Callback = types.Callback

// Callbacks represents a map of named callbacks.
type Callbacks = types.Callbacks

// Schema represents an OpenAPI schema.
type Schema = types.Schema

// New creates a new OpenAPI document with default values.
// The document will have openapi set to "3.1.0" by default when marshaled.
func New() *Document {
	return types.New()
}

// String creates a new string schema.
func String() *Schema {
	return (&Schema{}).String()
}

// Integer creates a new integer schema.
func Integer() *Schema {
	return (&Schema{}).Integer()
}

// Number creates a new number schema.
func Number() *Schema {
	return (&Schema{}).Number()
}

// Boolean creates a new boolean schema.
func Boolean() *Schema {
	return (&Schema{}).Boolean()
}

// Object creates a new object schema.
func Object() *Schema {
	return (&Schema{}).Object()
}

// Array creates a new array schema.
func Array() *Schema {
	return (&Schema{}).Array()
}

// Null creates a new null schema.
func Null() *Schema {
	return (&Schema{}).Null()
}
