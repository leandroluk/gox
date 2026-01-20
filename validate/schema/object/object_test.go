// schema/object/object_test.go
package object_test

import (
	"encoding/json"
	"testing"

	"github.com/leandroluk/go/validate/internal/ast"
	"github.com/leandroluk/go/validate/internal/ruleset"
	"github.com/leandroluk/go/validate/internal/testkit"
	"github.com/leandroluk/go/validate/schema/object"
	"github.com/leandroluk/go/validate/schema/object/rule"
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
		schemaValue.Field(&target.Name).Text().Required().
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
		schemaValue.Field(&target.Name).Text().Required()

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

	_, err = s.Validate(ast.ObjectValue(map[string]ast.Value{}))
	validationError := testkit.RequireValidationError(t, err)

	if validationError.Issues[0].Code != "text.required" {
		t.Fatalf("expected code %q, got %q", "text.required", validationError.Issues[0].Code)
	}
}

type Cross struct {
	A string `json:"a"`
	B string `json:"b"`
}

func TestObject_RequiredIf_DeepPathWithArrayIndex(t *testing.T) {
	s := object.New(func(target *Sample, schemaValue *object.Schema[Sample]) {
		schemaValue.Field(&target.Name).Text().Required().
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
		schemaValue.Field(&target.A).Text().RequiredWith("b")
		schemaValue.Field(&target.B).Text()
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
		schemaValue.Field(&target.A).Text().RequiredWithout("b")
		schemaValue.Field(&target.B).Text()
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
		schemaValue.Field(&target.A).Text().ExcludedIf("role", rule.OpEq, "admin")
		schemaValue.Field(&target.Role).Text()
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
	s := object.New(func(target *SkipUnless, schemaValue *object.Schema[SkipUnless]) {
		schemaValue.Field(&target.A).Text().Required().SkipUnless("flag", rule.OpEq, true)
		schemaValue.Field(&target.Flag).Boolean()
	})

	_, err := s.Validate(ast.ObjectValue(map[string]ast.Value{
		"flag": ast.BooleanValue(false),
		"a":    ast.NumberValue("1"),
	}))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

type Numbers struct {
	A int `json:"a"`
	B int `json:"b"`
}

func TestObject_EqField(t *testing.T) {
	s := object.New(func(target *Numbers, schemaValue *object.Schema[Numbers]) {
		schemaValue.Field(&target.A).Number().EqField("b")
		schemaValue.Field(&target.B).Number()
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
		schemaValue.Field(&target.A).Number().NeField("b")
		schemaValue.Field(&target.B).Number()
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
		schemaValue.Field(&target.A).Number().GtField("b")
		schemaValue.Field(&target.B).Number()
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
		schemaValue.Field(&target.A).Number().GteField("b")
		schemaValue.Field(&target.B).Number()
	})
	_, err = gte.Validate(ast.ObjectValue(map[string]ast.Value{
		"a": ast.NumberValue("2"),
		"b": ast.NumberValue("2"),
	}))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	lt := object.New(func(target *Numbers, schemaValue *object.Schema[Numbers]) {
		schemaValue.Field(&target.A).Number().LtField("b")
		schemaValue.Field(&target.B).Number()
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
		schemaValue.Field(&target.A).Number().LteField("b")
		schemaValue.Field(&target.B).Number()
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
		schemaValue.Field(&target.A).Number().EqCSField("meta.b")
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
		schemaValue.Field(&target.A).Text().FieldContains("b")
		schemaValue.Field(&target.B).Text()
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
		schemaValue.Field(&target.A).Text().FieldExcludes("b")
		schemaValue.Field(&target.B).Text()
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

// Testes da API fluente
type UserFluent struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestFieldBuilder_TextSchema(t *testing.T) {
	s := object.New(func(target *UserFluent, schemaValue *object.Schema[UserFluent]) {
		schemaValue.Field(&target.Name).Text().Required().Min(3).Max(50)
	})

	input := json.RawMessage(`{"name": "John"}`)
	out, err := s.Validate(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if out.Name != "John" {
		t.Fatalf("expected Name to be 'John', got %q", out.Name)
	}
}

func TestFieldBuilder_TextSchema_Validation(t *testing.T) {
	s := object.New(func(target *UserFluent, schemaValue *object.Schema[UserFluent]) {
		schemaValue.Field(&target.Name).Text().Required().Min(3)
	})

	input := json.RawMessage(`{"name": "Jo"}`)
	_, err := s.Validate(input)
	if err == nil {
		t.Fatalf("expected validation error for short name")
	}
}

func TestFieldBuilder_NumberSchema(t *testing.T) {
	s := object.New(func(target *UserFluent, schemaValue *object.Schema[UserFluent]) {
		schemaValue.Field(&target.Name).Text()
		schemaValue.Field(&target.Age).Number().Min(0).Max(130)
	})

	input := json.RawMessage(`{"name": "John", "age": 30}`)
	out, err := s.Validate(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if out.Age != 30 {
		t.Fatalf("expected Age to be 30, got %d", out.Age)
	}
}

func TestFieldBuilder_NumberSchema_Validation(t *testing.T) {
	s := object.New(func(target *UserFluent, schemaValue *object.Schema[UserFluent]) {
		schemaValue.Field(&target.Age).Number().Min(0).Max(130)
	})

	input := json.RawMessage(`{"age": 150}`)
	_, err := s.Validate(input)
	if err == nil {
		t.Fatalf("expected validation error for age > 130")
	}
}

type SimpleBool struct {
	Active bool `json:"active"`
}

func TestFieldBuilder_BooleanSchema(t *testing.T) {
	s := object.New(func(target *SimpleBool, schemaValue *object.Schema[SimpleBool]) {
		schemaValue.Field(&target.Active).Boolean().Required()
	})

	input := json.RawMessage(`{"active": true}`)
	out, err := s.Validate(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !out.Active {
		t.Fatalf("expected Active to be true")
	}
}
