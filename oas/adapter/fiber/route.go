// github.com/leandroluk/gox/oas/adapter/fiber/route.go
package adapter

import (
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

// addRoute registers the route in Fiber and documents the operation in OAS.
// It accepts either standard fiber.Handler functions or a single func(*RouteBuilder).
func (g *Group) addRoute(method, path string, args ...any) fiber.Router {
	var fiberHandlers []fiber.Handler
	var routeBuilderFn func(*RouteBuilder)

	for _, arg := range args {
		switch v := arg.(type) {
		case func(*RouteBuilder):
			routeBuilderFn = v
		case fiber.Handler:
			fiberHandlers = append(fiberHandlers, v)
		}
	}

	fullOASPath := FiberPathToOAS(g.pathPrefix + path)

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

	if routeBuilderFn != nil {
		builder := &RouteBuilder{
			Operation: operation,
			handlers:  fiberHandlers,
		}
		routeBuilderFn(builder)
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

// Get adds a GET route with optional OAS documentation.
func (g *Group) Get(path string, handlers ...any) fiber.Router {
	return g.addRoute("GET", path, handlers...)
}

// Post adds a POST route with optional OAS documentation.
func (g *Group) Post(path string, handlers ...any) fiber.Router {
	return g.addRoute("POST", path, handlers...)
}

// Put adds a PUT route with optional OAS documentation.
func (g *Group) Put(path string, handlers ...any) fiber.Router {
	return g.addRoute("PUT", path, handlers...)
}

// Delete adds a DELETE route with optional OAS documentation.
func (g *Group) Delete(path string, handlers ...any) fiber.Router {
	return g.addRoute("DELETE", path, handlers...)
}

// Options adds an OPTIONS route with optional OAS documentation.
func (g *Group) Options(path string, handlers ...any) fiber.Router {
	return g.addRoute("OPTIONS", path, handlers...)
}

// Head adds a HEAD route with optional OAS documentation.
func (g *Group) Head(path string, handlers ...any) fiber.Router {
	return g.addRoute("HEAD", path, handlers...)
}

// Patch adds a PATCH route with optional OAS documentation.
func (g *Group) Patch(path string, handlers ...any) fiber.Router {
	return g.addRoute("PATCH", path, handlers...)
}

// Trace adds a TRACE route with optional OAS documentation.
func (g *Group) Trace(path string, handlers ...any) fiber.Router {
	return g.addRoute("TRACE", path, handlers...)
}
