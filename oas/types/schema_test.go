package types_test

import (
	"encoding/json"
	"testing"

	"github.com/leandroluk/gox/oas/types"
)

func TestRequiredAndOptionalProperty(t *testing.T) {
	schema := &types.Schema{}
	schema.Object().
		Required("name", func(s *types.Schema) {
			s.String().MinLength(3).MaxLength(100)
		}).
		Required("email", func(s *types.Schema) {
			s.String().Format("email")
		}).
		Optional("age", func(s *types.Schema) {
			s.Integer().Minimum(0)
		}).
		Optional("bio", func(s *types.Schema) {
			s.String().MaxLength(500)
		})

	data, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("Failed to marshal schema: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	// Verifica se os campos obrigatórios estão presentes
	required, ok := result["required"].([]any)
	if !ok {
		t.Fatal("Expected 'required' field to be an array")
	}

	if len(required) != 2 {
		t.Fatalf("Expected 2 required fields, got %d", len(required))
	}

	hasName := false
	hasEmail := false
	for _, r := range required {
		switch r.(string) {
		case "name":
			hasName = true
		case "email":
			hasEmail = true
		}
	}

	if !hasName {
		t.Error("Expected 'name' to be in required fields")
	}
	if !hasEmail {
		t.Error("Expected 'email' to be in required fields")
	}

	// Verifica se todas as propriedades estão presentes
	properties, ok := result["properties"].(map[string]any)
	if !ok {
		t.Fatal("Expected 'properties' field to be an object")
	}

	expectedProps := []string{"name", "email", "age", "bio"}
	for _, prop := range expectedProps {
		if _, exists := properties[prop]; !exists {
			t.Errorf("Expected property '%s' to exist", prop)
		}
	}
}

func TestRequiredPropertyDoesNotDuplicate(t *testing.T) {
	schema := &types.Schema{}
	schema.Object().
		Required("name", func(s *types.Schema) {
			s.String()
		}).
		Required("name", func(s *types.Schema) {
			s.String().MinLength(5)
		})

	data, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("Failed to marshal schema: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	required, ok := result["required"].([]any)
	if !ok {
		t.Fatal("Expected 'required' field to be an array")
	}

	if len(required) != 1 {
		t.Fatalf("Expected 1 required field (no duplicates), got %d", len(required))
	}

	if required[0].(string) != "name" {
		t.Errorf("Expected required field to be 'name', got '%s'", required[0].(string))
	}
}

func TestNullableType(t *testing.T) {
	schema := &types.Schema{}
	schema.String().Nullable()

	data, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("Failed to marshal schema: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	if result["type"] != "string" {
		t.Errorf("Expected type 'string', got %v", result["type"])
	}

	if result["nullable"] != true {
		t.Errorf("Expected nullable true, got %v", result["nullable"])
	}
}
