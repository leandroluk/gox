package types

import (
	"encoding/json"
	"fmt"
)

// ServerVariable represents a server variable for server URL template substitution.
type ServerVariable struct {
	enum        []string
	default_    string
	description *string
}

// Enum adds literal values to the enum.
func (s *ServerVariable) Enum(values ...string) *ServerVariable {
	s.enum = values
	return s
}

// Default sets the default value.
func (s *ServerVariable) Default(value string) *ServerVariable {
	s.default_ = value
	return s
}

// Description sets the description.
func (s *ServerVariable) Description(value string) *ServerVariable {
	s.description = &value
	return s
}

// Validate validates the ServerVariable.
func (s ServerVariable) Validate() error {
	if s.default_ == "" {
		return fmt.Errorf("ServerVariable.default is required")
	}
	return nil
}

// MarshalJSON marshals the ServerVariable to JSON.
func (s ServerVariable) MarshalJSON() ([]byte, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		Enum        []string `json:"enum,omitempty"`
		Default     string   `json:"default"`
		Description *string  `json:"description,omitempty"`
	}{
		Enum:        s.enum,
		Default:     s.default_,
		Description: s.description,
	})
}

// UnmarshalJSON unmarshals the ServerVariable from JSON.
func (s *ServerVariable) UnmarshalJSON(data []byte) error {
	type Alias ServerVariable
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	return json.Unmarshal(data, &aux)
}
