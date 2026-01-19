package types

import (
	"encoding/json"
	"fmt"
)

// ExternalDocs represents external documentation.
type ExternalDocs struct {
	description *string
	url         string
}

// Description sets the description.
func (e *ExternalDocs) Description(value string) *ExternalDocs {
	e.description = &value
	return e
}

// URL sets the URL.
func (e *ExternalDocs) URL(value string) *ExternalDocs {
	e.url = value
	return e
}

// Validate validates the ExternalDocs.
func (e ExternalDocs) Validate() error {
	if e.url == "" {
		return fmt.Errorf("ExternalDocs.url is required")
	}
	return nil
}

// MarshalJSON marshals the ExternalDocs to JSON.
func (e ExternalDocs) MarshalJSON() ([]byte, error) {
	if err := e.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		Description *string `json:"description,omitempty"`
		URL         string  `json:"url"`
	}{
		Description: e.description,
		URL:         e.url,
	})
}

// UnmarshalJSON unmarshals the ExternalDocs from JSON.
func (e *ExternalDocs) UnmarshalJSON(data []byte) error {
	type Alias ExternalDocs
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(e),
	}
	return json.Unmarshal(data, &aux)
}
