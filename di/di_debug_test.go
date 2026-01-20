package di

import (
	"testing"
)

func TestDebugMode(t *testing.T) {
	// Enable debug mode
	Debug()

	if !debugMode {
		t.Error("Debug() did not set debugMode to true")
	}

	// Helper to capture stdout would be nice, but for now we ensures it doesn't panic on normal usage
	RegisterAs[string](func() string {
		return "test"
	})

	val := Resolve[string]()
	if val != "test" {
		t.Errorf("Expected 'test', got '%s'", val)
	}
}
