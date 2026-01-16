// schema/object/object_test.go
package object_test

import (
	"testing"

	"github.com/leandroluk/go/v/internal/ast"
	"github.com/leandroluk/go/v/internal/engine"
	"github.com/leandroluk/go/v/internal/ruleset"
	"github.com/leandroluk/go/v/internal/testkit"
	"github.com/leandroluk/go/v/schema/object"
	"github.com/leandroluk/go/v/schema/object/rule"
)

type Sample struct {
	Name string `json:"name"`
}

func TestObject_MissingAndNullAreIgnoredByDefault(t *testing.T) {
	s := object.New(func(target *Sample, schemaValue *object.Schema[Sample]) {})

	if _, err := s.Validate(ast.MissingValue()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if _, err := s.Validate(ast.NullValue()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestObject_Required(t *testing.T) {
	s := object.New(func(target *Sample, schemaValue *object.Schema[Sample]) {}).Required()

	_, err := s.Validate(ast.NullValue())
	validationError := testkit.RequireValidationError(t, err)

	if validationError.Issues[0].Message != "required" {
		t.Fatalf("expected message %q, got %q", "required", validationError.Issues[0].Message)
	}
}

func TestObject_TypeMismatchMeta(t *testing.T) {
	s := object.New(func(target *Sample, schemaValue *object.Schema[Sample]) {})

	_, err := s.Validate(ast.StringValue("x"))
	validationError := testkit.RequireValidationError(t, err)

	issue := validationError.Issues[0]
	if issue.Meta["expected"] != "object" {
		t.Fatalf("expected meta.expected=%q, got %#v", "object", issue.Meta["expected"])
	}
	if issue.Meta["actual"] != "string" {
		t.Fatalf("expected meta.actual=%q, got %#v", "string", issue.Meta["actual"])
	}
}

func TestObject_StructOnly_DoesNotValidateFieldsOrFieldConditions(t *testing.T) {
	s := object.New(func(target *Sample, schemaValue *object.Schema[Sample]) {
		schemaValue.
			Field(&target.Name, func(context *engine.Context, value any) (any, bool) {
				stop := context.AddIssue("field.validator", "should not run")
				return "", stop
			}).
			RequiredIf("other", rule.OpPresent, nil)
	}).StructOnly()

	_, err := s.Validate(ast.ObjectValue(map[string]ast.Value{
		"name":  ast.StringValue("ok"),
		"other": ast.BooleanValue(true),
	}))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	_, err = s.Validate(ast.ObjectValue(map[string]ast.Value{
		"other": ast.BooleanValue(true),
	}))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestObject_NoStructLevel_ValidatesFieldsButSkipsObjectRules(t *testing.T) {
	s := object.New(func(target *Sample, schemaValue *object.Schema[Sample]) {
		schemaValue.Field(&target.Name, func(context *engine.Context, value any) (any, bool) {
			v := value.(ast.Value)

			if v.Kind != ast.KindString {
				stop := context.AddIssue("field.validator", "invalid")
				return "", stop
			}
			return v.String, false
		})
		schemaValue.Custom(func(value Sample, reporter ruleset.Reporter) bool {
			return reporter.AddIssue("struct.rule", "should not run")
		})
	}).NoStructLevel()

	_, err := s.Validate(ast.ObjectValue(map[string]ast.Value{
		"name": ast.StringValue("ok"),
	}))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	_, err = s.Validate(ast.ObjectValue(map[string]ast.Value{
		"name": ast.NumberValue("1"),
	}))
	validationError := testkit.RequireValidationError(t, err)

	if validationError.Issues[0].Code != "field.validator" {
		t.Fatalf("expected code %q, got %q", "field.validator", validationError.Issues[0].Code)
	}
}

type Cross struct {
	A string `json:"a"`
	B string `json:"b"`
}

func TestObject_RequiredIf_DeepPathWithArrayIndex(t *testing.T) {
	s := object.New(func(target *Sample, schemaValue *object.Schema[Sample]) {
		schemaValue.
			Field(&target.Name, func(context *engine.Context, value any) (any, bool) {
				v := value.(ast.Value)

				if v.Kind != ast.KindString {
					stop := context.AddIssue("field.type", "invalid")
					return "", stop
				}
				return v.String, false
			}).
			RequiredIf("meta.items[0].flag", rule.OpEq, true)
	})

	_, err := s.Validate(ast.ObjectValue(map[string]ast.Value{
		"meta": ast.ObjectValue(map[string]ast.Value{
			"items": ast.ArrayValue([]ast.Value{
				ast.ObjectValue(map[string]ast.Value{
					"flag": ast.BooleanValue(true),
				}),
			}),
		}),
	}))
	validationError := testkit.RequireValidationError(t, err)

	issue := validationError.Issues[0]
	if issue.Path != "name" {
		t.Fatalf("expected path %q, got %q", "name", issue.Path)
	}
	if issue.Code != object.CodeFieldRequiredIf {
		t.Fatalf("expected code %q, got %q", object.CodeFieldRequiredIf, issue.Code)
	}
}

func TestObject_RequiredWith_MakesFieldRequiredWhenOtherPresent(t *testing.T) {
	s := object.New(func(target *Cross, schemaValue *object.Schema[Cross]) {
		schemaValue.
			Field(&target.A, nil).
			RequiredWith("b")
		schemaValue.Field(&target.B, nil)
	})

	_, err := s.Validate(ast.ObjectValue(map[string]ast.Value{
		"b": ast.StringValue("x"),
	}))
	validationError := testkit.RequireValidationError(t, err)

	issue := validationError.Issues[0]
	if issue.Path != "a" {
		t.Fatalf("expected path %q, got %q", "a", issue.Path)
	}
	if issue.Code != object.CodeFieldRequiredWith {
		t.Fatalf("expected code %q, got %q", object.CodeFieldRequiredWith, issue.Code)
	}
}

func TestObject_RequiredWithout_MakesFieldRequiredWhenOtherMissing(t *testing.T) {
	s := object.New(func(target *Cross, schemaValue *object.Schema[Cross]) {
		schemaValue.
			Field(&target.A, nil).
			RequiredWithout("b")
		schemaValue.Field(&target.B, nil)
	})

	_, err := s.Validate(ast.ObjectValue(map[string]ast.Value{}))
	validationError := testkit.RequireValidationError(t, err)

	issue := validationError.Issues[0]
	if issue.Path != "a" {
		t.Fatalf("expected path %q, got %q", "a", issue.Path)
	}
	if issue.Code != object.CodeFieldRequiredWithout {
		t.Fatalf("expected code %q, got %q", object.CodeFieldRequiredWithout, issue.Code)
	}

	_, err = s.Validate(ast.ObjectValue(map[string]ast.Value{
		"b": ast.StringValue("ok"),
	}))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

type Excluded struct {
	A    string `json:"a"`
	Role string `json:"role"`
}

func TestObject_ExcludedIf_FailsWhenConditionMetAndFieldPresent(t *testing.T) {
	s := object.New(func(target *Excluded, schemaValue *object.Schema[Excluded]) {
		schemaValue.
			Field(&target.A, nil).
			ExcludedIf("role", rule.OpEq, "admin")
		schemaValue.Field(&target.Role, nil)
	})

	_, err := s.Validate(ast.ObjectValue(map[string]ast.Value{
		"role": ast.StringValue("admin"),
		"a":    ast.StringValue("x"),
	}))
	validationError := testkit.RequireValidationError(t, err)

	issue := validationError.Issues[0]
	if issue.Path != "a" {
		t.Fatalf("expected path %q, got %q", "a", issue.Path)
	}
	if issue.Code != object.CodeFieldExcludedIf {
		t.Fatalf("expected code %q, got %q", object.CodeFieldExcludedIf, issue.Code)
	}

	_, err = s.Validate(ast.ObjectValue(map[string]ast.Value{
		"role": ast.StringValue("admin"),
	}))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

type SkipUnless struct {
	A    string `json:"a"`
	Flag bool   `json:"flag"`
}

func TestObject_SkipUnless_SkipsFieldValidationWhenNotMet(t *testing.T) {
	called := false

	s := object.New(func(target *SkipUnless, schemaValue *object.Schema[SkipUnless]) {
		schemaValue.
			Field(&target.A, func(context *engine.Context, value any) (any, bool) {
				called = true
				stop := context.AddIssue("should.not.run", "boom")
				return "", stop
			}).
			SkipUnless("flag", rule.OpEq, true)

		schemaValue.Field(&target.Flag, nil)
	})

	_, err := s.Validate(ast.ObjectValue(map[string]ast.Value{
		"flag": ast.BooleanValue(false),
		"a":    ast.NumberValue("1"),
	}))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if called {
		t.Fatalf("expected validator not to be called")
	}
}

type Numbers struct {
	A int `json:"a"`
	B int `json:"b"`
}

func TestObject_EqField(t *testing.T) {
	s := object.New(func(target *Numbers, schemaValue *object.Schema[Numbers]) {
		schemaValue.Field(&target.A, nil).EqField("b")
		schemaValue.Field(&target.B, nil)
	})

	_, err := s.Validate(ast.ObjectValue(map[string]ast.Value{
		"a": ast.NumberValue("1"),
		"b": ast.NumberValue("1"),
	}))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	_, err = s.Validate(ast.ObjectValue(map[string]ast.Value{
		"a": ast.NumberValue("1"),
		"b": ast.NumberValue("2"),
	}))
	validationError := testkit.RequireValidationError(t, err)

	issue := validationError.Issues[0]
	if issue.Path != "a" {
		t.Fatalf("expected path %q, got %q", "a", issue.Path)
	}
	if issue.Code != object.CodeFieldEqField {
		t.Fatalf("expected code %q, got %q", object.CodeFieldEqField, issue.Code)
	}
}

func TestObject_NeField(t *testing.T) {
	s := object.New(func(target *Numbers, schemaValue *object.Schema[Numbers]) {
		schemaValue.Field(&target.A, nil).NeField("b")
		schemaValue.Field(&target.B, nil)
	})

	_, err := s.Validate(ast.ObjectValue(map[string]ast.Value{
		"a": ast.NumberValue("1"),
		"b": ast.NumberValue("2"),
	}))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	_, err = s.Validate(ast.ObjectValue(map[string]ast.Value{
		"a": ast.NumberValue("1"),
		"b": ast.NumberValue("1"),
	}))
	validationError := testkit.RequireValidationError(t, err)

	if validationError.Issues[0].Code != object.CodeFieldNeField {
		t.Fatalf("expected code %q, got %q", object.CodeFieldNeField, validationError.Issues[0].Code)
	}
}

func TestObject_GtGteLtLteField(t *testing.T) {
	gt := object.New(func(target *Numbers, schemaValue *object.Schema[Numbers]) {
		schemaValue.Field(&target.A, nil).GtField("b")
		schemaValue.Field(&target.B, nil)
	})
	_, err := gt.Validate(ast.ObjectValue(map[string]ast.Value{
		"a": ast.NumberValue("2"),
		"b": ast.NumberValue("1"),
	}))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	_, err = gt.Validate(ast.ObjectValue(map[string]ast.Value{
		"a": ast.NumberValue("1"),
		"b": ast.NumberValue("2"),
	}))
	if testkit.RequireValidationError(t, err).Issues[0].Code != object.CodeFieldGtField {
		t.Fatalf("expected code %q", object.CodeFieldGtField)
	}

	gte := object.New(func(target *Numbers, schemaValue *object.Schema[Numbers]) {
		schemaValue.Field(&target.A, nil).GteField("b")
		schemaValue.Field(&target.B, nil)
	})
	_, err = gte.Validate(ast.ObjectValue(map[string]ast.Value{
		"a": ast.NumberValue("2"),
		"b": ast.NumberValue("2"),
	}))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	lt := object.New(func(target *Numbers, schemaValue *object.Schema[Numbers]) {
		schemaValue.Field(&target.A, nil).LtField("b")
		schemaValue.Field(&target.B, nil)
	})
	_, err = lt.Validate(ast.ObjectValue(map[string]ast.Value{
		"a": ast.NumberValue("1"),
		"b": ast.NumberValue("2"),
	}))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	_, err = lt.Validate(ast.ObjectValue(map[string]ast.Value{
		"a": ast.NumberValue("2"),
		"b": ast.NumberValue("1"),
	}))
	if testkit.RequireValidationError(t, err).Issues[0].Code != object.CodeFieldLtField {
		t.Fatalf("expected code %q", object.CodeFieldLtField)
	}

	lte := object.New(func(target *Numbers, schemaValue *object.Schema[Numbers]) {
		schemaValue.Field(&target.A, nil).LteField("b")
		schemaValue.Field(&target.B, nil)
	})
	_, err = lte.Validate(ast.ObjectValue(map[string]ast.Value{
		"a": ast.NumberValue("2"),
		"b": ast.NumberValue("2"),
	}))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

type OnlyA struct {
	A int `json:"a"`
}

func TestObject_EqCSField(t *testing.T) {
	s := object.New(func(target *OnlyA, schemaValue *object.Schema[OnlyA]) {
		schemaValue.Field(&target.A, nil).EqCSField("meta.b")
	})

	_, err := s.Validate(ast.ObjectValue(map[string]ast.Value{
		"a": ast.NumberValue("1"),
		"meta": ast.ObjectValue(map[string]ast.Value{
			"b": ast.NumberValue("1"),
		}),
	}))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	_, err = s.Validate(ast.ObjectValue(map[string]ast.Value{
		"a": ast.NumberValue("1"),
		"meta": ast.ObjectValue(map[string]ast.Value{
			"b": ast.NumberValue("2"),
		}),
	}))
	if testkit.RequireValidationError(t, err).Issues[0].Code != object.CodeFieldEqCSField {
		t.Fatalf("expected code %q", object.CodeFieldEqCSField)
	}
}

type Texts struct {
	A string `json:"a"`
	B string `json:"b"`
}

func TestObject_FieldContainsAndExcludes(t *testing.T) {
	contains := object.New(func(target *Texts, schemaValue *object.Schema[Texts]) {
		schemaValue.Field(&target.A, nil).FieldContains("b")
		schemaValue.Field(&target.B, nil)
	})

	_, err := contains.Validate(ast.ObjectValue(map[string]ast.Value{
		"a": ast.StringValue("hello world"),
		"b": ast.StringValue("world"),
	}))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	_, err = contains.Validate(ast.ObjectValue(map[string]ast.Value{
		"a": ast.StringValue("hello"),
		"b": ast.StringValue("world"),
	}))
	if testkit.RequireValidationError(t, err).Issues[0].Code != object.CodeFieldContains {
		t.Fatalf("expected code %q", object.CodeFieldContains)
	}

	excludes := object.New(func(target *Texts, schemaValue *object.Schema[Texts]) {
		schemaValue.Field(&target.A, nil).FieldExcludes("b")
		schemaValue.Field(&target.B, nil)
	})

	_, err = excludes.Validate(ast.ObjectValue(map[string]ast.Value{
		"a": ast.StringValue("hello"),
		"b": ast.StringValue("world"),
	}))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	_, err = excludes.Validate(ast.ObjectValue(map[string]ast.Value{
		"a": ast.StringValue("hello world"),
		"b": ast.StringValue("world"),
	}))
	if testkit.RequireValidationError(t, err).Issues[0].Code != object.CodeFieldExcludes {
		t.Fatalf("expected code %q", object.CodeFieldExcludes)
	}
}
