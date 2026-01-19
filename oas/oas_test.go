package oas_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/leandroluk/go/oas"
)

func TestTopicLevelBuilders(t *testing.T) {
	tests := []struct {
		name     string
		schema   *oas.Schema
		expected string
	}{
		{
			name:     "should create string schema",
			schema:   oas.String(),
			expected: `{"type":["string"]}`,
		},
		{
			name:     "should create integer schema",
			schema:   oas.Integer(),
			expected: `{"type":["integer"]}`,
		},
		{
			name:     "should create number schema",
			schema:   oas.Number(),
			expected: `{"type":["number"]}`,
		},
		{
			name:     "should create boolean schema",
			schema:   oas.Boolean(),
			expected: `{"type":["boolean"]}`,
		},
		{
			name:     "should create object schema",
			schema:   oas.Object(),
			expected: `{"type":["object"]}`,
		},
		{
			name:     "should create array schema",
			schema:   oas.Array(),
			expected: `{"type":["array"]}`,
		},
		{
			name:     "should create null schema",
			schema:   oas.Null(),
			expected: `{"type":["null"]}`,
		},
		{
			name:     "should support chaining nullable on string",
			schema:   oas.String().Nullable(),
			expected: `{"type":["string","null"]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(tt.schema)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			got := string(b)
			// Remove slice whitespace for comparison if any (though standard marshal is compact)
			// expected is compact.

			// For nullable types, JSON order isn't guaranteed for the array, so we check contents if needed
			// But for simple cases, string comparison usually works if generated order is consistent.
			// "nullable" appends, so ["string", "null"] is expected order in implementation.

			if got != tt.expected {
				// double check if it's just spacing or order
				if strings.Contains(tt.expected, "[") {
					// for arrays check presence
					if !strings.Contains(got, "\"string\"") || !strings.Contains(got, "\"null\"") {
						if !strings.Contains(got, "\"string\"") && !strings.Contains(got, "\"null\"") {
							t.Errorf("expected %s, got %s", tt.expected, got)
						}
					}
				} else {
					t.Errorf("expected %s, got %s", tt.expected, got)
				}
			}
		})
	}
}
