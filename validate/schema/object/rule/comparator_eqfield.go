// schema/object/rule/comparator_eqfield.go
package rule

import (
	"github.com/leandroluk/gox/validate/internal/ast"
	"github.com/leandroluk/gox/validate/internal/engine"
)

type eqFieldComparator struct {
	code  string
	other string
}

func EqField(code string, other string) Comparator {
	return eqFieldComparator{code: code, other: other}
}

func (c eqFieldComparator) Apply(context *engine.Context, root ast.Value, child ast.Value) bool {
	otherValue := ast.Query(root, c.other)
	if valuesEqual(child, otherValue) {
		return false
	}

	return context.AddIssueWithMeta(c.code, "must be equal", map[string]any{
		"other":  c.other,
		"actual": astValueToMeta(child),
		"value":  astValueToMeta(otherValue),
	})
}
