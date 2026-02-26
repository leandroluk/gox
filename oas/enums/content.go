// Package enums provides type-safe enumerations for OpenAPI 3.1 specifications.
//
// This package defines strongly-typed constants for commonly used values in OpenAPI,
// preventing typos and enabling compile-time validation.
package enums

// ContentType represents MIME content type identifiers used in OpenAPI.
//
// These are commonly used in requestBody.content and response.content fields
// to specify the format of data being sent or received.
//
// Example usage:
//
//	response.Content(enums.ContentJson, func(m *MediaType) {
//	    m.Schema(func(s *Schema) { s.Object() })
//	})
type ContentType string

// Common content type constants for HTTP requests and responses.
const (
	// ContentJSON represents application/json content type.
	// Used for JSON-encoded request/response bodies.
	ContentJSON ContentType = "application/json"

	// ContentXML represents text/xml content type.
	// Used for XML-encoded data.
	ContentXML ContentType = "text/xml"

	// ContentHTML represents text/html content type.
	// Used for HTML responses.
	ContentHTML ContentType = "text/html"

	// ContentPlain represents text/plain content type.
	// Used for plain text data.
	ContentPLAIN ContentType = "text/plain"

	// ContentCSV represents text/csv content type.
	// Used for CSV-formatted data.
	ContentCSV ContentType = "text/csv"

	// ContentFORM represents application/x-www-form-urlencoded content type.
	// Used for standard HTML form submissions.
	ContentFORM ContentType = "application/x-www-form-urlencoded"

	// ContentMULTI represents multipart/form-data content type.
	// Used for file uploads and complex form data.
	ContentMULTI ContentType = "multipart/form-data"

	// ContentJPEG represents image/jpeg content type.
	ContentJPEG ContentType = "image/jpeg"

	// ContentPNG represents image/png content type.
	ContentPNG ContentType = "image/png"

	// ContentGIF represents image/gif content type.
	ContentGIF ContentType = "image/gif"

	// ContentSVG represents image/svg+xml content type.
	ContentSVG ContentType = "image/svg+xml"

	// ContentWEBP represents image/webp content type.
	ContentWEBP ContentType = "image/webp"
)
