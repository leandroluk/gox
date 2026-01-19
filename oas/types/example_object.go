package types

import "encoding/json"

// ExampleObject represents an example.
type ExampleObject struct {
	summary       *string
	description   *string
	value         any
	externalValue *string
}

// Summary sets the example summary.
func (e *ExampleObject) Summary(value string) *ExampleObject {
	e.summary = &value
	return e
}

// Description sets the example description.
func (e *ExampleObject) Description(value string) *ExampleObject {
	e.description = &value
	return e
}

// Value sets the example value.
func (e *ExampleObject) Value(value any) *ExampleObject {
	e.value = value
	return e
}

// ExternalValue sets the external value URL.
func (e *ExampleObject) ExternalValue(value string) *ExampleObject {
	e.externalValue = &value
	return e
}

// MarshalJSON marshals the ExampleObject to JSON.
func (e ExampleObject) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Summary       *string `json:"summary,omitempty"`
		Description   *string `json:"description,omitempty"`
		Value         any     `json:"value,omitempty"`
		ExternalValue *string `json:"externalValue,omitempty"`
	}{
		Summary:       e.summary,
		Description:   e.description,
		Value:         e.value,
		ExternalValue: e.externalValue,
	})
}

// UnmarshalJSON unmarshals the ExampleObject from JSON.
func (e *ExampleObject) UnmarshalJSON(data []byte) error {
	type Alias ExampleObject
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(e),
	}
	return json.Unmarshal(data, &aux)
}
