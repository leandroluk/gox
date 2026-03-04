package adapter

import (
	"github.com/gofiber/fiber/v2"
	"github.com/leandroluk/gox/oas"
)

// Fiber represents the Fiber wrapper with OpenAPI support
type Fiber struct {
	*fiber.App
	document *oas.Document
}

// NewFiber creates a new App with OAS features
func NewFiber(app *fiber.App) *Fiber {
	return &Fiber{
		App:      app,
		document: oas.New(),
	}
}

// OAS configures the root OpenAPI document
func (a *Fiber) OAS(fn func(*DocumentBuilder)) *Fiber {
	builder := &DocumentBuilder{document: a.document}
	fn(builder)
	return a
}

// Document returns the OpenAPI document to be serialized.
//
//	app.Get("/swagger.json", func(c *fiber.Ctx) error {
//	    return c.JSON(app.Document())
//	})
func (a *Fiber) Document() *oas.Document {
	return a.document
}

// Group creates a new group with OpenAPI context
func (a *Fiber) Group(prefix string, handlers ...fiber.Handler) *Group {
	return &Group{
		Router:     a.App.Group(prefix, handlers...),
		document:   a.document,
		pathPrefix: prefix,
	}
}
