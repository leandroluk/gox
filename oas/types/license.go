package types

import (
	"encoding/json"
	"fmt"
)

// License represents license information for the API.
type License struct {
	name string
	url  *string
}

// Name sets the license name.
func (l *License) Name(value string) *License {
	l.name = value
	return l
}

// URL sets the license URL.
func (l *License) URL(value string) *License {
	l.url = &value
	return l
}

// Validate validates the License.
func (l License) Validate() error {
	if l.name == "" {
		return fmt.Errorf("License.name is required")
	}
	return nil
}

// MarshalJSON marshals the License to JSON.
func (l License) MarshalJSON() ([]byte, error) {
	if err := l.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		Name string  `json:"name"`
		URL  *string `json:"url,omitempty"`
	}{
		Name: l.name,
		URL:  l.url,
	})
}

// UnmarshalJSON unmarshals the License from JSON.
func (l *License) UnmarshalJSON(data []byte) error {
	type Alias License
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(l),
	}
	return json.Unmarshal(data, &aux)
}
