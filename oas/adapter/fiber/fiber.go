// github.com/leandroluk/gox/oas/adapter/fiber/fiber.go
package adapter

import (
	"regexp"

	"github.com/gofiber/fiber/v2"
	"github.com/leandroluk/gox/oas"
)

// App represents the Fiber wrapper with OpenAPI support
type App struct {
	*fiber.App
	document *oas.Document
}

// Wrap creates a new App with OAS features
func Wrap(app *fiber.App) *App {
	return &App{
		App:      app,
		document: oas.New(),
	}
}

// OAS configures the root OpenAPI document
func (a *App) OAS(fn func(*DocumentBuilder)) *App {
	builder := &DocumentBuilder{document: a.document}
	fn(builder)
	return a
}

// Document returns the OpenAPI document to be serialized.
//
//	app.Get("/swagger.json", func(c *fiber.Ctx) error {
//	    return c.JSON(app.Document())
//	})
func (a *App) Document() *oas.Document {
	return a.document
}

// Group creates a new group with OpenAPI context
func (a *App) Group(prefix string, handlers ...fiber.Handler) *Group {
	return &Group{
		Router:     a.App.Group(prefix, handlers...),
		document:   a.document,
		pathPrefix: prefix,
	}
}

// FiberPathToOAS converts Fiber paths to OpenAPI format.
// Examples:
//
//	/users/:id        → /users/{id}
//	/users/:id?       → /users/{id}   (optional Fiber parameters)
//	/users/:id/:field → /users/{id}/{field}
func FiberPathToOAS(path string) string {
	re := regexp.MustCompile(`:([a-zA-Z0-9_]+)\??`)
	return re.ReplaceAllString(path, `{$1}`)
}

// DocumentBuilder is a builder for the root OpenAPI document
type DocumentBuilder struct {
	document *oas.Document
}

func (b *DocumentBuilder) OpenAPI(version string) *DocumentBuilder {
	b.document.OpenAPI(version)
	return b
}

func (b *DocumentBuilder) Info(fn func(*oas.Info)) *DocumentBuilder {
	b.document.Info(fn)
	return b
}

func (b *DocumentBuilder) Servers(values []*oas.Server) *DocumentBuilder {
	b.document.Servers(values)
	return b
}

func (b *DocumentBuilder) Server(url string, optionalBuild ...func(s *oas.Server)) *DocumentBuilder {
	b.document.Server(url, optionalBuild...)
	return b
}

func (b *DocumentBuilder) Paths(values map[string]*oas.Path) *DocumentBuilder {
	b.document.Paths(values)
	return b
}

func (b *DocumentBuilder) Path(name string, build func(p *oas.Path)) *DocumentBuilder {
	b.document.Path(name, build)
	return b
}

func (b *DocumentBuilder) Webhooks(values map[string]*oas.Path) *DocumentBuilder {
	b.document.Webhooks(values)
	return b
}

func (b *DocumentBuilder) Webhook(name string, build func(p *oas.Path)) *DocumentBuilder {
	b.document.Webhook(name, build)
	return b
}

func (b *DocumentBuilder) Components(build func(c *oas.Components)) *DocumentBuilder {
	b.document.Components(build)
	return b
}

func (b *DocumentBuilder) Security(name string, scopes ...string) *DocumentBuilder {
	b.document.Security(name, scopes...)
	return b
}

func (b *DocumentBuilder) Tags(values []*oas.Tag) *DocumentBuilder {
	b.document.Tags(values)
	return b
}

func (b *DocumentBuilder) Tag(name string, optionalBuild ...func(t *oas.Tag)) *DocumentBuilder {
	b.document.Tag(name, optionalBuild...)
	return b
}

func (b *DocumentBuilder) ExternalDocs(build func(e *oas.ExternalDocs)) *DocumentBuilder {
	b.document.ExternalDocs(build)
	return b
}

func (b *DocumentBuilder) ExternalDoc(url string, optionalBuild ...func(e *oas.ExternalDocs)) *DocumentBuilder {
	b.document.ExternalDoc(url, optionalBuild...)
	return b
}

func (b *DocumentBuilder) WithBearerToken(name string) *DocumentBuilder {
	b.document.WithBearerToken(name)
	return b
}

func (b *DocumentBuilder) WithApiKey(name string, in string) *DocumentBuilder {
	b.document.WithApiKey(name, in)
	return b
}

func (b *DocumentBuilder) WithSecurityScheme(name string, build func(s *oas.SecurityScheme)) *DocumentBuilder {
	b.document.WithSecurityScheme(name, build)
	return b
}

func (b *DocumentBuilder) Validate() error {
	return b.document.Validate()
}

func (b *DocumentBuilder) MarshalJSON() ([]byte, error) {
	return b.document.MarshalJSON()
}

func (b *DocumentBuilder) UnmarshalJSON(data []byte) error {
	return b.document.UnmarshalJSON(data)
}
