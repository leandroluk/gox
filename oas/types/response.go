package types

import (
	"encoding/json"

	"github.com/leandroluk/gox/oas/enums"
)

// Response represents a response.
type Response struct {
	description string
	content     map[enums.ContentType]*MediaType
	headers     map[string]*Header
	links       map[string]*Link
}

func (r *Response) Description(value string) *Response {
	r.description = value
	return r
}

func (r *Response) Content(name enums.ContentType, build func(m *MediaType)) *Response {
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
func (r *Response) Json(build func(m *MediaType)) *Response {
	return r.Content(enums.ContentJson, build)
}

func (r *Response) Xml(build func(m *MediaType)) *Response {
	return r.Content(enums.ContentXml, build)
}

func (r *Response) Html(build func(m *MediaType)) *Response {
	return r.Content(enums.ContentHtml, build)
}

func (r *Response) Plain(build func(m *MediaType)) *Response {
	return r.Content(enums.ContentPlain, build)
}

func (r *Response) Header(name string, build func(h *Header)) *Response {
	if r.headers == nil {
		r.headers = make(map[string]*Header)
	}
	h, ok := r.headers[name]
	if !ok {
		h = &Header{}
		r.headers[name] = h
	}
	if build != nil {
		build(h)
	}
	return r
}

func (r *Response) Link(name string, build func(l *Link)) *Response {
	if r.links == nil {
		r.links = make(map[string]*Link)
	}
	l, ok := r.links[name]
	if !ok {
		l = &Link{}
		r.links[name] = l
	}
	if build != nil {
		build(l)
	}
	return r
}

func (r Response) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Description string                           `json:"description"`
		Content     map[enums.ContentType]*MediaType `json:"content,omitempty"`
		Headers     map[string]*Header               `json:"headers,omitempty"`
		Links       map[string]*Link                 `json:"links,omitempty"`
	}{
		Description: r.description,
		Content:     r.content,
		Headers:     r.headers,
		Links:       r.links,
	})
}

// UnmarshalJSON unmarshals the Response from JSON.
func (r *Response) UnmarshalJSON(data []byte) error {
	type Alias Response
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	return json.Unmarshal(data, &aux)
}
