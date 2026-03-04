package adapter_test

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/leandroluk/gox/oas"
	adapter "github.com/leandroluk/gox/oas/adapter/fiber"
)

func TestAdapter_GetRouteDocumentation(t *testing.T) {
	fiberApp := fiber.New()
	app := adapter.NewFiber(fiberApp).OAS(func(b *adapter.DocumentBuilder) {
		b.OpenAPI("3.0.3")
		b.Info(func(i *oas.Info) {
			i.Title("Test API").Version("1.0.0")
		})
	})

	api := app.Group("/api")
	users := api.Group("/users").OAS(func(b *adapter.GroupBuilder) {
		b.Tag("Users").Description("Operations about user")
	})

	users.Get("/:id", func(r *adapter.RouteBuilder) {
		r.Summary("Get user by ID")
		r.Handlers(func(c *fiber.Ctx) error {
			return c.SendString(c.Params("id"))
		})
	})

	doc := app.Document()
	if doc == nil {
		t.Fatal("expected document to be initialized, got nil")
	}

	bytes, err := doc.MarshalJSON()
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(bytes, &parsed); err != nil {
		t.Fatalf("failed to unmarshal generated JSON: %v", err)
	}

	info, ok := parsed["info"].(map[string]interface{})
	if !ok || info["title"] != "Test API" {
		t.Errorf("expected info.title to be 'Test API', got %v", info["title"])
	}

	paths, ok := parsed["paths"].(map[string]interface{})
	if !ok {
		t.Fatal("expected paths to be generated")
	}

	userPath, ok := paths["/api/users/{id}"].(map[string]interface{})
	if !ok {
		t.Fatal("expected path /api/users/{id} to exist")
	}

	getOp, ok := userPath["get"].(map[string]interface{})
	if !ok {
		t.Fatal("expected GET operation on /api/users/{id}")
	}

	if getOp["summary"] != "Get user by ID" {
		t.Errorf("expected summary 'Get user by ID', got %v", getOp["summary"])
	}
}

func TestAdapter_JSONEndpoint(t *testing.T) {
	fiberApp := fiber.New()
	app := adapter.NewFiber(fiberApp)
	app.OAS(func(b *adapter.DocumentBuilder) {
		b.OpenAPI("3.0.3")
		b.Info(func(i *oas.Info) {
			i.Title("Endpoint Test API").Version("1.0.0")
		})
	})

	app.Get("/swagger.json", func(c *fiber.Ctx) error {
		return c.JSON(app.Document())
	})

	req := httptest.NewRequest("GET", "/swagger.json", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON body: %v", err)
	}

	info, _ := parsed["info"].(map[string]interface{})
	if info["title"] != "Endpoint Test API" {
		t.Errorf("expected title to be 'Endpoint Test API', got %v", info["title"])
	}
}
