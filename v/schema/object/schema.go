// schema/object/schema.go
package object

import (
	"reflect"

	"github.com/leandroluk/go/v/internal/defaults"
	"github.com/leandroluk/go/v/internal/engine"
	"github.com/leandroluk/go/v/internal/ruleset"
	"github.com/leandroluk/go/v/schema"
	"github.com/leandroluk/go/v/schema/object/rule"
)

type ConditionOp = rule.ConditionOp

type Schema[T any] struct {
	required bool

	structOnly    bool
	noStructLevel bool

	defaultProvider defaults.Provider[T]

	fields []field[T]
	rules  []ruleset.RuleFn[T]

	lastFieldIndex int

	buildTarget *T
	buildError  error
}

func New[T any](builder func(target *T, schemaValue *Schema[T])) *Schema[T] {
	schemaValue := &Schema[T]{
		defaultProvider: defaults.None[T](),
		fields:          make([]field[T], 0),
		rules:           make([]ruleset.RuleFn[T], 0),
		lastFieldIndex:  -1,
	}

	var target T
	schemaValue.buildTarget = &target

	if builder != nil {
		builder(&target, schemaValue)
	}

	schemaValue.buildTarget = nil
	return schemaValue
}

func (s *Schema[T]) Required() *Schema[T] {
	s.required = true
	return s
}

func (s *Schema[T]) StructOnly() *Schema[T] {
	s.structOnly = true
	s.noStructLevel = false
	return s
}

func (s *Schema[T]) NoStructLevel() *Schema[T] {
	s.noStructLevel = true
	s.structOnly = false
	return s
}

func (s *Schema[T]) Default(value T) *Schema[T] {
	s.defaultProvider = defaults.Value(value)
	return s
}

func (s *Schema[T]) DefaultFunc(fn func() T) *Schema[T] {
	s.defaultProvider = defaults.Func(fn)
	return s
}

func (s *Schema[T]) Custom(ruleValue ruleset.RuleFn[T]) *Schema[T] {
	if ruleValue != nil {
		s.rules = append(s.rules, ruleValue)
	}
	return s
}

func (s *Schema[T]) Field(fieldPointer any, validator func(context *engine.Context, value any) (any, bool)) *Schema[T] {
	if s == nil {
		return s
	}
	if s.buildError != nil {
		return s
	}
	if s.buildTarget == nil {
		s.buildError = ErrInvalidBuilderUsage
		return s
	}

	compiled, err := newField(s.buildTarget, fieldPointer, validator)
	if err != nil {
		s.buildError = err
		return s
	}

	s.fields = append(s.fields, compiled)
	s.lastFieldIndex = len(s.fields) - 1
	return s
}

func (s *Schema[T]) RequiredIf(path string, op ConditionOp, expected any) *Schema[T] {
	fieldPointer := s.lastField()
	if fieldPointer == nil {
		s.buildError = ErrInvalidBuilderUsage
		return s
	}

	expectedValue, ok := anyToASTValue(expected)
	if !ok {
		s.buildError = rule.ErrUnsupportedExpectedValue
		return s
	}

	fieldPointer.requiredConditions = append(fieldPointer.requiredConditions, rule.RequiredIf(CodeFieldRequiredIf, path, op, expectedValue))
	return s
}

func (s *Schema[T]) RequiredWith(paths ...string) *Schema[T] {
	fieldPointer := s.lastField()
	if fieldPointer == nil {
		s.buildError = ErrInvalidBuilderUsage
		return s
	}

	fieldPointer.requiredConditions = append(fieldPointer.requiredConditions, rule.RequiredWith(CodeFieldRequiredWith, paths...))
	return s
}

func (s *Schema[T]) RequiredWithout(paths ...string) *Schema[T] {
	fieldPointer := s.lastField()
	if fieldPointer == nil {
		s.buildError = ErrInvalidBuilderUsage
		return s
	}

	fieldPointer.requiredConditions = append(fieldPointer.requiredConditions, rule.RequiredWithout(CodeFieldRequiredWithout, paths...))
	return s
}

