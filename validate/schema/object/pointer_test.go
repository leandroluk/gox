// schema/object/pointer_test.go
package object_test

import (
	"encoding/json"
	"testing"

	"github.com/leandroluk/gox/validate/internal/ast"
	"github.com/leandroluk/gox/validate/schema/object"
)

// AWS-style Config with optional pointer fields
type AWSConfig struct {
	Region         string  `json:"region"`
	Bucket         string  `json:"bucket"`
	AccessKey      string  `json:"accessKey"`
	SecretKey      string  `json:"secretKey"`
	Endpoint       *string `json:"endpoint,omitempty"`
	ForcePathStyle bool    `json:"forcePathStyle"`
}

func stringPtr(s string) *string {
	return &s
}

func TestObject_PointerFields_PreserveValuesWhenMissing(t *testing.T) {
	schema := object.New(func(target *AWSConfig, s *object.Schema[AWSConfig]) {
		s.Field(&target.Region).Text().Default("us-east-1")
		s.Field(&target.Bucket).Text().Required()
		s.Field(&target.AccessKey).Text().Required()
		s.Field(&target.SecretKey).Text().Required()
		s.Field(&target.Endpoint).Text()
		s.Field(&target.ForcePathStyle).Boolean().Default(false)
	})

	// Input with pointer field populated
	rawConfig := AWSConfig{
		Region:         "us-east-1",
		Bucket:         "control",
		AccessKey:      "username",
		SecretKey:      "password",
		Endpoint:       stringPtr("http://localhost:39000"),
		ForcePathStyle: true,
	}

	cfg, err := schema.Validate(rawConfig)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Endpoint should be preserved
	if cfg.Endpoint == nil {
		t.Fatal("expected Endpoint to be preserved, got nil")
	}
	if *cfg.Endpoint != "http://localhost:39000" {
		t.Fatalf("expected Endpoint to be %q, got %q", "http://localhost:39000", *cfg.Endpoint)
	}

	// Other fields should also be present
	if cfg.Bucket != "control" {
		t.Fatalf("expected Bucket to be %q, got %q", "control", cfg.Bucket)
	}
	if cfg.ForcePathStyle != true {
		t.Fatal("expected ForcePathStyle to be true")
	}
}

func TestObject_PointerFields_ReplaceValuesWhenPresent(t *testing.T) {
	schema := object.New(func(target *AWSConfig, s *object.Schema[AWSConfig]) {
		s.Field(&target.Bucket).Text().Required()
		s.Field(&target.Endpoint).Text()
	})

	// Input with original pointer value

	// JSON with different endpoint
	jsonInput := `{"bucket": "control", "endpoint": "http://production:9000"}`

	cfg, err := schema.Validate(json.RawMessage(jsonInput))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Endpoint should be replaced with JSON value
	if cfg.Endpoint == nil {
		t.Fatal("expected Endpoint to be present, got nil")
	}
	if *cfg.Endpoint != "http://production:9000" {
		t.Fatalf("expected Endpoint to be %q, got %q", "http://production:9000", *cfg.Endpoint)
	}
}

func TestObject_PointerFields_SetNilWhenNull(t *testing.T) {
	schema := object.New(func(target *AWSConfig, s *object.Schema[AWSConfig]) {
		s.Field(&target.Bucket).Text().Required()
		s.Field(&target.Endpoint).Text()
	})

	// Input with original pointer value

	// JSON with null endpoint (explicit null should clear it)
	jsonInput := `{"bucket": "control", "endpoint": null}`

	cfg, err := schema.Validate(json.RawMessage(jsonInput))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Endpoint should be nil (explicitly set to null)
	if cfg.Endpoint != nil {
		t.Fatalf("expected Endpoint to be nil, got %v", cfg.Endpoint)
	}
}

func TestObject_PointerFields_SetNilWhenMissingAndNoOriginal(t *testing.T) {
	schema := object.New(func(target *AWSConfig, s *object.Schema[AWSConfig]) {
		s.Field(&target.Bucket).Text().Required()
		s.Field(&target.Endpoint).Text()
	})

	// JSON input without endpoint field
	jsonInput := `{"bucket": "control"}`

	cfg, err := schema.Validate(json.RawMessage(jsonInput))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Endpoint should be nil (no original value to preserve)
	if cfg.Endpoint != nil {
		t.Fatalf("expected Endpoint to be nil, got %v", cfg.Endpoint)
	}
}

