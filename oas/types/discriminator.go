package types

import (
	"encoding/json"
	"fmt"
)

// Discriminator represents a discriminator.
type Discriminator struct {
	propertyName string
	mapping      map[string]string
}

// PropertyName sets the property name.
func (d *Discriminator) PropertyName(value string) *Discriminator {
	d.propertyName = value
	return d
}

// Mapping adds a mapping.
func (d *Discriminator) Mapping(key, value string) *Discriminator {
	if d.mapping == nil {
		d.mapping = make(map[string]string)
	}
	d.mapping[key] = value
	return d
}

// Validate validates the Discriminator.
func (d Discriminator) Validate() error {
	if d.propertyName == "" {
		return fmt.Errorf("Discriminator.propertyName is required")
	}
	return nil
}

// MarshalJSON marshals the Discriminator to JSON.
func (d Discriminator) MarshalJSON() ([]byte, error) {
	if err := d.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		PropertyName string            `json:"propertyName"`
		Mapping      map[string]string `json:"mapping,omitempty"`
	}{
		PropertyName: d.propertyName,
		Mapping:      d.mapping,
	})
}

// UnmarshalJSON unmarshals the Discriminator from JSON.
func (d *Discriminator) UnmarshalJSON(data []byte) error {
	type Alias Discriminator
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(d),
	}
	return json.Unmarshal(data, &aux)
}
