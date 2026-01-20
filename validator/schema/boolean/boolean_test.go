// schema/boolean/boolean_test.go
package boolean_test

import (
	"testing"

	"github.com/leandroluk/go/validator/internal/ast"
	"github.com/leandroluk/go/validator/internal/testkit"
	"github.com/leandroluk/go/validator/schema"
	"github.com/leandroluk/go/validator/schema/boolean"
)

func TestBoolean_MissingAndNullAreIgnoredByDefault(t *testing.T) {
	s := boolean.New()

	if _, err := s.Validate(ast.MissingValue()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if _, err := s.Validate(ast.NullValue()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestBoolean_Required(t *testing.T) {
	s := boolean.New().Required()

	_, err := s.Validate(ast.NullValue())
	validationError := testkit.RequireValidationError(t, err)

	if validationError.Issues[0].Message != "required" {
		t.Fatalf("expected message %q, got %q", "required", validationError.Issues[0].Message)
	}
}

func TestBoolean_TypeMismatchMeta(t *testing.T) {
	s := boolean.New()

	_, err := s.Validate(ast.NumberValue("1"))
	validationError := testkit.RequireValidationError(t, err)

	issue := validationError.Issues[0]
	if issue.Meta["expected"] != "boolean" {
		t.Fatalf("expected meta.expected=%q, got %#v", "boolean", issue.Meta["expected"])
	}
	if issue.Meta["actual"] != "number" {
		t.Fatalf("expected meta.actual=%q, got %#v", "number", issue.Meta["actual"])
	}
}

func TestBoolean_Coerce(t *testing.T) {
	s := boolean.New()

	got, err := s.Validate(ast.StringValue("true"), schema.WithCoerce(true))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != true {
		t.Fatalf("expected true, got %v", got)
	}

	got, err = s.Validate(ast.NumberValue("0"), schema.WithCoerce(true))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != false {
		t.Fatalf("expected false, got %v", got)
	}
}

func TestBoolean_IsDefault_DoesNotErrorOnFalse(t *testing.T) {
	s := boolean.New().IsDefault()

	got, err := s.Validate(false)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != false {
		t.Fatalf("expected false, got %v", got)
	}
}

func TestBoolean_IsDefault_DoesNotSkipRequiredPresence(t *testing.T) {
	s := boolean.New().IsDefault().Required()

	_, err := s.Validate(ast.MissingValue())
	validationError := testkit.RequireValidationError(t, err)

	if validationError.Issues[0].Code != boolean.CodeRequired {
		t.Fatalf("expected code %q, got %q", boolean.CodeRequired, validationError.Issues[0].Code)
	}
}
