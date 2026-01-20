// schema/object/rule/comparator_field_contains.go
package rule

import (
	"github.com/leandroluk/go/validator/internal/ast"
	"github.com/leandroluk/go/validator/internal/engine"
)

type fieldContainsComparator struct {
	code  string
	other string
}

func FieldContains(code string, other string) Comparator {
	return fieldContainsComparator{code: code, other: other}
}

func (c fieldContainsComparator) Apply(context *engine.Context, root ast.Value, child ast.Value) bool {
	otherValue := ast.Query(root, c.other)

	if child.Kind == ast.KindString && otherValue.Kind == ast.KindString {
		if otherValue.String == "" || (len(otherValue.String) > 0 && contains(child.String, otherValue.String)) {
			return false
		}
	}

	return context.AddIssueWithMeta(c.code, "must contain", map[string]any{
		"other":  c.other,
		"actual": astValueToMeta(child),
		"value":  astValueToMeta(otherValue),
	})
}

func contains(text string, needle string) bool {
	if needle == "" {
		return true
	}
	return len(text) >= len(needle) && (indexOf(text, needle) >= 0)
}

func indexOf(text string, needle string) int {
	// strings.Index, but without importing strings in this file (keeps deps tiny)
	for i := 0; i+len(needle) <= len(text); i++ {
		if text[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}
