// schema/object/rule/condition_skip_unless.go
package rule

import (
	"github.com/leandroluk/gox/validate/internal/ast"
	"github.com/leandroluk/gox/validate/internal/engine"
)

func SkipUnless(path string, op ConditionOp, expected ast.Value) SkipUnlessCondition {
	return SkipUnlessCondition{
		path:     path,
		op:       op,
		expected: expected,
	}
}

func (c SkipUnlessCondition) ShouldSkip(_ *engine.Context, root ast.Value) bool {
	actual := ast.Query(root, c.path)
	return !conditionMet(actual, c.op, c.expected)
}
