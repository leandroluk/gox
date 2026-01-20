// schema/object/rule/comparator_gtfield.go
package rule

import (
	"github.com/leandroluk/go/validate/internal/ast"
	"github.com/leandroluk/go/validate/internal/engine"
)

type gtFieldComparator struct {
	code  string
	other string
}

func GtField(code string, other string) Comparator {
	return gtFieldComparator{code: code, other: other}
}

func (c gtFieldComparator) Apply(context *engine.Context, root ast.Value, child ast.Value) bool {
	otherValue := ast.Query(root, c.other)

	order, ok := compareOrder(child, otherValue)
	if ok && order > 0 {
		return false
	}

	return context.AddIssueWithMeta(c.code, "must be greater", map[string]any{
		"other":  c.other,
		"actual": astValueToMeta(child),
		"value":  astValueToMeta(otherValue),
	})
}