func TestObject_PointerFields_RequiredValidation(t *testing.T) {
	schema := object.New(func(target *AWSConfig, s *object.Schema[AWSConfig]) {
		s.Field(&target.Bucket).Text().Required()
		s.Field(&target.Endpoint).Text().Required()
	})

	// JSON without required endpoint
	jsonInput := `{"bucket": "control"}`

	_, err := schema.Validate(json.RawMessage(jsonInput))
	if err == nil {
		t.Fatal("expected validation error for missing required field")
	}
}

func TestObject_PointerFields_WithASTValue(t *testing.T) {
	schema := object.New(func(target *AWSConfig, s *object.Schema[AWSConfig]) {
		s.Field(&target.Bucket).Text().Required()
		s.Field(&target.Endpoint).Text()
	})

	// Using AST directly (no original struct)
	astInput := ast.ObjectValue(map[string]ast.Value{
		"bucket": ast.StringValue("control"),
	})

	cfg, err := schema.Validate(astInput)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Endpoint should be nil (AST input, no original)
	if cfg.Endpoint != nil {
		t.Fatalf("expected Endpoint to be nil, got %v", cfg.Endpoint)
	}
}

type NestedConfig struct {
	Name     string  `json:"name"`
	Port     *int    `json:"port,omitempty"`
	Enabled  *bool   `json:"enabled,omitempty"`
	Timeout  *int64  `json:"timeout,omitempty"`
	Endpoint *string `json:"endpoint,omitempty"`
}

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

func int64Ptr(i int64) *int64 {
	return &i
}

func TestObject_MultiplePointerTypes_PreserveAll(t *testing.T) {
	schema := object.New(func(target *NestedConfig, s *object.Schema[NestedConfig]) {
		s.Field(&target.Name).Text().Required()
		s.Field(&target.Port).Number()
		s.Field(&target.Enabled).Boolean()
		s.Field(&target.Timeout).Number()
		s.Field(&target.Endpoint).Text()
	})

	rawConfig := NestedConfig{
		Name:     "test-service",
		Port:     intPtr(8080),
		Enabled:  boolPtr(true),
		Timeout:  int64Ptr(5000),
		Endpoint: stringPtr("http://api.example.com"),
	}

	cfg, err := schema.Validate(rawConfig)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// All pointer fields should be preserved
	if cfg.Port == nil || *cfg.Port != 8080 {
		t.Fatalf("expected Port to be 8080, got %v", cfg.Port)
	}
	if cfg.Enabled == nil || *cfg.Enabled != true {
		t.Fatalf("expected Enabled to be true, got %v", cfg.Enabled)
	}
	if cfg.Timeout == nil || *cfg.Timeout != 5000 {
		t.Fatalf("expected Timeout to be 5000, got %v", cfg.Timeout)
	}
	if cfg.Endpoint == nil || *cfg.Endpoint != "http://api.example.com" {
		t.Fatalf("expected Endpoint to be preserved, got %v", cfg.Endpoint)
	}
}

func TestObject_PointerFields_PartialUpdate(t *testing.T) {
	schema := object.New(func(target *NestedConfig, s *object.Schema[NestedConfig]) {
		s.Field(&target.Name).Text().Required()
		s.Field(&target.Port).Number()
		s.Field(&target.Enabled).Boolean()
		s.Field(&target.Endpoint).Text()
	})

	// JSON only updates port, other pointer fields should be preserved
	jsonInput := `{"name": "test-service", "port": 9000}`

	cfg, err := schema.Validate(json.RawMessage(jsonInput))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Port should be updated
	if cfg.Port == nil || *cfg.Port != 9000 {
		t.Fatalf("expected Port to be 9000, got %v", cfg.Port)
	}

	// Other original struct pointer values should NOT be preserved
	// (because we're validating JSON, not the original struct)
	if cfg.Enabled != nil {
		t.Fatalf("expected Enabled to be nil (JSON input doesn't preserve struct values), got %v", cfg.Enabled)
	}
	if cfg.Endpoint != nil {
		t.Fatalf("expected Endpoint to be nil (JSON input doesn't preserve struct values), got %v", cfg.Endpoint)
	}
}
