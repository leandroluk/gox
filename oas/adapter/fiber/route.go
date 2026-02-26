// github.com/leandroluk/gox/oas/wrap/fiber/route.go
package wrap

import (
	"github.com/gofiber/fiber/v2"
	"github.com/leandroluk/gox/oas"
)

// addRoute registers the route in Fiber and documents the operation in OAS.
// The fn parameter is optional: if nil, the route is documented only with
// the group tags and automatically extracted path parameters.
func (g *Group) addRoute(method, path string, handler fiber.Handler, fn func(*oas.Operation)) fiber.Router {
	// Registers the route in Fiber (the Fiber Router accumulates prefixes)
	g.Router.Add(method, path, handler)

	// Builds the full OAS path: group prefix + route path,
	// converting Fiber format (:param) to OpenAPI ({param})
	fullOASPath := fiberPathToOAS(g.pathPrefix + path)

	// Initializes the operation and applies group tags
	operation := oas.Operation{}
	if len(g.tags) > 0 {
		operation.Tags(g.tags...)
	}

	// Extracts path parameters automatically and adds them to the operation.
	// The user can override or complement this via fn.
	for _, paramName := range g.extractPathParams(path) {
		name := paramName // capture for closure (avoids loop bug in Go < 1.22)
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

	// Allows additional operation configuration by the caller
	if fn != nil {
		fn(&operation)
	}

	// Registers the operation in the OAS document
	capturedOp := operation // avoids loop variable pointer capture
	g.document.Path(fullOASPath, func(p *oas.Path) {
		opFn := func(o *oas.Operation) { *o = capturedOp }
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
func (g *Group) Get(path string, handler fiber.Handler, fn ...func(*oas.Operation)) fiber.Router {
	var opFn func(*oas.Operation)
	if len(fn) > 0 {
		opFn = fn[0]
	}
	return g.addRoute("GET", path, handler, opFn)
}

// Post adds a POST route with optional OAS documentation.
func (g *Group) Post(path string, handler fiber.Handler, fn ...func(*oas.Operation)) fiber.Router {
	var opFn func(*oas.Operation)
	if len(fn) > 0 {
		opFn = fn[0]
	}
	return g.addRoute("POST", path, handler, opFn)
}

// Put adds a PUT route with optional OAS documentation.
func (g *Group) Put(path string, handler fiber.Handler, fn ...func(*oas.Operation)) fiber.Router {
	var opFn func(*oas.Operation)
	if len(fn) > 0 {
		opFn = fn[0]
	}
	return g.addRoute("PUT", path, handler, opFn)
}

// Delete adds a DELETE route with optional OAS documentation.
func (g *Group) Delete(path string, handler fiber.Handler, fn ...func(*oas.Operation)) fiber.Router {
	var opFn func(*oas.Operation)
	if len(fn) > 0 {
		opFn = fn[0]
	}
	return g.addRoute("DELETE", path, handler, opFn)
}

// Options adds an OPTIONS route with optional OAS documentation.
func (g *Group) Options(path string, handler fiber.Handler, fn ...func(*oas.Operation)) fiber.Router {
	var opFn func(*oas.Operation)
	if len(fn) > 0 {
		opFn = fn[0]
	}
	return g.addRoute("OPTIONS", path, handler, opFn)
}

// Head adds a HEAD route with optional OAS documentation.
func (g *Group) Head(path string, handler fiber.Handler, fn ...func(*oas.Operation)) fiber.Router {
	var opFn func(*oas.Operation)
	if len(fn) > 0 {
		opFn = fn[0]
	}
	return g.addRoute("HEAD", path, handler, opFn)
}

// Patch adds a PATCH route with optional OAS documentation.
func (g *Group) Patch(path string, handler fiber.Handler, fn ...func(*oas.Operation)) fiber.Router {
	var opFn func(*oas.Operation)
	if len(fn) > 0 {
		opFn = fn[0]
	}
	return g.addRoute("PATCH", path, handler, opFn)
}

// Trace adds a TRACE route with optional OAS documentation.
func (g *Group) Trace(path string, handler fiber.Handler, fn ...func(*oas.Operation)) fiber.Router {
	var opFn func(*oas.Operation)
	if len(fn) > 0 {
		opFn = fn[0]
	}
	return g.addRoute("TRACE", path, handler, opFn)
}
