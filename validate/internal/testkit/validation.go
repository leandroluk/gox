// internal/testkit/validation.go
package testkit

import (
	"errors"
	"testing"

	"github.com/leandroluk/gox/validate/internal/issues"
)

func RequireValidationError(t *testing.T, err error) issues.ValidationError {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	var validationError issues.ValidationError
	if !errors.As(err, &validationError) {
		t.Fatalf("expected ValidationError, got %T: %v", err, err)
	}
	return validationError
}
