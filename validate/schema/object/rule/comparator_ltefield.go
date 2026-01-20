// schema/object/rule/comparator_ltefield.go
package rule

import (
	"github.com/leandroluk/go/validate/internal/ast"
	"github.com/leandroluk/go/validate/internal/engine"
)

type lteFieldComparator struct {
	code  string
	other string
}

func LteField(code string, other string) Comparator {
	return lteFieldComparator{code: code, other: other}
}

func (c lteFieldComparator) Apply(context *engine.Context, root ast.Value, child ast.Value) bool {
	otherValue := ast.Query(root, c.other)

	order, ok := compareOrder(child, otherValue)
	if ok && order <= 0 {
		return false
	}

	return context.AddIssueWithMeta(c.code, "must be lower or equal", map[string]any{
		"other":  c.other,
		"actual": astValueToMeta(child),
		"value":  astValueToMeta(otherValue),
	})
}
