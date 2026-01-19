package types

import (
	"encoding/json"
	"fmt"
)

// SecurityRequirement represents a security requirement.
type SecurityRequirement map[string][]string

// SecurityRequirements represents a list of security requirements.
type SecurityRequirements []SecurityRequirement

// Document represents an OpenAPI document.
type Document struct {
	openapi           string
	info              *Info
	jsonSchemaDialect *string
	servers           []*Server
	paths             map[string]*Path
	webhooks          map[string]*Path
	components        *Components
	security          SecurityRequirements
	tags              []*Tag
	externalDocs      *ExternalDocs
}

func New() *Document {
	return &Document{}
}

func (d *Document) OpenAPI(value string) *Document {
	d.openapi = value
	return d
}

func (d *Document) Info(build func(i *Info)) *Document {
	if d.info == nil {
		d.info = &Info{}
	}
	if build != nil {
		build(d.info)
	}
	return d
}

func (d *Document) JsonSchemaDialect(value string) *Document {
	d.jsonSchemaDialect = &value
	return d
}

func (d *Document) Servers(values []*Server) *Document {
	d.servers = values
	return d
}

func (d *Document) Server(url string, optionalBuild ...func(s *Server)) *Document {
	if d.servers == nil {
		d.servers = make([]*Server, 0)
	}
	s := &Server{}
	s.URL(url)
	if len(optionalBuild) > 0 && optionalBuild[0] != nil {
		optionalBuild[0](s)
	}
	d.servers = append(d.servers, s)
	return d
}

func (d *Document) Paths(values map[string]*Path) *Document {
	d.paths = values
	return d
}

func (d *Document) Path(name string, build func(p *Path)) *Document {
	if d.paths == nil {
		d.paths = make(map[string]*Path)
	}
	p, ok := d.paths[name]
	if !ok {
		p = &Path{}
		d.paths[name] = p
	}
	if build != nil {
		build(p)
	}
	return d
}

func (d *Document) Webhooks(values map[string]*Path) *Document {
	d.webhooks = values
	return d
}

func (d *Document) Webhook(name string, build func(p *Path)) *Document {
	if d.webhooks == nil {
		d.webhooks = make(map[string]*Path)
	}
	p, ok := d.webhooks[name]
	if !ok {
		p = &Path{}
		d.webhooks[name] = p
	}
	if build != nil {
		build(p)
	}
	return d
}

func (d *Document) Components(build func(c *Components)) *Document {
	if d.components == nil {
		d.components = &Components{}
	}
	if build != nil {
		build(d.components)
	}
	return d
}

func (d *Document) Security(name string, scopes ...string) *Document {
	if d.security == nil {
		d.security = make(SecurityRequirements, 0)
	}
	req := SecurityRequirement{
		name: scopes,
	}
	d.security = append(d.security, req)
	return d
}

func (d *Document) Tags(values []*Tag) *Document {
	d.tags = values
	return d
}

func (d *Document) Tag(name string, optionalBuild ...func(t *Tag)) *Document {
	if d.tags == nil {
		d.tags = make([]*Tag, 0)
	}
	t := &Tag{}
	t.Name(name)
	if len(optionalBuild) > 0 && optionalBuild[0] != nil {
		optionalBuild[0](t)
	}
	d.tags = append(d.tags, t)
	return d
}

func (d *Document) ExternalDocs(build func(e *ExternalDocs)) *Document {
	if d.externalDocs == nil {
		d.externalDocs = &ExternalDocs{}
	}
	if build != nil {
		build(d.externalDocs)
	}
	return d
}

// ExternalDoc sets external documentation with URL as required parameter.
func (d *Document) ExternalDoc(url string, optionalBuild ...func(e *ExternalDocs)) *Document {
	if d.externalDocs == nil {
		d.externalDocs = &ExternalDocs{}
	}
	d.externalDocs.URL(url)
	if len(optionalBuild) > 0 && optionalBuild[0] != nil {
		optionalBuild[0](d.externalDocs)
	}
	return d
}

// WithBearerToken registers a Bearer token security scheme.
func (d *Document) WithBearerToken(name string) *Document {
	return d.Components(func(c *Components) {
		c.SecurityScheme(name, func(s *SecurityScheme) {
			s.Type("http").Scheme("bearer").BearerFormat("JWT")
		})
	})
}

// WithApiKey registers an API key security scheme.
func (d *Document) WithApiKey(name string, in string) *Document {
	return d.Components(func(c *Components) {
		c.SecurityScheme(name, func(s *SecurityScheme) {
			s.Type("apiKey").Name(name).In(in)
		})
	})
}

// WithSecurityScheme registers a custom security scheme.
func (d *Document) WithSecurityScheme(name string, build func(s *SecurityScheme)) *Document {
	return d.Components(func(c *Components) {
		c.SecurityScheme(name, build)
	})
}

// Validate validates the Document.
func (d Document) Validate() error {
	if d.info == nil {
		return fmt.Errorf("Document.info is required")
	}
	return d.info.Validate()
}

// MarshalJSON marshals the Document to JSON.
func (d Document) MarshalJSON() ([]byte, error) {
	openapi := d.openapi
	if openapi == "" {
		openapi = "3.1.0"
	}
	if err := d.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		OpenAPI           string               `json:"openapi"`
		Info              *Info                `json:"info"`
		JsonSchemaDialect *string              `json:"jsonSchemaDialect,omitempty"`
		Servers           []*Server            `json:"servers,omitempty"`
		Paths             map[string]*Path     `json:"paths,omitempty"`
		Webhooks          map[string]*Path     `json:"webhooks,omitempty"`
		Components        *Components          `json:"components,omitempty"`
		Security          SecurityRequirements `json:"security,omitempty"`
		Tags              []*Tag               `json:"tags,omitempty"`
		ExternalDocs      *ExternalDocs        `json:"externalDocs,omitempty"`
	}{
		OpenAPI:           openapi,
		Info:              d.info,
		JsonSchemaDialect: d.jsonSchemaDialect,
		Servers:           d.servers,
		Paths:             d.paths,
		Webhooks:          d.webhooks,
		Components:        d.components,
		Security:          d.security,
		Tags:              d.tags,
		ExternalDocs:      d.externalDocs,
	})
}

// UnmarshalJSON unmarshals the Document from JSON.
func (d *Document) UnmarshalJSON(data []byte) error {
	type Alias Document
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(d),
	}
	return json.Unmarshal(data, &aux)
}
