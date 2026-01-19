package types

import "encoding/json"

// Contact represents contact information for the API.
type Contact struct {
	name  *string
	url   *string
	email *string
}

// Name sets the contact name.
func (c *Contact) Name(value string) *Contact {
	c.name = &value
	return c
}

// URL sets the contact URL.
func (c *Contact) URL(value string) *Contact {
	c.url = &value
	return c
}

// Email sets the contact email.
func (c *Contact) Email(value string) *Contact {
	c.email = &value
	return c
}

// MarshalJSON marshals the Contact to JSON.
func (c Contact) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name  *string `json:"name,omitempty"`
		URL   *string `json:"url,omitempty"`
		Email *string `json:"email,omitempty"`
	}{
		Name:  c.name,
		URL:   c.url,
		Email: c.email,
	})
}

// UnmarshalJSON unmarshals the Contact from JSON.
func (c *Contact) UnmarshalJSON(data []byte) error {
	type Alias Contact
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	return nil
}
