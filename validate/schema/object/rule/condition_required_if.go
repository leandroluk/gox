// schema/object/rule/condition_required_if.go
package rule

import (
	"github.com/leandroluk/gox/validate/internal/ast"
	"github.com/leandroluk/gox/validate/internal/engine"
)

type requiredIfCondition struct {
	code     string
	path     string
	op       ConditionOp
	expected ast.Value
}

func RequiredIf(code string, path string, op ConditionOp, expected ast.Value) RequiredCondition {
	return requiredIfCondition{
		code:     code,
		path:     path,
		op:       op,
		expected: expected,
	}
}

func (c requiredIfCondition) Apply(context *engine.Context, root ast.Value, _ ast.Value, childPresent bool) bool {
	if childPresent {
		return false
	}

	actual := ast.Query(root, c.path)
	if !conditionMet(actual, c.op, c.expected) {
		return false
	}

	return context.AddIssueWithMeta(c.code, "required", map[string]any{
		"path":     c.path,
		"op":       string(c.op),
		"expected": astValueToMeta(c.expected),
	})
}
