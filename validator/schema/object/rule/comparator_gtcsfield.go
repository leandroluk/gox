// schema/object/rule/comparator_gtcsfield.go
package rule

import (
	"github.com/leandroluk/go/validator/internal/ast"
	"github.com/leandroluk/go/validator/internal/engine"
)

type gtCSFieldComparator struct {
	code string
	path string
}

func GtCSField(code string, path string) Comparator {
	return gtCSFieldComparator{code: code, path: path}
}

func (c gtCSFieldComparator) Apply(context *engine.Context, root ast.Value, child ast.Value) bool {
	otherValue := ast.Query(root, c.path)

	order, ok := compareOrder(child, otherValue)
	if ok && order > 0 {
		return false
	}

	return context.AddIssueWithMeta(c.code, "must be greater", map[string]any{
		"path":   c.path,
		"actual": astValueToMeta(child),
		"value":  astValueToMeta(otherValue),
	})
}
