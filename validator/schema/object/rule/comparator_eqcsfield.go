// schema/object/rule/comparator_eqcsfield.go
package rule

import (
	"github.com/leandroluk/go/validator/internal/ast"
	"github.com/leandroluk/go/validator/internal/engine"
)

type eqCSFieldComparator struct {
	code string
	path string
}

func EqCSField(code string, path string) Comparator {
	return eqCSFieldComparator{code: code, path: path}
}

func (c eqCSFieldComparator) Apply(context *engine.Context, root ast.Value, child ast.Value) bool {
	otherValue := ast.Query(root, c.path)
	if valuesEqual(child, otherValue) {
		return false
	}

	return context.AddIssueWithMeta(c.code, "must be equal", map[string]any{
		"path":   c.path,
		"actual": astValueToMeta(child),
		"value":  astValueToMeta(otherValue),
	})
}
