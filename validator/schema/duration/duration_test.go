// schema/duration/duration_test.go
package duration_test

import (
	"testing"
	"time"

	"github.com/leandroluk/go/validator/internal/ast"
	"github.com/leandroluk/go/validator/internal/testkit"
	"github.com/leandroluk/go/validator/schema"
	"github.com/leandroluk/go/validator/schema/duration"
)

func TestDuration_MissingAndNullAreIgnoredByDefault(t *testing.T) {
	s := duration.New()

	if _, err := s.Validate(ast.MissingValue()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if _, err := s.Validate(ast.NullValue()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestDuration_Required(t *testing.T) {
	s := duration.New().Required()

	_, err := s.Validate(ast.NullValue())
	validationError := testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Message != "required" {
		t.Fatalf("expected message %q, got %q", "required", validationError.Issues[0].Message)
	}
}

func TestDuration_TypeMismatchMeta(t *testing.T) {
	s := duration.New()

	_, err := s.Validate(ast.BooleanValue(true))
	validationError := testkit.RequireValidationError(t, err)

	issue := validationError.Issues[0]
	if issue.Meta["expected"] != "duration" {
		t.Fatalf("expected meta.expected=%q, got %#v", "duration", issue.Meta["expected"])
	}
	if issue.Meta["actual"] != "boolean" {
		t.Fatalf("expected meta.actual=%q, got %#v", "boolean", issue.Meta["actual"])
	}
}

func TestDuration_InvalidString(t *testing.T) {
	s := duration.New()

	_, err := s.Validate(ast.StringValue("nope"))
	validationError := testkit.RequireValidationError(t, err)

	issue := validationError.Issues[0]
	if issue.Meta["value"] != "nope" {
		t.Fatalf("expected meta.value=%q, got %#v", "nope", issue.Meta["value"])
	}
}

func TestDuration_ParseAndMinMax(t *testing.T) {
	s := duration.New().Min(2 * time.Second).Max(4 * time.Second)

	got, err := s.Validate(ast.StringValue("3s"))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != 3*time.Second {
		t.Fatalf("expected %v, got %v", 3*time.Second, got)
	}

	_, err = s.Validate(ast.StringValue("1s"))
	validationError := testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Meta["min"] == nil {
		t.Fatalf("expected meta.min")
	}

	_, err = s.Validate(ast.StringValue("5s"))
	validationError = testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Meta["max"] == nil {
		t.Fatalf("expected meta.max")
	}
}

func TestDuration_NumberIsNanoseconds(t *testing.T) {
	s := duration.New()

	got, err := s.Validate(ast.NumberValue("1000000000"))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != 1*time.Second {
		t.Fatalf("expected %v, got %v", 1*time.Second, got)
	}
}

func TestDuration_UnixSecondsAndMillis(t *testing.T) {
	s := duration.New()

	_, err := s.Validate(5)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	got, err := s.Validate(5, schema.WithCoerce(true), schema.WithCoerceDurationSeconds(true))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != 5*time.Second {
		t.Fatalf("expected %v, got %v", 5*time.Second, got)
	}

	got, err = s.Validate(5000, schema.WithCoerce(true), schema.WithCoerceDurationMilliseconds(true))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != 5000*time.Millisecond {
		t.Fatalf("expected %v, got %v", 5000*time.Millisecond, got)
	}
}

func TestDuration_Comparators(t *testing.T) {
	base := 3 * time.Second

	cases := []struct {
		name     string
		schema   *duration.Schema
		input    time.Duration
		wantCode string
	}{
		{"eq_ok", duration.New().Eq(base), base, ""},
		{"eq_fail", duration.New().Eq(base), base + time.Second, duration.CodeEq},

		{"ne_ok", duration.New().Ne(base), base + time.Second, ""},
		{"ne_fail", duration.New().Ne(base), base, duration.CodeNe},

		{"gt_ok", duration.New().Gt(base), base + time.Second, ""},
		{"gt_fail", duration.New().Gt(base), base, duration.CodeGt},

		{"gte_ok", duration.New().Gte(base), base, ""},
		{"gte_fail", duration.New().Gte(base), base - time.Second, duration.CodeGte},

		{"lt_ok", duration.New().Lt(base), base - time.Second, ""},
		{"lt_fail", duration.New().Lt(base), base, duration.CodeLt},

		{"lte_ok", duration.New().Lte(base), base, ""},
		{"lte_fail", duration.New().Lte(base), base + time.Second, duration.CodeLte},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.schema.Validate(tc.input)

			if tc.wantCode == "" {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				return
			}

			validationError := testkit.RequireValidationError(t, err)
			if validationError.Issues[0].Code != tc.wantCode {
				t.Fatalf("expected code %q, got %q", tc.wantCode, validationError.Issues[0].Code)
			}
		})
	}
}
