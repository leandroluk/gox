package adapter

import (
	"regexp"

	"github.com/gofiber/fiber/v2"
	"github.com/leandroluk/gox/oas"
)

// RouteBuilder joins OAS operation configuration and Fiber handlers.
type RouteBuilder struct {
	*oas.Operation
	handlers []fiber.Handler
}

// Handlers registers the Fiber handlers for this route.
func (r *RouteBuilder) Handlers(handlers ...fiber.Handler) *RouteBuilder {
	r.handlers = handlers
	return r
}

// FiberPathToOAS converts Fiber paths to OpenAPI format.
// Examples:
//
//	/users/:id        → /users/{id}
//	/users/:id?       → /users/{id}   (optional Fiber parameters)
//	/users/:id/:field → /users/{id}/{field}
func (g *Group) fiberPathToOAS(path string) string {
	re := regexp.MustCompile(`:([a-zA-Z0-9_]+)\??`)
	return re.ReplaceAllString(path, `{$1}`)
}

// addRoute registers the route in Fiber and documents the operation in OAS.
// It enforces the use of a RouteBuilder callback.
func (g *Group) addRoute(method, path string, fn func(*RouteBuilder)) fiber.Router {

	fullOASPath := g.fiberPathToOAS(g.pathPrefix + path)

	operation := &oas.Operation{}
	if len(g.tags) > 0 {
		operation.Tags(g.tags...)
	}

	for _, paramName := range g.ExtractPathParams(path) {
		name := paramName // capture for closure
		operation.Parameter(func(p *oas.Parameter) {
			p.In("path").
				Name(name).
				Description(name + " parameter").
				Required(true).
				Schema(func(s *oas.Schema) {
					s.String() // default type; override via fn if necessary
				})
		})
	}

	var fiberHandlers []fiber.Handler
	if fn != nil {
		builder := &RouteBuilder{
			Operation: operation,
			handlers:  make([]fiber.Handler, 0),
		}
		fn(builder)
		fiberHandlers = builder.handlers
	}

	if len(fiberHandlers) > 0 {
		g.Router.Add(method, path, fiberHandlers...)
	}

	// Registers the operation in the OAS document
	g.document.Path(fullOASPath, func(p *oas.Path) {
		opFn := func(o *oas.Operation) { *o = *operation }
		switch method {
		case "GET":
			p.Get(opFn)
		case "POST":
			p.Post(opFn)
		case "PUT":
			p.Put(opFn)
		case "DELETE":
			p.Delete(opFn)
		case "PATCH":
			p.Patch(opFn)
		case "OPTIONS":
			p.Options(opFn)
		case "HEAD":
			p.Head(opFn)
		case "TRACE":
			p.Trace(opFn)
		}
	})

	return g.Router
}

// Get adds a GET route requiring OAS documentation builder.
func (g *Group) Get(path string, fn func(*RouteBuilder)) fiber.Router {
	return g.addRoute("GET", path, fn)
}

// Post adds a POST route requiring OAS documentation builder.
func (g *Group) Post(path string, fn func(*RouteBuilder)) fiber.Router {
	return g.addRoute("POST", path, fn)
}

// Put adds a PUT route requiring OAS documentation builder.
func (g *Group) Put(path string, fn func(*RouteBuilder)) fiber.Router {
	return g.addRoute("PUT", path, fn)
}

// Delete adds a DELETE route requiring OAS documentation builder.
func (g *Group) Delete(path string, fn func(*RouteBuilder)) fiber.Router {
	return g.addRoute("DELETE", path, fn)
}

// Options adds an OPTIONS route requiring OAS documentation builder.
func (g *Group) Options(path string, fn func(*RouteBuilder)) fiber.Router {
	return g.addRoute("OPTIONS", path, fn)
}

// Head adds a HEAD route requiring OAS documentation builder.
func (g *Group) Head(path string, fn func(*RouteBuilder)) fiber.Router {
	return g.addRoute("HEAD", path, fn)
}

// Patch adds a PATCH route requiring OAS documentation builder.
func (g *Group) Patch(path string, fn func(*RouteBuilder)) fiber.Router {
	return g.addRoute("PATCH", path, fn)
}

// Trace adds a TRACE route requiring OAS documentation builder.
func (g *Group) Trace(path string, fn func(*RouteBuilder)) fiber.Router {
	return g.addRoute("TRACE", path, fn)
}
