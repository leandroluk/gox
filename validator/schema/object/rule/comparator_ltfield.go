// schema/object/rule/comparator_ltfield.go
package rule

import (
	"github.com/leandroluk/go/validator/internal/ast"
	"github.com/leandroluk/go/validator/internal/engine"
)

type ltFieldComparator struct {
	code  string
	other string
}

func LtField(code string, other string) Comparator {
	return ltFieldComparator{code: code, other: other}
}

func (c ltFieldComparator) Apply(context *engine.Context, root ast.Value, child ast.Value) bool {
	otherValue := ast.Query(root, c.other)

	order, ok := compareOrder(child, otherValue)
	if ok && order < 0 {
		return false
	}

	return context.AddIssueWithMeta(c.code, "must be lower", map[string]any{
		"other":  c.other,
		"actual": astValueToMeta(child),
		"value":  astValueToMeta(otherValue),
	})
}
