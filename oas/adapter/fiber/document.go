package adapter

import "github.com/leandroluk/gox/oas"

// DocumentBuilder is a builder for the root OpenAPI document
type DocumentBuilder struct {
	document *oas.Document
}

// OpenAPI sets the OpenAPI version for the document.
func (b *DocumentBuilder) OpenAPI(version string) *DocumentBuilder {
	b.document.OpenAPI(version)
	return b
}

// Info configures the OpenAPI Info object using a builder callback.
func (b *DocumentBuilder) Info(fn func(*oas.Info)) *DocumentBuilder {
	b.document.Info(fn)
	return b
}

// Servers sets multiple Server objects for the document.
func (b *DocumentBuilder) Servers(values []*oas.Server) *DocumentBuilder {
	b.document.Servers(values)
	return b
}

// Server adds a new Server object. The configuration callback is optional.
func (b *DocumentBuilder) Server(url string, optionalBuild ...func(s *oas.Server)) *DocumentBuilder {
	b.document.Server(url, optionalBuild...)
	return b
}

// Paths sets all Path objects at once.
func (b *DocumentBuilder) Paths(values map[string]*oas.Path) *DocumentBuilder {
	b.document.Paths(values)
	return b
}

// Path registers a new Path object configuration callback.
func (b *DocumentBuilder) Path(name string, build func(p *oas.Path)) *DocumentBuilder {
	b.document.Path(name, build)
	return b
}

// Webhooks sets all Webhook paths at once.
func (b *DocumentBuilder) Webhooks(values map[string]*oas.Path) *DocumentBuilder {
	b.document.Webhooks(values)
	return b
}

// Webhook registers a new Webhook configuration callback.
func (b *DocumentBuilder) Webhook(name string, build func(p *oas.Path)) *DocumentBuilder {
	b.document.Webhook(name, build)
	return b
}

// Components configures the OpenAPI Components object.
func (b *DocumentBuilder) Components(build func(c *oas.Components)) *DocumentBuilder {
	b.document.Components(build)
	return b
}

// Security adds a global Security Requirement object.
func (b *DocumentBuilder) Security(name string, scopes ...string) *DocumentBuilder {
	b.document.Security(name, scopes...)
	return b
}

// Tags sets multiple Tag objects at once.
func (b *DocumentBuilder) Tags(values []*oas.Tag) *DocumentBuilder {
	b.document.Tags(values)
	return b
}

// Tag adds a Tag object. The configuration callback is optional.
func (b *DocumentBuilder) Tag(name string, optionalBuild ...func(t *oas.Tag)) *DocumentBuilder {
	b.document.Tag(name, optionalBuild...)
	return b
}

// ExternalDocs configures the global External Documentation object.
func (b *DocumentBuilder) ExternalDocs(build func(e *oas.ExternalDocs)) *DocumentBuilder {
	b.document.ExternalDocs(build)
	return b
}

// ExternalDoc adds global external documentation. The configuration callback is optional.
func (b *DocumentBuilder) ExternalDoc(url string, optionalBuild ...func(e *oas.ExternalDocs)) *DocumentBuilder {
	b.document.ExternalDoc(url, optionalBuild...)
	return b
}

// WithBearerToken is a helper to easily configure a global Bearer Token security scheme.
func (b *DocumentBuilder) WithBearerToken(name string) *DocumentBuilder {
	b.document.WithBearerToken(name)
	return b
}

// WithApiKey is a helper to easily configure a global API Key security scheme.
func (b *DocumentBuilder) WithApiKey(name string, in string) *DocumentBuilder {
	b.document.WithApiKey(name, in)
	return b
}

// WithSecurityScheme configures any arbitrary global security scheme.
func (b *DocumentBuilder) WithSecurityScheme(name string, build func(s *oas.SecurityScheme)) *DocumentBuilder {
	b.document.WithSecurityScheme(name, build)
	return b
}

// Validate asserts the OpenAPI document correctness according to specifications.
func (b *DocumentBuilder) Validate() error {
	return b.document.Validate()
}

// MarshalJSON encodes the document into OpenAPI standard JSON.
func (b *DocumentBuilder) MarshalJSON() ([]byte, error) {
	return b.document.MarshalJSON()
}

// UnmarshalJSON parses OpenAPI standard JSON into the document tree.
func (b *DocumentBuilder) UnmarshalJSON(data []byte) error {
	return b.document.UnmarshalJSON(data)
}
