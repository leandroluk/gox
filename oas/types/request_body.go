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

func (r *RequestBody) Content(name enums.ContentType, build func(m *MediaType)) *RequestBody {
	if r.content == nil {
		r.content = make(map[enums.ContentType]*MediaType)
	}
	m, ok := r.content[name]
	if !ok {
		m = &MediaType{}
		r.content[name] = m
	}
	if build != nil {
		build(m)
	}
	return r
}

// Content type helpers
func (r *RequestBody) Json(build func(m *MediaType)) *RequestBody {
	return r.Content(enums.ContentJson, build)
}

func (r *RequestBody) Xml(build func(m *MediaType)) *RequestBody {
	return r.Content(enums.ContentXml, build)
}

func (r *RequestBody) Form(build func(m *MediaType)) *RequestBody {
	return r.Content(enums.ContentForm, build)
}

func (r *RequestBody) Multipart(build func(m *MediaType)) *RequestBody {
	return r.Content(enums.ContentMultipart, build)
}

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
