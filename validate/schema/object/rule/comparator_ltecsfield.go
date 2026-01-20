// schema/object/rule/comparator_ltecsfield.go
package rule

import (
	"github.com/leandroluk/gox/validate/internal/ast"
	"github.com/leandroluk/gox/validate/internal/engine"
)

type lteCSFieldComparator struct {
	code string
	path string
}

func LteCSField(code string, path string) Comparator {
	return lteCSFieldComparator{code: code, path: path}
}

func (c lteCSFieldComparator) Apply(context *engine.Context, root ast.Value, child ast.Value) bool {
	otherValue := ast.Query(root, c.path)

	order, ok := compareOrder(child, otherValue)
	if ok && order <= 0 {
		return false
	}

	return context.AddIssueWithMeta(c.code, "must be lower or equal", map[string]any{
		"path":   c.path,
		"actual": astValueToMeta(child),
		"value":  astValueToMeta(otherValue),
	})
}
