package types

import (
	"encoding/json"

	"github.com/leandroluk/gox/oas/enums"
)

// Header represents a header.
type Header struct {
	description *string
	style       *string
	explode     *bool
	schema      *Schema
	content     map[enums.ContentType]*MediaType
	example     any
	examples    map[string]*ExampleObject
	deprecated  *bool
}

func (h *Header) Description(value string) *Header {
	h.description = &value
	return h
}

func (h *Header) Style(value string) *Header {
	h.style = &value
	return h
}

func (h *Header) Explode(value bool) *Header {
	h.explode = &value
	return h
}

func (h *Header) Schema(build func(s *Schema)) *Header {
	if h.schema == nil {
		h.schema = &Schema{}
	}
	if build != nil {
		build(h.schema)
	}
	return h
}

func (h *Header) Content(name enums.ContentType, build func(m *MediaType)) *Header {
	if h.content == nil {
		h.content = make(map[enums.ContentType]*MediaType)
	}
	m, ok := h.content[name]
	if !ok {
		m = &MediaType{}
		h.content[name] = m
	}
	if build != nil {
		build(m)
	}
	return h
}

func (h *Header) Example(value any) *Header {
	h.example = value
	return h
}

func (h *Header) ExampleNamed(name string, build func(e *ExampleObject)) *Header {
	if h.examples == nil {
		h.examples = make(map[string]*ExampleObject)
	}
	ex, ok := h.examples[name]
	if !ok {
		ex = &ExampleObject{}
		h.examples[name] = ex
	}
	if build != nil {
		build(ex)
	}
	return h
}

func (h *Header) Deprecated(value bool) *Header {
	h.deprecated = &value
	return h
}

func (h Header) MarshalJSON() ([]byte, error) {
	style := h.style
	explode := h.explode

	// Apply default style if not set
	if style == nil {
		defaultStyle := "simple"
		style = &defaultStyle
	}

	// Apply default explode if not set
	if explode == nil {
		defaultExplode := false
		explode = &defaultExplode
	}

	return json.Marshal(struct {
		Description *string                          `json:"description,omitempty"`
		Style       *string                          `json:"style,omitempty"`
		Explode     *bool                            `json:"explode,omitempty"`
		Schema      *Schema                          `json:"schema,omitempty"`
		Content     map[enums.ContentType]*MediaType `json:"content,omitempty"`
		Example     any                              `json:"example,omitempty"`
		Examples    map[string]*ExampleObject        `json:"examples,omitempty"`
		Deprecated  *bool                            `json:"deprecated,omitempty"`
	}{
		Description: h.description,
		Style:       style,
		Explode:     explode,
		Schema:      h.schema,
		Content:     h.content,
		Example:     h.example,
		Examples:    h.examples,
		Deprecated:  h.deprecated,
	})
}

// UnmarshalJSON unmarshals the Header from JSON.
func (h *Header) UnmarshalJSON(data []byte) error {
	type Alias Header
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(h),
	}
	return json.Unmarshal(data, &aux)
}
