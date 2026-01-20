// schema/object/rule/comparator_gtefield.go
package rule

import (
	"github.com/leandroluk/go/validator/internal/ast"
	"github.com/leandroluk/go/validator/internal/engine"
)

type gteFieldComparator struct {
	code  string
	other string
}

func GteField(code string, other string) Comparator {
	return gteFieldComparator{code: code, other: other}
}

func (c gteFieldComparator) Apply(context *engine.Context, root ast.Value, child ast.Value) bool {
	otherValue := ast.Query(root, c.other)

	order, ok := compareOrder(child, otherValue)
	if ok && order >= 0 {
		return false
	}

	return context.AddIssueWithMeta(c.code, "must be greater or equal", map[string]any{
		"other":  c.other,
		"actual": astValueToMeta(child),
		"value":  astValueToMeta(otherValue),
	})
}
