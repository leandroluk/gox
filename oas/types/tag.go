package types

import (
	"encoding/json"
	"fmt"
)

// Tag represents a tag.
type Tag struct {
	name         string
	description  *string
	externalDocs *ExternalDocs
}

// Name sets the tag name.
func (t *Tag) Name(value string) *Tag {
	t.name = value
	return t
}

// Description sets the tag description.
func (t *Tag) Description(value string) *Tag {
	t.description = &value
	return t
}

// ExternalDocs configures external documentation.
func (t *Tag) ExternalDocs(build func(e *ExternalDocs)) *Tag {
	if t.externalDocs == nil {
		t.externalDocs = &ExternalDocs{}
	}
	if build != nil {
		build(t.externalDocs)
	}
	return t
}

// ExternalDoc sets external documentation with URL as required parameter.
func (t *Tag) ExternalDoc(url string, optionalBuild ...func(e *ExternalDocs)) *Tag {
	if t.externalDocs == nil {
		t.externalDocs = &ExternalDocs{}
	}
	t.externalDocs.URL(url)
	if len(optionalBuild) > 0 && optionalBuild[0] != nil {
		optionalBuild[0](t.externalDocs)
	}
	return t
}

// Validate validates the Tag.
func (t Tag) Validate() error {
	if t.name == "" {
		return fmt.Errorf("Tag.name is required")
	}
	if t.externalDocs != nil {
		if err := t.externalDocs.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// MarshalJSON marshals the Tag to JSON.
func (t Tag) MarshalJSON() ([]byte, error) {
	if err := t.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		Name         string        `json:"name"`
		Description  *string       `json:"description,omitempty"`
		ExternalDocs *ExternalDocs `json:"externalDocs,omitempty"`
	}{
		Name:         t.name,
		Description:  t.description,
		ExternalDocs: t.externalDocs,
	})
}

// UnmarshalJSON unmarshals the Tag from JSON.
func (t *Tag) UnmarshalJSON(data []byte) error {
	type Alias Tag
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(t),
	}
	return json.Unmarshal(data, &aux)
}
