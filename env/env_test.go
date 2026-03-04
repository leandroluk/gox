package env_test

import (
	"os"
	"testing"

	"github.com/leandroluk/gox/env"
)

func TestEnv_LoadAndExpand(t *testing.T) {
	content := `
	ENV_FOO    =   "foo"
	ENV_BAR     =   "${ENV_FOO}.bar" # inline comment
	PORT = 8080
	`
	tmpFile := "test.env"
	os.WriteFile(tmpFile, []byte(content), 0644)
	defer os.Remove(tmpFile)

	env.Load(tmpFile)

	t.Run("Test Expansion and Spaces", func(t *testing.T) {
		bar := env.Get[string]("ENV_BAR")
		if bar != "foo.bar" {
			t.Errorf("Expected 'foo.bar', got %q", bar)
		}
	})

	t.Run("Test Type Conversion", func(t *testing.T) {
		port := env.Get[int]("PORT")
		if port != 8080 {
			t.Errorf("Expected 8080, got %d", port)
		}
	})

	t.Run("Test Default Value", func(t *testing.T) {
		missing := env.Get[int]("MISSING", 42)
		if missing != 42 {
			t.Errorf("Expected default 42, got %d", missing)
		}
	})
}
