// schema/object/rule/condition_excluded_if.go
package rule

import (
	"github.com/leandroluk/go/validator/internal/ast"
	"github.com/leandroluk/go/validator/internal/engine"
)

func ExcludedIf(code string, path string, op ConditionOp, expected ast.Value) ExcludedIfCondition {
	return ExcludedIfCondition{
		code:     code,
		path:     path,
		op:       op,
		expected: expected,
	}
}

func (c ExcludedIfCondition) Apply(context *engine.Context, root ast.Value, _ ast.Value, childPresent bool) (skip bool, stop bool) {
	actual := ast.Query(root, c.path)
	if !conditionMet(actual, c.op, c.expected) {
		return false, false
	}

	if childPresent {
		stop = context.AddIssueWithMeta(c.code, "excluded", map[string]any{
			"path":     c.path,
			"op":       string(c.op),
			"expected": astValueToMeta(c.expected),
		})
		return true, stop
	}

	return true, false
}
