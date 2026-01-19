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
	// ContentJson represents application/json content type.
	// Used for JSON-encoded request/response bodies.
	ContentJson ContentType = "application/json"

	// ContentXml represents text/xml content type.
	// Used for XML-encoded data.
	ContentXml ContentType = "text/xml"

	// ContentHtml represents text/html content type.
	// Used for HTML responses.
	ContentHtml ContentType = "text/html"

	// ContentPlain represents text/plain content type.
	// Used for plain text data.
	ContentPlain ContentType = "text/plain"

	// ContentCsv represents text/csv content type.
	// Used for CSV-formatted data.
	ContentCsv ContentType = "text/csv"

	// ContentForm represents application/x-www-form-urlencoded content type.
	// Used for standard HTML form submissions.
	ContentForm ContentType = "application/x-www-form-urlencoded"

	// ContentMultipart represents multipart/form-data content type.
	// Used for file uploads and complex form data.
	ContentMultipart ContentType = "multipart/form-data"

	// ContentJpeg represents image/jpeg content type.
	ContentJpeg ContentType = "image/jpeg"

	// ContentPng represents image/png content type.
	ContentPng ContentType = "image/png"

	// ContentGif represents image/gif content type.
	ContentGif ContentType = "image/gif"

	// ContentSvg represents image/svg+xml content type.
	ContentSvg ContentType = "image/svg+xml"

	// ContentWebp represents image/webp content type.
	ContentWebp ContentType = "image/webp"
)
