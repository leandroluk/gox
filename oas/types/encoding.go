package types

import "encoding/json"

// Encoding represents encoding properties.
type Encoding struct {
	contentType   *string
	headers       map[string]*Header
	style         *string
	explode       *bool
	allowReserved *bool
}

func (e *Encoding) ContentType(value string) *Encoding {
	e.contentType = &value
	return e
}

func (e *Encoding) Header(name string, build func(h *Header)) *Encoding {
	if e.headers == nil {
		e.headers = make(map[string]*Header)
	}
	h, ok := e.headers[name]
	if !ok {
		h = &Header{}
		e.headers[name] = h
	}
	if build != nil {
		build(h)
	}
	return e
}

func (e *Encoding) Style(value string) *Encoding {
	e.style = &value
	return e
}

func (e *Encoding) Explode(value bool) *Encoding {
	e.explode = &value
	return e
}

func (e *Encoding) AllowReserved(value bool) *Encoding {
	e.allowReserved = &value
	return e
}

func (e Encoding) MarshalJSON() ([]byte, error) {
	explode := e.explode

	// Apply default explode if not set
	if explode == nil {
		defaultExplode := true
		explode = &defaultExplode
	}

	return json.Marshal(struct {
		ContentType   *string            `json:"contentType,omitempty"`
		Headers       map[string]*Header `json:"headers,omitempty"`
		Style         *string            `json:"style,omitempty"`
		Explode       *bool              `json:"explode,omitempty"`
		AllowReserved *bool              `json:"allowReserved,omitempty"`
	}{
		ContentType:   e.contentType,
		Headers:       e.headers,
		Style:         e.style,
		Explode:       explode,
		AllowReserved: e.allowReserved,
	})
}

// UnmarshalJSON unmarshals the Encoding from JSON.
func (e *Encoding) UnmarshalJSON(data []byte) error {
	type Alias Encoding
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(e),
	}
	return json.Unmarshal(data, &aux)
}
