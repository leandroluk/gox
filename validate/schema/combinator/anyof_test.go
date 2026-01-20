// schema/combinator/anyof_test.go
package combinator_test

import (
	"testing"

	"github.com/leandroluk/go/validate/internal/testkit"
	"github.com/leandroluk/go/validate/schema"
	"github.com/leandroluk/go/validate/schema/combinator"
	"github.com/leandroluk/go/validate/schema/text"
)

func TestAnyOf_PassesWhenLaterSchemaPasses_EvenWithFailFast(t *testing.T) {
	s := combinator.AnyOf(
		text.New().Min(5),
		text.New().Pattern(`^\d+$`),
	)

	got, err := s.Validate("123", schema.WithFailFast(true))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != "123" {
		t.Fatalf("expected %q, got %q", "123", got)
	}
}

func TestAnyOf_FailDoesNotAccumulateIntermediateErrors(t *testing.T) {
	s := combinator.AnyOf(
		text.New().Min(5),
		text.New().Pattern(`^\d+$`),
	)

	_, err := s.Validate("ab")
	validationError := testkit.RequireValidationError(t, err)

	if len(validationError.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(validationError.Issues))
	}
}