func (s *Schema[T]) ExcludedIf(path string, op ConditionOp, expected any) *Schema[T] {
	fieldPointer := s.lastField()
	if fieldPointer == nil {
		s.buildError = ErrInvalidBuilderUsage
		return s
	}

	expectedValue, ok := anyToASTValue(expected)
	if !ok {
		s.buildError = rule.ErrUnsupportedExpectedValue
		return s
	}

	fieldPointer.excludedConditions = append(fieldPointer.excludedConditions, rule.ExcludedIf(CodeFieldExcludedIf, path, op, expectedValue))
	return s
}

func (s *Schema[T]) SkipUnless(path string, op ConditionOp, expected any) *Schema[T] {
	fieldPointer := s.lastField()
	if fieldPointer == nil {
		s.buildError = ErrInvalidBuilderUsage
		return s
	}

	expectedValue, ok := anyToASTValue(expected)
	if !ok {
		s.buildError = rule.ErrUnsupportedExpectedValue
		return s
	}

	fieldPointer.skipUnlessConditions = append(fieldPointer.skipUnlessConditions, rule.SkipUnless(path, op, expectedValue))
	return s
}

func (s *Schema[T]) EqField(other string) *Schema[T] {
	return s.addComparator(rule.EqField(CodeFieldEqField, other))
}
func (s *Schema[T]) NeField(other string) *Schema[T] {
	return s.addComparator(rule.NeField(CodeFieldNeField, other))
}
func (s *Schema[T]) GtField(other string) *Schema[T] {
	return s.addComparator(rule.GtField(CodeFieldGtField, other))
}
func (s *Schema[T]) GteField(other string) *Schema[T] {
	return s.addComparator(rule.GteField(CodeFieldGteField, other))
}
func (s *Schema[T]) LtField(other string) *Schema[T] {
	return s.addComparator(rule.LtField(CodeFieldLtField, other))
}
func (s *Schema[T]) LteField(other string) *Schema[T] {
	return s.addComparator(rule.LteField(CodeFieldLteField, other))
}

func (s *Schema[T]) EqCSField(path string) *Schema[T] {
	return s.addComparator(rule.EqCSField(CodeFieldEqCSField, path))
}
func (s *Schema[T]) NeCSField(path string) *Schema[T] {
	return s.addComparator(rule.NeCSField(CodeFieldNeCSField, path))
}
func (s *Schema[T]) GtCSField(path string) *Schema[T] {
	return s.addComparator(rule.GtCSField(CodeFieldGtCSField, path))
}
func (s *Schema[T]) GteCSField(path string) *Schema[T] {
	return s.addComparator(rule.GteCSField(CodeFieldGteCSField, path))
}
func (s *Schema[T]) LtCSField(path string) *Schema[T] {
	return s.addComparator(rule.LtCSField(CodeFieldLtCSField, path))
}
func (s *Schema[T]) LteCSField(path string) *Schema[T] {
	return s.addComparator(rule.LteCSField(CodeFieldLteCSField, path))
}

func (s *Schema[T]) FieldContains(other string) *Schema[T] {
	return s.addComparator(rule.FieldContains(CodeFieldContains, other))
}
func (s *Schema[T]) FieldExcludes(other string) *Schema[T] {
	return s.addComparator(rule.FieldExcludes(CodeFieldExcludes, other))
}

func (s *Schema[T]) addComparator(comparator rule.Comparator) *Schema[T] {
	fieldPointer := s.lastField()
	if fieldPointer == nil {
		s.buildError = ErrInvalidBuilderUsage
		return s
	}
	fieldPointer.comparators = append(fieldPointer.comparators, comparator)
	return s
}

func (s *Schema[T]) lastField() *field[T] {
	if s == nil {
		return nil
	}
	if s.lastFieldIndex < 0 || s.lastFieldIndex >= len(s.fields) {
		return nil
	}
	return &s.fields[s.lastFieldIndex]
}

func (s *Schema[T]) Validate(input any, optionList ...schema.Option) (T, error) {
	options := schema.ApplyOptions(optionList...)
	return s.validateWithOptions(input, options)
}

func (s *Schema[T]) ValidateAny(input any, options schema.Options) (any, error) {
	return s.validateWithOptions(input, options)
}

func (s *Schema[T]) OutputType() reflect.Type {
	var pointer *T
	return reflect.TypeOf(pointer).Elem()
}
