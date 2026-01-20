// schema/object/rule/comparator_field_excludes.go
package rule

import (
	"github.com/leandroluk/go/validate/internal/ast"
	"github.com/leandroluk/go/validate/internal/engine"
)

type fieldExcludesComparator struct {
	code  string
	other string
}

func FieldExcludes(code string, other string) Comparator {
	return fieldExcludesComparator{code: code, other: other}
}

func (c fieldExcludesComparator) Apply(context *engine.Context, root ast.Value, child ast.Value) bool {
	otherValue := ast.Query(root, c.other)

	if child.Kind == ast.KindString && otherValue.Kind == ast.KindString {
		if otherValue.String == "" || indexOf(child.String, otherValue.String) < 0 {
			return false
		}
	}

	return context.AddIssueWithMeta(c.code, "must not contain", map[string]any{
		"other":  c.other,
		"actual": astValueToMeta(child),
		"value":  astValueToMeta(otherValue),
	})
}
