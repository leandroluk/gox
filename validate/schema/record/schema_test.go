// schema/record/record_test.go
package record_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/leandroluk/gox/validate/internal/ast"
	"github.com/leandroluk/gox/validate/internal/engine"
	"github.com/leandroluk/gox/validate/internal/issues"
	"github.com/leandroluk/gox/validate/internal/testkit"
	"github.com/leandroluk/gox/validate/schema/record"
	"github.com/leandroluk/gox/validate/schema/text"
)

func TestRecord_MissingAndNullAreIgnoredByDefault(t *testing.T) {
	s := record.New[int]().Min(1)

	if _, err := s.Validate(ast.MissingValue()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if _, err := s.Validate(ast.NullValue()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestRecord_Required(t *testing.T) {
	s := record.New[int]().Required()

	_, err := s.Validate(ast.NullValue())
	validationError := testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Message != "required" {
		t.Fatalf("expected message %q, got %q", "required", validationError.Issues[0].Message)
	}
}

func TestRecord_TypeMismatchMeta(t *testing.T) {
	s := record.New[int]()

	_, err := s.Validate(ast.StringValue("x"))
	validationError := testkit.RequireValidationError(t, err)

	issue := validationError.Issues[0]
	if issue.Meta["expected"] != "record" {
		t.Fatalf("expected meta.expected=%q, got %#v", "record", issue.Meta["expected"])
	}
	if issue.Meta["actual"] != "string" {
		t.Fatalf("expected meta.actual=%q, got %#v", "string", issue.Meta["actual"])
	}
}

func TestRecord_MinMax(t *testing.T) {
	s := record.New[int]().Min(2).Max(2)

	_, err := s.Validate(map[string]int{"a": 1})
	validationError := testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Meta["min"] != 2 {
		t.Fatalf("expected min=2, got %#v", validationError.Issues[0].Meta["min"])
	}

	_, err = s.Validate(map[string]int{"a": 1, "b": 2, "c": 3})
	validationError = testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Meta["max"] != 2 {
		t.Fatalf("expected max=2, got %#v", validationError.Issues[0].Meta["max"])
	}
}

func TestRecord_ValuePathUsesKey(t *testing.T) {
	s := record.New[int]().Values(func(context *engine.Context, value ast.Value) (int, bool) {
		stop := context.AddIssue("record.value.custom", "boom")
		return 0, stop
	})

	_, err := s.Validate(map[string]int{"a-b": 1})
	validationError := testkit.RequireValidationError(t, err)

	issue := validationError.Issues[0]
	if issue.Path != `["a-b"]` {
		t.Fatalf("expected path %q, got %q", `["a-b"]`, issue.Path)
	}
}

func TestRecord_Unique_FailsOnDuplicateValue(t *testing.T) {
	s := record.New[int]().Unique()

	_, err := s.Validate(map[string]int{"a": 1, "b": 1})
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
	if issue.Code != "record.unique" {
		t.Fatalf("expected code %q, got %q", "record.unique", issue.Code)
	}
}

func TestRecord_KeysSchema_PrefixesKeyCodesAndSkipsValueValidationWhenKeyFails(t *testing.T) {
	s := record.New[int]().
		Keys(text.New().Min(2)).
		Values(func(context *engine.Context, value ast.Value) (int, bool) {
			stop := context.AddIssue("value.run", "should not run")
			return 0, stop
		})

	_, err := s.Validate(map[string]int{"a": 1})
	validationError := testkit.RequireValidationError(t, err)

	if len(validationError.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(validationError.Issues))
	}

	issue := validationError.Issues[0]
	if issue.Path != `["a"]` {
		t.Fatalf("expected path %q, got %q", `["a"]`, issue.Path)
	}
	if issue.Code != "record.key.min" {
		t.Fatalf("expected code %q, got %q", "record.key.min", issue.Code)
	}
}

func TestRecord_KeysFunc_PrefixesKeyCodesAndSkipsValueValidationWhenKeyFails(t *testing.T) {
	s := record.New[int]().
		KeysFunc(func(context *engine.Context, key string) bool {
			if key == "ok" {
				return false
			}
			return context.AddIssue("custom", "bad key")
		}).
		Values(func(context *engine.Context, value ast.Value) (int, bool) {
			stop := context.AddIssue("value.run", "should not run")
			return 0, stop
		})

	_, err := s.Validate(map[string]int{"nope": 1})
	validationError := testkit.RequireValidationError(t, err)

	if len(validationError.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(validationError.Issues))
	}

	issue := validationError.Issues[0]
	if issue.Path != `["nope"]` {
		t.Fatalf("expected path %q, got %q", `["nope"]`, issue.Path)
	}
	if issue.Code != "record.key.custom" {
		t.Fatalf("expected code %q, got %q", "record.key.custom", issue.Code)
	}
	if !strings.HasPrefix(issue.Code, "record.key.") {
		t.Fatalf("expected key code prefix %q, got %q", "record.key.", issue.Code)
	}
}

func TestRecord_LengthComparators(t *testing.T) {
	cases := []struct {
		name     string
		schema   *record.Schema[int]
		input    map[string]int
		wantCode string
	}{
		{"len_ok", record.New[int]().Len(2), map[string]int{"a": 1, "b": 2}, ""},
		{"len_fail", record.New[int]().Len(2), map[string]int{"a": 1}, record.CodeLen},

		{"eq_ok", record.New[int]().Eq(2), map[string]int{"a": 1, "b": 2}, ""},
		{"eq_fail", record.New[int]().Eq(2), map[string]int{"a": 1}, record.CodeEq},

		{"ne_ok", record.New[int]().Ne(2), map[string]int{"a": 1}, ""},
		{"ne_fail", record.New[int]().Ne(2), map[string]int{"a": 1, "b": 2}, record.CodeNe},

		{"gt_ok", record.New[int]().Gt(2), map[string]int{"a": 1, "b": 2, "c": 3}, ""},
		{"gt_fail", record.New[int]().Gt(2), map[string]int{"a": 1, "b": 2}, record.CodeGt},

		{"gte_ok", record.New[int]().Gte(2), map[string]int{"a": 1, "b": 2}, ""},
		{"gte_fail", record.New[int]().Gte(2), map[string]int{"a": 1}, record.CodeGte},

		{"lt_ok", record.New[int]().Lt(2), map[string]int{"a": 1}, ""},
		{"lt_fail", record.New[int]().Lt(2), map[string]int{"a": 1, "b": 2}, record.CodeLt},

		{"lte_ok", record.New[int]().Lte(2), map[string]int{"a": 1, "b": 2}, ""},
		{"lte_fail", record.New[int]().Lte(2), map[string]int{"a": 1, "b": 2, "c": 3}, record.CodeLte},
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
