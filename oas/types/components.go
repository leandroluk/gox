package types

import (
	"encoding/json"
)

// Callback map of paths.
type Callback map[string]*Path

// Callbacks map of callbacks.
type Callbacks map[string]Callback

// Components represents reusable components.
type Components struct {
	schemas         map[string]*Schema
	responses       map[string]*Response
	parameters      map[string]*Parameter
	examples        map[string]*ExampleObject
	requestBodies   map[string]*RequestBody
	headers         map[string]*Header
	securitySchemes map[string]*SecurityScheme
	links           map[string]*Link
	callbacks       Callbacks
	pathItems       map[string]*Path
}

func (c *Components) Schema(name string, build func(s *Schema)) *Components {
	if c.schemas == nil {
		c.schemas = make(map[string]*Schema)
	}
	s, ok := c.schemas[name]
	if !ok {
		s = &Schema{}
		c.schemas[name] = s
	}
	if build != nil {
		build(s)
	}
	return c
}

func (c *Components) Response(name string, build func(r *Response)) *Components {
	if c.responses == nil {
		c.responses = make(map[string]*Response)
	}
	r, ok := c.responses[name]
	if !ok {
		r = &Response{}
		c.responses[name] = r
	}
	if build != nil {
		build(r)
	}
	return c
}

func (c *Components) Parameter(name string, build func(p *Parameter)) *Components {
	if c.parameters == nil {
		c.parameters = make(map[string]*Parameter)
	}
	p, ok := c.parameters[name]
	if !ok {
		p = &Parameter{}
		c.parameters[name] = p
	}
	if build != nil {
		build(p)
	}
	return c
}

func (c *Components) Example(name string, build func(e *ExampleObject)) *Components {
	if c.examples == nil {
		c.examples = make(map[string]*ExampleObject)
	}
	e, ok := c.examples[name]
	if !ok {
		e = &ExampleObject{}
		c.examples[name] = e
	}
	if build != nil {
		build(e)
	}
	return c
}

func (c *Components) RequestBody(name string, build func(r *RequestBody)) *Components {
	if c.requestBodies == nil {
		c.requestBodies = make(map[string]*RequestBody)
	}
	r, ok := c.requestBodies[name]
	if !ok {
		r = &RequestBody{}
		c.requestBodies[name] = r
	}
	if build != nil {
		build(r)
	}
	return c
}

func (c *Components) Header(name string, build func(h *Header)) *Components {
	if c.headers == nil {
		c.headers = make(map[string]*Header)
	}
	h, ok := c.headers[name]
	if !ok {
		h = &Header{}
		c.headers[name] = h
	}
	if build != nil {
		build(h)
	}
	return c
}

func (c *Components) SecurityScheme(name string, build func(s *SecurityScheme)) *Components {
	if c.securitySchemes == nil {
		c.securitySchemes = make(map[string]*SecurityScheme)
	}
	s, ok := c.securitySchemes[name]
	if !ok {
		s = &SecurityScheme{}
		c.securitySchemes[name] = s
	}
	if build != nil {
		build(s)
	}
	return c
}

func (c *Components) Link(name string, build func(l *Link)) *Components {
	if c.links == nil {
		c.links = make(map[string]*Link)
	}
	l, ok := c.links[name]
	if !ok {
		l = &Link{}
		c.links[name] = l
	}
	if build != nil {
		build(l)
	}
	return c
}

func (c *Components) Callback(name string, build func(cb Callback)) *Components {
	if c.callbacks == nil {
		c.callbacks = make(Callbacks)
	}
	cb, ok := c.callbacks[name]
	if !ok {
		cb = make(Callback)
		c.callbacks[name] = cb
	}
	if build != nil {
		build(cb)
	}
	return c
}

func (c *Components) Path(name string, build func(p *Path)) *Components {
	if c.pathItems == nil {
		c.pathItems = make(map[string]*Path)
	}
	p, ok := c.pathItems[name]
	if !ok {
		p = &Path{}
		c.pathItems[name] = p
	}
	if build != nil {
		build(p)
	}
	return c
}

// MarshalJSON marshals the Components to JSON.
func (c Components) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Schemas         map[string]*Schema         `json:"schemas,omitempty"`
		Responses       map[string]*Response       `json:"responses,omitempty"`
		Parameters      map[string]*Parameter      `json:"parameters,omitempty"`
		Examples        map[string]*ExampleObject  `json:"examples,omitempty"`
		RequestBodies   map[string]*RequestBody    `json:"requestBodies,omitempty"`
		Headers         map[string]*Header         `json:"headers,omitempty"`
		SecuritySchemes map[string]*SecurityScheme `json:"securitySchemes,omitempty"`
		Links           map[string]*Link           `json:"links,omitempty"`
		Callbacks       Callbacks                  `json:"callbacks,omitempty"`
		PathItems       map[string]*Path           `json:"pathItems,omitempty"`
	}{
		Schemas:         c.schemas,
		Responses:       c.responses,
		Parameters:      c.parameters,
		Examples:        c.examples,
		RequestBodies:   c.requestBodies,
		Headers:         c.headers,
		SecuritySchemes: c.securitySchemes,
		Links:           c.links,
		Callbacks:       c.callbacks,
		PathItems:       c.pathItems,
	})
}

// UnmarshalJSON unmarshals the Components from JSON.
func (c *Components) UnmarshalJSON(data []byte) error {
	type Alias Components
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	return json.Unmarshal(data, &aux)
}
