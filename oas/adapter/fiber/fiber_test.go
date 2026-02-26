package wrap

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/leandroluk/gox/oas"
)

func TestAdapter_GetRouteDocumentation(t *testing.T) {
	fiberApp := fiber.New()
	app := Adapter(fiberApp).OAS(func(b *DocumentBuilder) {
		b.OpenAPI("3.0.3")
		b.Info(func(i *oas.Info) {
			i.Title("Test API").Version("1.0.0")
		})
	})

	api := app.Group("/api")
	users := api.Group("/users").OAS(func(b *GroupBuilder) {
		b.Tag("Users").Description("Operations about user")
	})

	users.Get("/:id", func(c *fiber.Ctx) error {
		return c.SendString(c.Params("id"))
	}, func(op *oas.Operation) {
		op.Summary("Get user by ID")
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

func TestAdapter_SwaggerJSONEndpoint(t *testing.T) {
	fiberApp := fiber.New()
	app := Adapter(fiberApp)
	app.OAS(func(b *DocumentBuilder) {
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

func TestExtractPathParams(t *testing.T) {
	group := Group{}
	tests := []struct {
		route    string
		expected []string
	}{
		{"/users/:id", []string{"id"}},
		{"/users/:id/:field", []string{"id", "field"}},
		{"/files/:name?", []string{"name"}},
		{"/no/params", []string{}},
	}

	for _, tt := range tests {
		result := group.extractPathParams(tt.route)
		if len(result) != len(tt.expected) {
			t.Errorf("Extract params for %s failed, expected %d, got %d", tt.route, len(tt.expected), len(result))
		}
		for i, p := range result {
			if p != tt.expected[i] {
				t.Errorf("Expected param %s at index %d, got %s", tt.expected[i], i, p)
			}
		}
	}
}

func TestFiberPathToOAS(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/users/:id", "/users/{id}"},
		{"/users/:id?", "/users/{id}"},
		{"/users/:id/:field", "/users/{id}/{field}"},
		{"/search", "/search"},
	}

	for _, tt := range tests {
		result := fiberPathToOAS(tt.input)
		if result != tt.expected {
			t.Errorf("fiberPathToOAS(%s) = %s, expected %s", tt.input, result, tt.expected)
		}
	}
}
