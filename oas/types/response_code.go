package types

import "strconv"

// ResponseWithCode wraps a Response with its status code.
// This allows a fluent API where the status code is attached to the response object
// before being added to the Operation.
type ResponseWithCode struct {
	Code string
	*Response
}

// Description sets the description of the response.
func (r *ResponseWithCode) Description(value string) *ResponseWithCode {
	r.Response.Description(value)
	return r
}

// ContentJSON sets the content type to application/json with the given schema.
func (r *ResponseWithCode) ContentJSON(schema *Schema) *ResponseWithCode {
	r.Response.Json(func(m *MediaType) {
		m.Schema(func(s *Schema) {
			*s = *schema
		})
	})
	return r
}

// NewResponseCode creates a new ResponseWithCode.
func NewResponseCode(code int) *ResponseWithCode {
	return &ResponseWithCode{
		Code:     strconv.Itoa(code),
		Response: &Response{},
	}
}
