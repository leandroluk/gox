package types

import "encoding/json"

// MediaType represents a media type.
type MediaType struct {
	schema   *Schema
	example  any
	examples map[string]*ExampleObject
	encoding map[string]*Encoding
}

func (m *MediaType) Schema(build func(s *Schema)) *MediaType {
	if m.schema == nil {
		m.schema = &Schema{}
	}
	if build != nil {
		build(m.schema)
	}
	return m
}

func (m *MediaType) Example(value any) *MediaType {
	m.example = value
	return m
}

func (m *MediaType) ExampleNamed(name string, build func(e *ExampleObject)) *MediaType {
	if m.examples == nil {
		m.examples = make(map[string]*ExampleObject)
	}
	ex, ok := m.examples[name]
	if !ok {
		ex = &ExampleObject{}
		m.examples[name] = ex
	}
	if build != nil {
		build(ex)
	}
	return m
}

func (m *MediaType) Encoding(name string, build func(e *Encoding)) *MediaType {
	if m.encoding == nil {
		m.encoding = make(map[string]*Encoding)
	}
	enc, ok := m.encoding[name]
	if !ok {
		enc = &Encoding{}
		m.encoding[name] = enc
	}
	if build != nil {
		build(enc)
	}
	return m
}

func (m MediaType) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Schema   *Schema                   `json:"schema,omitempty"`
		Example  any                       `json:"example,omitempty"`
		Examples map[string]*ExampleObject `json:"examples,omitempty"`
		Encoding map[string]*Encoding      `json:"encoding,omitempty"`
	}{
		Schema:   m.schema,
		Example:  m.example,
		Examples: m.examples,
		Encoding: m.encoding,
	})
}

// UnmarshalJSON unmarshals the MediaType from JSON.
func (m *MediaType) UnmarshalJSON(data []byte) error {
	type Alias MediaType
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(m),
	}
	return json.Unmarshal(data, &aux)
}
