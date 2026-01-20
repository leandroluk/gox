// schema/object/rule/comparator_necsfield.go
package rule

import (
	"github.com/leandroluk/go/validator/internal/ast"
	"github.com/leandroluk/go/validator/internal/engine"
)

type neCSFieldComparator struct {
	code string
	path string
}

func NeCSField(code string, path string) Comparator {
	return neCSFieldComparator{code: code, path: path}
}

func (c neCSFieldComparator) Apply(context *engine.Context, root ast.Value, child ast.Value) bool {
	otherValue := ast.Query(root, c.path)
	if !valuesEqual(child, otherValue) {
		return false
	}

	return context.AddIssueWithMeta(c.code, "must not be equal", map[string]any{
		"path":   c.path,
		"actual": astValueToMeta(child),
		"value":  astValueToMeta(otherValue),
	})
}
