// schema/object/rule/comparator_ltcsfield.go
package rule

import (
	"github.com/leandroluk/go/validator/internal/ast"
	"github.com/leandroluk/go/validator/internal/engine"
)

type ltCSFieldComparator struct {
	code string
	path string
}

func LtCSField(code string, path string) Comparator {
	return ltCSFieldComparator{code: code, path: path}
}

func (c ltCSFieldComparator) Apply(context *engine.Context, root ast.Value, child ast.Value) bool {
	otherValue := ast.Query(root, c.path)

	order, ok := compareOrder(child, otherValue)
	if ok && order < 0 {
		return false
	}

	return context.AddIssueWithMeta(c.code, "must be lower", map[string]any{
		"path":   c.path,
		"actual": astValueToMeta(child),
		"value":  astValueToMeta(otherValue),
	})
}
