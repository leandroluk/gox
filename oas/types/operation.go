package types

import "encoding/json"

// Operation represents an API operation.
type Operation struct {
	tags         []string
	summary      *string
	description  *string
	operationId  *string
	parameters   []*Parameter
	requestBody  *RequestBody
	responses    map[string]*Response
	deprecated   *bool
	externalDocs *ExternalDocs
	security     []map[string][]string
	servers      []*Server
}

func (o *Operation) Tags(values ...string) *Operation {
	o.tags = values
	return o
}

func (o *Operation) Summary(value string) *Operation {
	o.summary = &value
	return o
}

func (o *Operation) Description(value string) *Operation {
	o.description = &value
	return o
}

func (o *Operation) OperationId(value string) *Operation {
	o.operationId = &value
	return o
}

func (o *Operation) Parameter(build func(p *Parameter)) *Operation {
	p := &Parameter{}
	if build != nil {
		build(p)
	}
	o.parameters = append(o.parameters, p)
	return o
}

func (o *Operation) RequestBody(build func(r *RequestBody)) *Operation {
	if o.requestBody == nil {
		o.requestBody = &RequestBody{}
	}
	if build != nil {
		build(o.requestBody)
	}
	return o
}

func (o *Operation) Responses(builders ...*ResponseWithCode) *Operation {
	if o.responses == nil {
		o.responses = make(map[string]*Response)
	}
	for _, b := range builders {
		o.responses[b.Code] = b.Response
	}
	return o
}

func (o *Operation) Response(name string, build func(r *Response)) *Operation {
	if o.responses == nil {
		o.responses = make(map[string]*Response)
	}
	r, ok := o.responses[name]
	if !ok {
		r = &Response{}
		o.responses[name] = r
	}
	if build != nil {
		build(r)
	}
	return o
}

func (o *Operation) Deprecated(value bool) *Operation {
	o.deprecated = &value
	return o
}

func (o *Operation) ExternalDocs(build func(e *ExternalDocs)) *Operation {
	if o.externalDocs == nil {
		o.externalDocs = &ExternalDocs{}
	}
	if build != nil {
		build(o.externalDocs)
	}
	return o
}

// ExternalDoc sets external documentation with URL as required parameter.
func (o *Operation) ExternalDoc(url string, optionalBuild ...func(e *ExternalDocs)) *Operation {
	if o.externalDocs == nil {
		o.externalDocs = &ExternalDocs{}
	}
	o.externalDocs.URL(url)
	if len(optionalBuild) > 0 && optionalBuild[0] != nil {
		optionalBuild[0](o.externalDocs)
	}
	return o
}

func (o *Operation) Security(name string, scopes ...string) *Operation {
	if o.security == nil {
		o.security = make([]map[string][]string, 0)
	}
	// Security requirement is a map, usually one entry per requirement object in the list
	req := map[string][]string{
		name: scopes,
	}
	o.security = append(o.security, req)
	return o
}

func (o *Operation) Server(url string, optionalBuild ...func(s *Server)) *Operation {
	if o.servers == nil {
		o.servers = make([]*Server, 0)
	}
	// Servers are a list, so we always append new or maybe we should have a reliable way to find?
	// Spec says generic list. For builder pattern, usually append.
	s := &Server{}
	s.URL(url)
	if len(optionalBuild) > 0 && optionalBuild[0] != nil {
		optionalBuild[0](s)
	}
	o.servers = append(o.servers, s)
	return o
}

// UseBearerToken applies a Bearer token security requirement.
func (o *Operation) UseBearerToken(name string, scopes ...string) *Operation {
	return o.Security(name, scopes...)
}

// UseApiKey applies an API key security requirement.
func (o *Operation) UseApiKey(name string) *Operation {
	return o.Security(name)
}

// UseSecurityScheme applies a custom security scheme requirement.
func (o *Operation) UseSecurityScheme(name string, scopes ...string) *Operation {
	return o.Security(name, scopes...)
}

func (o Operation) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Tags         []string              `json:"tags,omitempty"`
		Summary      *string               `json:"summary,omitempty"`
		Description  *string               `json:"description,omitempty"`
		OperationId  *string               `json:"operationId,omitempty"`
		Parameters   []*Parameter          `json:"parameters,omitempty"`
		RequestBody  *RequestBody          `json:"requestBody,omitempty"`
		Responses    map[string]*Response  `json:"responses,omitempty"`
		Deprecated   *bool                 `json:"deprecated,omitempty"`
		ExternalDocs *ExternalDocs         `json:"externalDocs,omitempty"`
		Security     []map[string][]string `json:"security,omitempty"`
		Servers      []*Server             `json:"servers,omitempty"`
	}{
		Tags:         o.tags,
		Summary:      o.summary,
		Description:  o.description,
		OperationId:  o.operationId,
		Parameters:   o.parameters,
		RequestBody:  o.requestBody,
		Responses:    o.responses,
		Deprecated:   o.deprecated,
		ExternalDocs: o.externalDocs,
		Security:     o.security,
		Servers:      o.servers,
	})
}

// UnmarshalJSON unmarshals the Operation from JSON.
func (o *Operation) UnmarshalJSON(data []byte) error {
	type Alias Operation
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(o),
	}
	return json.Unmarshal(data, &aux)
}
