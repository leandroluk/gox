package types

import (
	"encoding/json"
	"fmt"
)

type OAuthFlows struct {
	Implicit          *OAuthFlow `json:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty"`
}

// SecurityScheme represents a security scheme.
type SecurityScheme struct {
	type_            string
	description      *string
	name             *string
	in               *string
	scheme           *string
	bearerFormat     *string
	flows            *OAuthFlows
	openIdConnectUrl *string
}

func (s *SecurityScheme) Type(value string) *SecurityScheme {
	s.type_ = value
	return s
}

func (s *SecurityScheme) Description(value string) *SecurityScheme {
	s.description = &value
	return s
}

func (s *SecurityScheme) Name(value string) *SecurityScheme {
	s.name = &value
	return s
}

func (s *SecurityScheme) In(value string) *SecurityScheme {
	s.in = &value
	return s
}

func (s *SecurityScheme) Scheme(value string) *SecurityScheme {
	s.scheme = &value
	return s
}

func (s *SecurityScheme) BearerFormat(value string) *SecurityScheme {
	s.bearerFormat = &value
	return s
}

func (s *SecurityScheme) Flows(build func(f *OAuthFlows)) *SecurityScheme {
	if s.flows == nil {
		s.flows = &OAuthFlows{}
	}
	if build != nil {
		build(s.flows)
	}
	return s
}

func (s *SecurityScheme) OpenIdConnectUrl(value string) *SecurityScheme {
	s.openIdConnectUrl = &value
	return s
}

// Validate validates the SecurityScheme.
func (s SecurityScheme) Validate() error {
	if s.type_ == "" {
		return fmt.Errorf("SecurityScheme.type is required")
	}
	return nil
}

func (s SecurityScheme) MarshalJSON() ([]byte, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		Type             string      `json:"type"`
		Description      *string     `json:"description,omitempty"`
		Name             *string     `json:"name,omitempty"`
		In               *string     `json:"in,omitempty"`
		Scheme           *string     `json:"scheme,omitempty"`
		BearerFormat     *string     `json:"bearerFormat,omitempty"`
		Flows            *OAuthFlows `json:"flows,omitempty"`
		OpenIdConnectUrl *string     `json:"openIdConnectUrl,omitempty"`
	}{
		Type:             s.type_,
		Description:      s.description,
		Name:             s.name,
		In:               s.in,
		Scheme:           s.scheme,
		BearerFormat:     s.bearerFormat,
		Flows:            s.flows,
		OpenIdConnectUrl: s.openIdConnectUrl,
	})
}

// UnmarshalJSON unmarshals the SecurityScheme from JSON.
func (s *SecurityScheme) UnmarshalJSON(data []byte) error {
	type Alias SecurityScheme
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	return json.Unmarshal(data, &aux)
}
