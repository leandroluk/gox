package types

import (
	"encoding/json"
	"fmt"

	"github.com/leandroluk/go/oas/enums"
)

// Parameter represents an OpenAPI parameter.
type Parameter struct {
	name            string
	in              string
	description     *string
	required        *bool
	schema          *Schema
	content         map[enums.ContentType]*MediaType
	style           *string
	explode         *bool
	example         any
	examples        map[string]*ExampleObject
	deprecated      *bool
	allowEmptyValue *bool
	allowReserved   *bool
}

// Name sets the parameter name.
func (p *Parameter) Name(value string) *Parameter {
	p.name = value
	return p
}

// In sets the parameter location (query, header, path, cookie).
func (p *Parameter) In(value string) *Parameter {
	p.in = value
	return p
}

// Description sets the parameter description.
func (p *Parameter) Description(value string) *Parameter {
	p.description = &value
	return p
}

// Required sets whether the parameter is required.
func (p *Parameter) Required(value bool) *Parameter {
	p.required = &value
	return p
}

// Schema sets the schema for the parameter.
// Assuming Schema is also refactored to types package, potentially as *Schema
// Since Schema is in the list to be refactored, we can use *Schema here.
// However, Schema type definition might not exist in `types` package yet.
// Since we are adding files one by one, `Schema` struct needs to be defined to avoid compilation error.
// Code below assumes `Schema` will be defined in `types` package.
func (p *Parameter) Schema(build func(s *Schema)) *Parameter {
	if p.schema == nil {
		p.schema = &Schema{}
	}
	if build != nil {
		build(p.schema)
	}
	return p
}

// Content adds a media type content.
func (p *Parameter) Content(name enums.ContentType, build func(m *MediaType)) *Parameter {
	if p.content == nil {
		p.content = make(map[enums.ContentType]*MediaType)
	}
	m, ok := p.content[name]
	if !ok {
		m = &MediaType{}
		p.content[name] = m
	}
	if build != nil {
		build(m)
	}
	return p
}

// Validate validates the Parameter.
func (p Parameter) Validate() error {
	if p.name == "" {
		return fmt.Errorf("Parameter.name is required")
	}
	if p.in == "" {
		return fmt.Errorf("Parameter.in is required")
	}
	return nil
}

// MarshalJSON marshals the Parameter to JSON.
func (p Parameter) MarshalJSON() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	style := p.style
	explode := p.explode

	// Apply default style based on parameter location if not set
	if style == nil {
		var defaultStyle string
		switch p.in {
		case "query", "cookie":
			defaultStyle = "form"
		case "path", "header":
			defaultStyle = "simple"
		default:
			defaultStyle = "simple"
		}
		style = &defaultStyle
	}

	// Apply default explode based on style if not set
	if explode == nil {
		defaultExplode := *style == "form"
		explode = &defaultExplode
	}

	return json.Marshal(struct {
		Name            string                           `json:"name"`
		In              string                           `json:"in"`
		Description     *string                          `json:"description,omitempty"`
		Required        *bool                            `json:"required,omitempty"`
		Schema          *Schema                          `json:"schema,omitempty"`
		Content         map[enums.ContentType]*MediaType `json:"content,omitempty"`
		Style           *string                          `json:"style,omitempty"`
		Explode         *bool                            `json:"explode,omitempty"`
		Example         any                              `json:"example,omitempty"`
		Examples        map[string]*ExampleObject        `json:"examples,omitempty"`
		Deprecated      *bool                            `json:"deprecated,omitempty"`
		AllowEmptyValue *bool                            `json:"allowEmptyValue,omitempty"`
		AllowReserved   *bool                            `json:"allowReserved,omitempty"`
	}{
		Name:            p.name,
		In:              p.in,
		Description:     p.description,
		Required:        p.required,
		Schema:          p.schema,
		Content:         p.content,
		Style:           style,
		Explode:         explode,
		Example:         p.example,
		Examples:        p.examples,
		Deprecated:      p.deprecated,
		AllowEmptyValue: p.allowEmptyValue,
		AllowReserved:   p.allowReserved,
	})
}

// UnmarshalJSON unmarshals the Parameter from JSON.
func (p *Parameter) UnmarshalJSON(data []byte) error {
	type Alias Parameter
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(p),
	}
	return json.Unmarshal(data, &aux)
}
