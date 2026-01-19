package types

import "encoding/json"

// Link represents a link.
type Link struct {
	operationRef *string
	operationId  *string
	parameters   map[string]any
	requestBody  any
	description  *string
	server       *Server
}

func (l *Link) OperationRef(value string) *Link {
	l.operationRef = &value
	return l
}

func (l *Link) OperationID(value string) *Link {
	l.operationId = &value
	return l
}

func (l *Link) Parameter(name string, value any) *Link {
	if l.parameters == nil {
		l.parameters = make(map[string]any)
	}
	l.parameters[name] = value
	return l
}

func (l *Link) RequestBody(value any) *Link {
	l.requestBody = value
	return l
}

func (l *Link) Description(value string) *Link {
	l.description = &value
	return l
}

func (l *Link) Server(url string, optionalBuild ...func(s *Server)) *Link {
	if l.server == nil {
		l.server = &Server{}
	}
	l.server.URL(url)
	if len(optionalBuild) > 0 && optionalBuild[0] != nil {
		optionalBuild[0](l.server)
	}
	return l
}

func (l Link) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		OperationRef *string        `json:"operationRef,omitempty"`
		OperationId  *string        `json:"operationId,omitempty"`
		Parameters   map[string]any `json:"parameters,omitempty"`
		RequestBody  any            `json:"requestBody,omitempty"`
		Description  *string        `json:"description,omitempty"`
		Server       *Server        `json:"server,omitempty"`
	}{
		OperationRef: l.operationRef,
		OperationId:  l.operationId,
		Parameters:   l.parameters,
		RequestBody:  l.requestBody,
		Description:  l.description,
		Server:       l.server,
	})
}

// UnmarshalJSON unmarshals the Link from JSON.
func (l *Link) UnmarshalJSON(data []byte) error {
	type Alias Link
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(l),
	}
	return json.Unmarshal(data, &aux)
}
