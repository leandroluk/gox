// schema/object/rule/comparator_nefield.go
package rule

import (
	"github.com/leandroluk/gox/validate/internal/ast"
	"github.com/leandroluk/gox/validate/internal/engine"
)

type neFieldComparator struct {
	code  string
	other string
}

func NeField(code string, other string) Comparator {
	return neFieldComparator{code: code, other: other}
}

func (c neFieldComparator) Apply(context *engine.Context, root ast.Value, child ast.Value) bool {
	otherValue := ast.Query(root, c.other)
	if !valuesEqual(child, otherValue) {
		return false
	}

	return context.AddIssueWithMeta(c.code, "must not be equal", map[string]any{
		"other":  c.other,
		"actual": astValueToMeta(child),
		"value":  astValueToMeta(otherValue),
	})
}
