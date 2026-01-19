package types

import "encoding/json"

// Server represents a server.
type Server struct {
	url         string
	description *string
	variables   map[string]*ServerVariable
}

// URL sets the server URL.
func (s *Server) URL(value string) *Server {
	s.url = value
	return s
}

// Description sets the server description.
func (s *Server) Description(value string) *Server {
	s.description = &value
	return s
}

// Variable configures a server variable.
func (s *Server) Variable(name string, build func(v *ServerVariable)) *Server {
	if s.variables == nil {
		s.variables = make(map[string]*ServerVariable)
	}
	v := &ServerVariable{}
	if build != nil {
		build(v)
	}
	s.variables[name] = v
	return s
}

// MarshalJSON marshals the Server to JSON.
func (s Server) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		URL         string                     `json:"url"`
		Description *string                    `json:"description,omitempty"`
		Variables   map[string]*ServerVariable `json:"variables,omitempty"`
	}{
		URL:         s.url,
		Description: s.description,
		Variables:   s.variables,
	})
}

// UnmarshalJSON unmarshals the Server from JSON.
func (s *Server) UnmarshalJSON(data []byte) error {
	type Alias Server
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	return json.Unmarshal(data, &aux)
}
