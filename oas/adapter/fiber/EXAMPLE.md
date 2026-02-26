This is a full example using the resource

```go
package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/leandroluk/gox/oas"
	adapter "github.com/leandroluk/gox/oas/adapter/fiber"
)

func main() {
	app := adapter.Wrap(fiber.New()).OAS(func(b *adapter.DocumentBuilder) {
		b.OpenAPI("3.0.3")
		b.Info(func(i *oas.Info) {
			i.Title("My API").Version("1.0.0")
		})
		b.WithBearerToken("bearer")
		b.Server("http://localhost:3000", func(s *oas.Server) {
			s.Description("Development")
		})
	})

	api := app.Group("/api")

	// User group with tag and description registered in OAS
	users := api.Group("/users").OAS(func(b *adapter.GroupBuilder) {
		b.Tag("Users").Description("User management endpoints")
	})

	// Normal handlers
	users.Get("/:id", getUser)
	users.Put("/:id", updateUser)
	users.Get("/:id/:field", getUserField)

	// Overrides automatic documentation for the :id parameter using RouteBuilder
	users.Get("/:id/profile", func(r *adapter.RouteBuilder) {
		r.Summary("Get user profile").
			Parameter(func(p *oas.Parameter) {
				p.In("path").
					Name("id").
					Description("User UUID").
					Required(true).
					Schema(func(s *oas.Schema) { s.String().Format("uuid") })
			}).
			Response("200", func(r *oas.Response) {
				r.Description("Profile data").
					Json(func(m *oas.MediaType) {
						m.Schema(func(s *oas.Schema) {
							s.Object().
								Property("id", func(p *oas.Schema) { p.String().Format("uuid") }).
								Property("name", func(p *oas.Schema) { p.String() })
						})
					})
			}).
			Handlers(getUserProfile)
	})

	// Protected route with Bearer token using RouteBuilder
	users.Post("", func(r *adapter.RouteBuilder) {
		r.Summary("Create user").
			UseBearerToken("bearer", "users:write").
			RequestBody(func(rb *oas.RequestBody) {
				rb.Required(true).
					Json(func(m *oas.MediaType) {
						m.Schema(func(s *oas.Schema) {
							s.Object().
								Required("name", func(p *oas.Schema) { p.String() }).
								Required("email", func(p *oas.Schema) { p.String().Format("email") })
						})
					})
			}).
			Response("201", func(r *oas.Response) {
				r.Description("User created")
			}).
			Handlers(createUser)
	})

	// Serves the OAS document as JSON
	// Uses app.Document() to access the *oas.Document (private field exposed via method)
	app.Get("/swagger.json", func(c *fiber.Ctx) error {
		return c.JSON(app.Document())
	})

	app.Listen(":3000")
}

func getUser(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"id": c.Params("id")})
}

func updateUser(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"id": c.Params("id"), "updated": true})
}

func getUserField(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"id": c.Params("id"), "field": c.Params("field")})
}

func getUserProfile(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"id": c.Params("id"), "profile": "data"})
}

func createUser(c *fiber.Ctx) error {
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"created": true})
}
```