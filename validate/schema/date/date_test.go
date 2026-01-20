// schema/date/date_test.go
package date_test

import (
	"testing"
	"time"

	"github.com/leandroluk/gox/validate/internal/ast"
	"github.com/leandroluk/gox/validate/internal/testkit"
	"github.com/leandroluk/gox/validate/schema"
	"github.com/leandroluk/gox/validate/schema/date"
)

func TestDate_MissingAndNullAreIgnoredByDefault(t *testing.T) {
	s := date.New()

	if _, err := s.Validate(ast.MissingValue()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if _, err := s.Validate(ast.NullValue()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestDate_Required(t *testing.T) {
	s := date.New().Required()

	_, err := s.Validate(ast.MissingValue())
	validationError := testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Message != "required" {
		t.Fatalf("expected message %q, got %q", "required", validationError.Issues[0].Message)
	}
}

func TestDate_TypeMismatchMeta(t *testing.T) {
	s := date.New()

	_, err := s.Validate(ast.BooleanValue(true))
	validationError := testkit.RequireValidationError(t, err)

	issue := validationError.Issues[0]
	if issue.Meta["expected"] != "date" {
		t.Fatalf("expected meta.expected=%q, got %#v", "date", issue.Meta["expected"])
	}
	if issue.Meta["actual"] != "boolean" {
		t.Fatalf("expected meta.actual=%q, got %#v", "boolean", issue.Meta["actual"])
	}
}

func TestDate_InvalidString(t *testing.T) {
	s := date.New()

	_, err := s.Validate(ast.StringValue("nope"), schema.WithDateLayouts(time.RFC3339Nano))
	validationError := testkit.RequireValidationError(t, err)

	issue := validationError.Issues[0]
	if issue.Meta["value"] != "nope" {
		t.Fatalf("expected meta.value=%q, got %#v", "nope", issue.Meta["value"])
	}
}

func TestDate_ParseAndMinMax(t *testing.T) {
	s := date.New().
		Min(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)).
		Max(time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC))

	input := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC).Format(time.RFC3339Nano)

	got, err := s.Validate(ast.StringValue(input), schema.WithDateLayouts(time.RFC3339Nano))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !got.Equal(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected parsed value: %v", got)
	}

	_, err = s.Validate(
		ast.StringValue(time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC).Format(time.RFC3339Nano)),
		schema.WithDateLayouts(time.RFC3339Nano),
	)
	validationError := testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Meta["min"] == nil {
		t.Fatalf("expected meta.min")
	}

	_, err = s.Validate(
		ast.StringValue(time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC).Format(time.RFC3339Nano)),
		schema.WithDateLayouts(time.RFC3339Nano),
	)
	validationError = testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Meta["max"] == nil {
		t.Fatalf("expected meta.max")
	}
}

func TestDate_CoerceUnixSeconds(t *testing.T) {
	s := date.New()

	got, err := s.Validate(ast.NumberValue("1700000000"), schema.WithCoerce(true))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	expected := time.Unix(1700000000, 0).UTC()
	if !got.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

func TestDate_AdditionalLayout_And_Unix(t *testing.T) {
	s := date.New()

	_, err := s.Validate("2026-01-03 12:34:56")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	_, err = s.Validate("2026-01-03 12:34:56", schema.WithAdditionalDateLayouts("2006-01-02 15:04:05"))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	_, err = s.Validate(1700000000)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	_, err = s.Validate(1700000000, schema.WithCoerce(true), schema.WithCoerceDateUnixSeconds(true))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestDate_Comparators(t *testing.T) {
	base := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	layout := time.RFC3339Nano

	cases := []struct {
		name     string
		schema   *date.Schema
		input    time.Time
		wantCode string
	}{
		{"eq_ok", date.New().Eq(base), base, ""},
		{"eq_fail", date.New().Eq(base), base.Add(time.Second), date.CodeEq},

		{"ne_ok", date.New().Ne(base), base.Add(time.Second), ""},
		{"ne_fail", date.New().Ne(base), base, date.CodeNe},

		{"gt_ok", date.New().Gt(base), base.Add(time.Second), ""},
		{"gt_fail", date.New().Gt(base), base, date.CodeGt},

		{"gte_ok", date.New().Gte(base), base, ""},
		{"gte_fail", date.New().Gte(base), base.Add(-time.Second), date.CodeGte},

		{"lt_ok", date.New().Lt(base), base.Add(-time.Second), ""},
		{"lt_fail", date.New().Lt(base), base, date.CodeLt},

		{"lte_ok", date.New().Lte(base), base, ""},
		{"lte_fail", date.New().Lte(base), base.Add(time.Second), date.CodeLte},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.schema.Validate(tc.input.Format(layout), schema.WithDateLayouts(layout))

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

func TestDate_DateTimeOnly_RejectsDateOnlyLayout(t *testing.T) {
	s := date.New().DateTime()

	_, err := s.Validate("2024-01-02", schema.WithDateLayouts("2006-01-02"))
	validationError := testkit.RequireValidationError(t, err)

	if validationError.Issues[0].Code != date.CodeDateTime {
		t.Fatalf("expected code %q, got %q", date.CodeDateTime, validationError.Issues[0].Code)
	}

	_, err = s.Validate("2024-01-02T03:04:05Z", schema.WithDateLayouts(time.RFC3339))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestDate_TimeLocation_AppliesToLayoutsWithoutZone(t *testing.T) {
	loc := time.FixedZone("BRT", -3*3600)
	s := date.New()

	got, err := s.Validate(
		"2024-01-02 12:34:56",
		schema.WithDateLayouts("2006-01-02 15:04:05"),
		schema.WithTimeLocation(loc),
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	expected := time.Date(2024, 1, 2, 12, 34, 56, 0, loc)
	if !got.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}
