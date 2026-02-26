package types

import (
	"encoding/json"

	"github.com/leandroluk/gox/oas/enums"
)

// RequestBody represents a request body.
type RequestBody struct {
	description *string
	content     map[enums.ContentType]*MediaType
	required    *bool
}

func (r *RequestBody) Description(value string) *RequestBody {
	r.description = &value
	return r
}

func (r *RequestBody) Content(name enums.ContentType, b MediaTypeFn) *RequestBody {
	if r.content == nil {
		r.content = make(map[enums.ContentType]*MediaType)
	}
	m, ok := r.content[name]
	if !ok {
		m = &MediaType{}
		r.content[name] = m
	}
	if b != nil {
		b(m)
	}
	return r
}

func (r *RequestBody) Json(b MediaTypeFn) *RequestBody  { return r.Content(enums.ContentJSON, b) }
func (r *RequestBody) Xml(b MediaTypeFn) *RequestBody   { return r.Content(enums.ContentXML, b) }
func (r *RequestBody) Form(b MediaTypeFn) *RequestBody  { return r.Content(enums.ContentFORM, b) }
func (r *RequestBody) Multi(b MediaTypeFn) *RequestBody { return r.Content(enums.ContentMULTI, b) }
func (r *RequestBody) Html(b MediaTypeFn) *RequestBody  { return r.Content(enums.ContentHTML, b) }
func (r *RequestBody) Plain(b MediaTypeFn) *RequestBody { return r.Content(enums.ContentPLAIN, b) }
func (r *RequestBody) Csv(b MediaTypeFn) *RequestBody   { return r.Content(enums.ContentCSV, b) }
func (r *RequestBody) Jpeg(b MediaTypeFn) *RequestBody  { return r.Content(enums.ContentJPEG, b) }
func (r *RequestBody) Png(b MediaTypeFn) *RequestBody   { return r.Content(enums.ContentPNG, b) }
func (r *RequestBody) Gif(b MediaTypeFn) *RequestBody   { return r.Content(enums.ContentGIF, b) }
func (r *RequestBody) Svg(b MediaTypeFn) *RequestBody   { return r.Content(enums.ContentSVG, b) }
func (r *RequestBody) Webp(b MediaTypeFn) *RequestBody  { return r.Content(enums.ContentWEBP, b) }

func (r *RequestBody) Required(value bool) *RequestBody {
	r.required = &value
	return r
}

func (r RequestBody) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Description *string                          `json:"description,omitempty"`
		Content     map[enums.ContentType]*MediaType `json:"content"`
		Required    *bool                            `json:"required,omitempty"`
	}{
		Description: r.description,
		Content:     r.content,
		Required:    r.required,
	})
}

// UnmarshalJSON unmarshals the RequestBody from JSON.
func (r *RequestBody) UnmarshalJSON(data []byte) error {
	type Alias RequestBody
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	return json.Unmarshal(data, &aux)
}
