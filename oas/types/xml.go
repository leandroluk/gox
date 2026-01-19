package types

import "encoding/json"

// Xml represents XML metadata.
type Xml struct {
	name      *string
	namespace *string
	prefix    *string
	attribute *bool
	wrapped   *bool
}

// Name sets the name.
func (x *Xml) Name(value string) *Xml {
	x.name = &value
	return x
}

// Namespace sets the namespace.
func (x *Xml) Namespace(value string) *Xml {
	x.namespace = &value
	return x
}

// Prefix sets the prefix.
func (x *Xml) Prefix(value string) *Xml {
	x.prefix = &value
	return x
}

// Attribute sets whether it is an attribute.
func (x *Xml) Attribute(value bool) *Xml {
	x.attribute = &value
	return x
}

// Wrapped sets whether it is wrapped.
func (x *Xml) Wrapped(value bool) *Xml {
	x.wrapped = &value
	return x
}

// MarshalJSON marshals the Xml to JSON.
func (x Xml) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name      *string `json:"name,omitempty"`
		Namespace *string `json:"namespace,omitempty"`
		Prefix    *string `json:"prefix,omitempty"`
		Attribute *bool   `json:"attribute,omitempty"`
		Wrapped   *bool   `json:"wrapped,omitempty"`
	}{
		Name:      x.name,
		Namespace: x.namespace,
		Prefix:    x.prefix,
		Attribute: x.attribute,
		Wrapped:   x.wrapped,
	})
}

// UnmarshalJSON unmarshals the Xml from JSON.
func (x *Xml) UnmarshalJSON(data []byte) error {
	type Alias Xml
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(x),
	}
	return json.Unmarshal(data, &aux)
}
