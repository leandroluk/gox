// schema/combinator/oneof_test.go
package combinator_test

import (
	"testing"

	"github.com/leandroluk/gox/validate/internal/testkit"
	"github.com/leandroluk/gox/validate/schema/combinator"
	"github.com/leandroluk/gox/validate/schema/text"
)

func TestOneOf_PassesWhenExactlyOneMatches(t *testing.T) {
	s := combinator.OneOf(
		text.New().Pattern(`^a$`),
		text.New().Pattern(`^b$`),
	)

	got, err := s.Validate("a")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != "a" {
		t.Fatalf("expected %q, got %q", "a", got)
	}
}

func TestOneOf_FailsWhenNoneMatches(t *testing.T) {
	s := combinator.OneOf(
		text.New().Pattern(`^a$`),
		text.New().Pattern(`^b$`),
	)

	_, err := s.Validate("c")
	validationError := testkit.RequireValidationError(t, err)

	if len(validationError.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(validationError.Issues))
	}
}

func TestOneOf_FailsWhenMoreThanOneMatches(t *testing.T) {
	s := combinator.OneOf(
		text.New().Min(1),
		text.New().Pattern(`^[a-z]+$`),
	)

	_, err := s.Validate("abc")
	validationError := testkit.RequireValidationError(t, err)

	if len(validationError.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(validationError.Issues))
	}
	if validationError.Issues[0].Code != combinator.CodeOneOf {
		t.Fatalf("expected code %q, got %q", combinator.CodeOneOf, validationError.Issues[0].Code)
	}
}
