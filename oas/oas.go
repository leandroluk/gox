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

// Callback map of paths.
type Callback = types.Callback

// Callbacks map of callbacks.
type Callbacks = types.Callbacks

// Components represents reusable components.
type Components = types.Components

// Contact represents contact information for the API.
type Contact = types.Contact

// Discriminator represents a discriminator.
type Discriminator = types.Discriminator

// SecurityRequirement represents a security requirement.
type SecurityRequirement = types.SecurityRequirement

// SecurityRequirements represents a list of security requirements.
type SecurityRequirements = types.SecurityRequirements

// Document represents an OpenAPI document.
type Document = types.Document

// Encoding represents encoding properties.
type Encoding = types.Encoding

// ExampleObject represents an example.
type ExampleObject = types.ExampleObject

// ExternalDocs represents external documentation.
type ExternalDocs = types.ExternalDocs

// Header represents a header.
type Header = types.Header

// Info represents the metadata for the API.
type Info = types.Info

// License represents license information for the API.
type License = types.License

// Link represents a link.
type Link = types.Link

// MediaType represents a media type.
type MediaType = types.MediaType

// OAuthFlow represents an OAuth flow.
type OAuthFlow = types.OAuthFlow

// Operation represents an API operation.
type Operation = types.Operation

// Parameter represents an OpenAPI parameter.
type Parameter = types.Parameter

// Path represents an OAS path item (endpoints at a specific path).
type Path = types.Path

// RequestBody represents a request body.
type RequestBody = types.RequestBody

// ResponseWithCode wraps a Response with its status code.
// This allows a fluent API where the status code is attached to the response object
// before being added to the Operation.
type ResponseWithCode = types.ResponseWithCode

// Response represents a response.
type Response = types.Response

// Schema represents an OAS 3.1 schema.
type Schema = types.Schema

// OAuthFlows represents the configuration for the supported OAuth Flows.
type OAuthFlows = types.OAuthFlows

// SecurityScheme represents a security scheme.
type SecurityScheme = types.SecurityScheme

// ServerVariable represents a server variable for server URL template substitution.
type ServerVariable = types.ServerVariable

// Server represents a server.
type Server = types.Server

// Tag represents a tag.
type Tag = types.Tag

// Xml represents XML metadata.
type Xml = types.Xml

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
