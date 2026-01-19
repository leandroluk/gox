package types

import (
	"encoding/json"
	"fmt"
)

// Info represents the metadata for the API.
type Info struct {
	title          string
	summary        *string
	description    *string
	termsOfService *string
	contact        *Contact
	license        *License
	version        string
}

// Title sets the API title.
func (i *Info) Title(value string) *Info {
	i.title = value
	return i
}

// Summary sets the API summary.
func (i *Info) Summary(value string) *Info {
	i.summary = &value
	return i
}

// Description sets the API description.
func (i *Info) Description(value string) *Info {
	i.description = &value
	return i
}

// TermsOfService sets the terms of service URL.
func (i *Info) TermsOfService(value string) *Info {
	i.termsOfService = &value
	return i
}

// Contact returns the Contact object for configuration.
func (i *Info) Contact() *Contact {
	if i.contact == nil {
		i.contact = &Contact{}
	}
	return i.contact
}

// License returns the License object for configuration.
func (i *Info) License() *License {
	if i.license == nil {
		i.license = &License{}
	}
	return i.license
}

// Version sets the API version.
func (i *Info) Version(value string) *Info {
	i.version = value
	return i
}

// Validate validates the Info.
func (i Info) Validate() error {
	if i.title == "" {
		return fmt.Errorf("Info.title is required")
	}
	if i.version == "" {
		return fmt.Errorf("Info.version is required")
	}
	if i.license != nil {
		if err := i.license.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// MarshalJSON marshals the Info to JSON.
func (i Info) MarshalJSON() ([]byte, error) {
	if err := i.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		Title          string   `json:"title"`
		Summary        *string  `json:"summary,omitempty"`
		Description    *string  `json:"description,omitempty"`
		TermsOfService *string  `json:"termsOfService,omitempty"`
		Contact        *Contact `json:"contact,omitempty"`
		License        *License `json:"license,omitempty"`
		Version        string   `json:"version"`
	}{
		Title:          i.title,
		Summary:        i.summary,
		Description:    i.description,
		TermsOfService: i.termsOfService,
		Contact:        i.contact,
		License:        i.license,
		Version:        i.version,
	})
}

// UnmarshalJSON unmarshals the Info from JSON.
func (i *Info) UnmarshalJSON(data []byte) error {
	type Alias Info
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(i),
	}
	return json.Unmarshal(data, &aux)
}
