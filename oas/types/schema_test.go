package types_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/leandroluk/go/oas/types"
)

func TestSchema_Nullable(t *testing.T) {
	tests := []struct {
		name     string
		schema   *types.Schema
		expected string
	}{
		{
			name:     "should add null to string type",
			schema:   (&types.Schema{}).String().Nullable(),
			expected: `{"type":["string","null"]}`,
		},
		{
			name:     "should add null to object type",
			schema:   (&types.Schema{}).Object().Nullable(),
			expected: `{"type":["object","null"]}`,
		},
		{
			name:     "should not duplicate null if already present",
			schema:   (&types.Schema{}).String().Nullable().Nullable(),
			expected: `{"type":["string","null"]}`,
		},
		{
			name:     "should add null to empty type",
			schema:   (&types.Schema{}).Nullable(),
			expected: `{"type":["null"]}`,
		},
		{
			name:     "should add null to integer type",
			schema:   (&types.Schema{}).Integer().Nullable(),
			expected: `{"type":["integer","null"]}`,
		},
		{
			name:     "should add null to boolean type",
			schema:   (&types.Schema{}).Boolean().Nullable(),
			expected: `{"type":["boolean","null"]}`,
		},
		{
			name:     "should add null to array type",
			schema:   (&types.Schema{}).Array().Nullable(),
			expected: `{"type":["array","null"]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(tt.schema)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			// JSON order might differ, but for small arrays it's usually deterministic.
			// However a robust check would unmarshal and compare or use a json compare lib.
			// For this specific case with known simple struct, direct string compare might be flaky if spacing usage changes.
			// Let's normalize by unmarshaling to interface and comparing, or just string compare if confident.
			// Since we want to avoid deps, let's try direct string match but be removing generic spacing if any.
			got := string(b)
			// standard json.Marshal produces no spaces.

			// Check if keys are in expected order. Since it's a struct, "type" is the only field set,
			// so it should be easy.
			// Wait, omitempty might hide other fields.

			if got != tt.expected {
				// types slice order might vary? standard lib preserves order of slice.
				// map keys are sorted.

				// fallback: check if it contains expected parts
				if !strings.Contains(got, `"type"`) || !strings.Contains(got, `"null"`) {
					t.Errorf("expected %s, got %s", tt.expected, got)
				}

				// Strict check
				if got != tt.expected {
					// It's possible "null", "string" vs "string", "null"
					// My implementation appends null, so "string", "null" is expected if started with String().
					t.Errorf("expected %s, got %s", tt.expected, got)
				}
			}
		})
	}
}
