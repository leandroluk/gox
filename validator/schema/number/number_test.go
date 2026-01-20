// schema/number/number_test.go
package number_test

import (
	"errors"
	"math"
	"testing"

	"github.com/leandroluk/go/validator/internal/ast"
	"github.com/leandroluk/go/validator/internal/engine"
	"github.com/leandroluk/go/validator/internal/issues"
	"github.com/leandroluk/go/validator/internal/testkit"
	"github.com/leandroluk/go/validator/schema"
	"github.com/leandroluk/go/validator/schema/number"
)

func TestNumber_MissingAndNullAreIgnoredByDefault(t *testing.T) {
	s := number.New[int]().Min(1)

	if _, err := s.Validate(ast.MissingValue()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if _, err := s.Validate(ast.NullValue()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestNumber_Required(t *testing.T) {
	s := number.New[int]().Required()

	_, err := s.Validate(ast.MissingValue())
	validationError := testkit.RequireValidationError(t, err)

	if len(validationError.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(validationError.Issues))
	}
	if validationError.Issues[0].Message != "required" {
		t.Fatalf("expected message %q, got %q", "required", validationError.Issues[0].Message)
	}
}

func TestNumber_Default(t *testing.T) {
	s := number.New[int]().Default(7)

	got, err := s.Validate(ast.MissingValue())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != 7 {
		t.Fatalf("expected %d, got %d", 7, got)
	}

	got, err = s.Validate(ast.NullValue(), schema.WithDefaultOnNull(true))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != 7 {
		t.Fatalf("expected %d, got %d", 7, got)
	}
}

func TestNumber_TypeMismatchMeta(t *testing.T) {
	s := number.New[int]()

	_, err := s.Validate(ast.BooleanValue(true))
	validationError := testkit.RequireValidationError(t, err)

	issue := validationError.Issues[0]
	if issue.Meta["expected"] != "number" {
		t.Fatalf("expected meta.expected=%q, got %#v", "number", issue.Meta["expected"])
	}
	if issue.Meta["actual"] != "boolean" {
		t.Fatalf("expected meta.actual=%q, got %#v", "boolean", issue.Meta["actual"])
	}
}

func TestNumber_InvalidStringWithCoerce(t *testing.T) {
	s := number.New[int]()

	_, err := s.Validate(ast.StringValue("abc"), schema.WithCoerce(true))
	validationError := testkit.RequireValidationError(t, err)

	issue := validationError.Issues[0]
	if issue.Meta["value"] != "abc" {
		t.Fatalf("expected meta.value=%q, got %#v", "abc", issue.Meta["value"])
	}
}

func TestNumber_MinMax(t *testing.T) {
	s := number.New[int]().Min(10).Max(20)

	_, err := s.Validate(5)
	validationError := testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Meta["min"] != 10 {
		t.Fatalf("expected min=10, got %#v", validationError.Issues[0].Meta["min"])
	}

	_, err = s.Validate(25)
	validationError = testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Meta["max"] != 20 {
		t.Fatalf("expected max=20, got %#v", validationError.Issues[0].Meta["max"])
	}
}

func TestNumber_Comparators(t *testing.T) {
	cases := []struct {
		name     string
		schema   *number.Schema[int]
		input    int
		wantCode string
	}{
		{"eq_ok", number.New[int]().Eq(10), 10, ""},
		{"eq_fail", number.New[int]().Eq(10), 9, number.CodeEq},

		{"ne_ok", number.New[int]().Ne(10), 9, ""},
		{"ne_fail", number.New[int]().Ne(10), 10, number.CodeNe},

		{"gt_ok", number.New[int]().Gt(10), 11, ""},
		{"gt_fail", number.New[int]().Gt(10), 10, number.CodeGt},

		{"gte_ok", number.New[int]().Gte(10), 10, ""},
		{"gte_fail", number.New[int]().Gte(10), 9, number.CodeGte},

		{"lt_ok", number.New[int]().Lt(10), 9, ""},
		{"lt_fail", number.New[int]().Lt(10), 10, number.CodeLt},

		{"lte_ok", number.New[int]().Lte(10), 10, ""},
		{"lte_fail", number.New[int]().Lte(10), 11, number.CodeLte},
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

func TestNumber_ComparatorEq_AllowsNaNWhenExpectedNaN(t *testing.T) {
	s := number.New[float64]().Eq(math.NaN())

	got, err := s.Validate(math.NaN())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !math.IsNaN(got) {
		t.Fatalf("expected NaN output")
	}
}

func TestNumber_OneOf(t *testing.T) {
	s := number.New[int]().OneOf(1, 2, 3)

	_, err := s.Validate(4)
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
	if issue.Code != number.CodeOneOf {
		t.Fatalf("expected code %q, got %q", number.CodeOneOf, issue.Code)
	}
}

func TestNumber_CoerceTrimSpace_And_Underscore(t *testing.T) {
	s := number.New[int]()

	_, err := s.Validate(" 12 ", schema.WithCoerce(true))
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	got, err := s.Validate(" 12 ", schema.WithCoerce(true), schema.WithCoerceTrimSpace(true))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != 12 {
		t.Fatalf("expected %d, got %d", 12, got)
	}

	_, err = s.Validate("1_000", schema.WithCoerce(true))
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	got, err = s.Validate("1_000", schema.WithCoerce(true), schema.WithCoerceNumberUnderscore(true))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != 1000 {
		t.Fatalf("expected %d, got %d", 1000, got)
	}
}

func TestNumber_IsDefault_SkipsOtherRules_ButNotRequiredPresence(t *testing.T) {
	s := number.New[int]().IsDefault().Min(10).Eq(999)

	got, err := s.Validate(0)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != 0 {
		t.Fatalf("expected %d, got %d", 0, got)
	}

	_, err = s.Validate(5)
	validationError := testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Code != number.CodeMin {
		t.Fatalf("expected code %q, got %q", number.CodeMin, validationError.Issues[0].Code)
	}

	sRequired := number.New[int]().IsDefault().Required().Min(10)

	_, err = sRequired.Validate(ast.MissingValue())
	validationError = testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Code != number.CodeRequired {
		t.Fatalf("expected code %q, got %q", number.CodeRequired, validationError.Issues[0].Code)
	}

	got, err = sRequired.Validate(0)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != 0 {
		t.Fatalf("expected %d, got %d", 0, got)
	}
}

func TestEngine_OmitNil_Vs_OmitZero(t *testing.T) {
	type Payload struct {
		A *int   `json:"a"`
		B int    `json:"b"`
		C *int   `json:"c"`
		D string `json:"d,omitempty"`
		E string `json:"e"`
	}

	zero := 0
	payload := Payload{
		A: nil,
		B: 0,
		C: &zero,
		D: "",
		E: "x",
	}

	optionsNil := schema.ApplyOptions(schema.WithOmitNil(true))
	gotNil, err := engine.InputToASTWithOptions(payload, optionsNil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if gotNil.Kind != ast.KindObject {
		t.Fatalf("expected object, got %s", gotNil.Kind.String())
	}

	objNil := gotNil.Object
	if _, ok := objNil["a"]; ok {
		t.Fatalf("expected field a to be omitted by omitnil")
	}
	if _, ok := objNil["b"]; !ok {
		t.Fatalf("expected field b to be present with omitnil")
	}
	if _, ok := objNil["c"]; !ok {
		t.Fatalf("expected field c to be present with omitnil")
	}
	if _, ok := objNil["d"]; ok {
		t.Fatalf("expected field d to be omitted by json omitempty")
	}

	optionsZero := schema.ApplyOptions(schema.WithOmitZero(true))
	gotZero, err := engine.InputToASTWithOptions(payload, optionsZero)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if gotZero.Kind != ast.KindObject {
		t.Fatalf("expected object, got %s", gotZero.Kind.String())
	}

	objZero := gotZero.Object
	if _, ok := objZero["a"]; ok {
		t.Fatalf("expected field a to be omitted by omitzero (nil pointer is zero)")
	}
	if _, ok := objZero["b"]; ok {
		t.Fatalf("expected field b to be omitted by omitzero (0 is zero)")
	}
	if value, ok := objZero["c"]; !ok {
		t.Fatalf("expected field c to be present with omitzero (non-nil pointer is not zero)")
	} else if value.Kind != ast.KindNumber || value.Number != "0" {
		t.Fatalf("expected field c to be number(0), got kind=%s value=%q", value.Kind.String(), value.Number)
	}
	if value, ok := objZero["e"]; !ok {
		t.Fatalf("expected field e to be present")
	} else if value.Kind != ast.KindString || value.String != "x" {
		t.Fatalf("expected field e to be string(%q), got kind=%s value=%q", "x", value.Kind.String(), value.String)
	}
}
