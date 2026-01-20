// schema/array/array_test.go
package array_test

import (
	"errors"
	"testing"

	"github.com/leandroluk/go/validate/internal/ast"
	"github.com/leandroluk/go/validate/internal/engine"
	"github.com/leandroluk/go/validate/internal/issues"
	"github.com/leandroluk/go/validate/internal/testkit"
	"github.com/leandroluk/go/validate/schema"
	"github.com/leandroluk/go/validate/schema/array"
)

func TestArray_MissingAndNullAreIgnoredByDefault(t *testing.T) {
	s := array.New[string]().Min(1)

	if _, err := s.Validate(ast.MissingValue()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if _, err := s.Validate(ast.NullValue()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestArray_Required(t *testing.T) {
	s := array.New[string]().Required()

	_, err := s.Validate(ast.MissingValue())
	validationError := testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Message != "required" {
		t.Fatalf("expected message %q, got %q", "required", validationError.Issues[0].Message)
	}
}

func TestArray_TypeMismatchMeta(t *testing.T) {
	s := array.New[string]()

	_, err := s.Validate(ast.StringValue("x"))
	validationError := testkit.RequireValidationError(t, err)

	issue := validationError.Issues[0]
	if issue.Meta["expected"] != "array" {
		t.Fatalf("expected meta.expected=%q, got %#v", "array", issue.Meta["expected"])
	}
	if issue.Meta["actual"] != "string" {
		t.Fatalf("expected meta.actual=%q, got %#v", "string", issue.Meta["actual"])
	}
}

func TestArray_CoerceSingleton(t *testing.T) {
	s := array.New[string]()

	got, err := s.Validate(ast.StringValue("hello"), schema.WithCoerce(true))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got) != 1 || got[0] != "hello" {
		t.Fatalf("expected [hello], got %#v", got)
	}
}

func TestArray_MinMax(t *testing.T) {
	s := array.New[string]().Min(2).Max(3)

	_, err := s.Validate([]string{"a"})
	validationError := testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Meta["min"] != 2 {
		t.Fatalf("expected min=2, got %#v", validationError.Issues[0].Meta["min"])
	}

	_, err = s.Validate([]string{"a", "b", "c", "d"})
	validationError = testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Meta["max"] != 3 {
		t.Fatalf("expected max=3, got %#v", validationError.Issues[0].Meta["max"])
	}
}

func TestArray_ItemPathUsesIndex(t *testing.T) {
	s := array.New[string]().Items(func(context *engine.Context, value ast.Value) (string, bool) {
		stop := context.AddIssue("array.item.custom", "boom")
		return "", stop
	})

	_, err := s.Validate([]string{"a", "b"})
	validationError := testkit.RequireValidationError(t, err)

	issue := validationError.Issues[0]
	if issue.Path != "[0]" {
		t.Fatalf("expected path %q, got %q", "[0]", issue.Path)
	}
}

func TestArray_Unique_FailsOnDuplicate(t *testing.T) {
	s := array.New[int]().Unique()

	_, err := s.Validate([]int{1, 2, 1})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var validationError issues.ValidationError
	if !errors.As(err, &validationError) {
		t.Fatalf("expected ValidationError, got %T: %v", err, err)
	}

	if len(validationError.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(validationError.Issues))
	}

	issue := validationError.Issues[0]
	if issue.Code != "array.unique" {
		t.Fatalf("expected code %q, got %q", "array.unique", issue.Code)
	}
}

func TestArray_LengthComparators(t *testing.T) {
	cases := []struct {
		name     string
		schema   *array.Schema[int]
		input    []int
		wantCode string
	}{
		{"len_ok", array.New[int]().Len(2), []int{1, 2}, ""},
		{"len_fail", array.New[int]().Len(2), []int{1}, array.CodeLen},

		{"eq_ok", array.New[int]().Eq(2), []int{1, 2}, ""},
		{"eq_fail", array.New[int]().Eq(2), []int{1}, array.CodeEq},

		{"ne_ok", array.New[int]().Ne(2), []int{1}, ""},
		{"ne_fail", array.New[int]().Ne(2), []int{1, 2}, array.CodeNe},

		{"gt_ok", array.New[int]().Gt(2), []int{1, 2, 3}, ""},
		{"gt_fail", array.New[int]().Gt(2), []int{1, 2}, array.CodeGt},

		{"gte_ok", array.New[int]().Gte(2), []int{1, 2}, ""},
		{"gte_fail", array.New[int]().Gte(2), []int{1}, array.CodeGte},

		{"lt_ok", array.New[int]().Lt(2), []int{1}, ""},
		{"lt_fail", array.New[int]().Lt(2), []int{1, 2}, array.CodeLt},

		{"lte_ok", array.New[int]().Lte(2), []int{1, 2}, ""},
		{"lte_fail", array.New[int]().Lte(2), []int{1, 2, 3}, array.CodeLte},
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
