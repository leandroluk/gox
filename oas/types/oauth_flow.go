package types

import (
	"encoding/json"
	"fmt"
)

// OAuthFlow represents an OAuth flow.
type OAuthFlow struct {
	authorizationUrl *string
	tokenUrl         *string
	refreshUrl       *string
	scopes           map[string]string
}

func (o *OAuthFlow) AuthorizationUrl(value string) *OAuthFlow {
	o.authorizationUrl = &value
	return o
}

func (o *OAuthFlow) TokenUrl(value string) *OAuthFlow {
	o.tokenUrl = &value
	return o
}

func (o *OAuthFlow) RefreshUrl(value string) *OAuthFlow {
	o.refreshUrl = &value
	return o
}

func (o *OAuthFlow) Scope(key, value string) *OAuthFlow {
	if o.scopes == nil {
		o.scopes = make(map[string]string)
	}
	o.scopes[key] = value
	return o
}

// Validate validates the OAuthFlow.
func (o OAuthFlow) Validate() error {
	if o.scopes == nil {
		return fmt.Errorf("OAuthFlow.scopes is required")
	}
	return nil
}

func (o OAuthFlow) MarshalJSON() ([]byte, error) {
	if err := o.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		AuthorizationUrl *string           `json:"authorizationUrl,omitempty"`
		TokenUrl         *string           `json:"tokenUrl,omitempty"`
		RefreshUrl       *string           `json:"refreshUrl,omitempty"`
		Scopes           map[string]string `json:"scopes"`
	}{
		AuthorizationUrl: o.authorizationUrl,
		TokenUrl:         o.tokenUrl,
		RefreshUrl:       o.refreshUrl,
		Scopes:           o.scopes,
	})
}

// UnmarshalJSON unmarshals the OAuthFlow from JSON.
func (o *OAuthFlow) UnmarshalJSON(data []byte) error {
	type Alias OAuthFlow
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(o),
	}
	return json.Unmarshal(data, &aux)
}
