// schema/object/rule/condition_required_with.go
package rule

import (
	"github.com/leandroluk/gox/validate/internal/ast"
	"github.com/leandroluk/gox/validate/internal/engine"
)

type requiredWithCondition struct {
	code  string
	paths []string
}

func RequiredWith(code string, paths ...string) RequiredCondition {
	copied := append([]string(nil), paths...)
	return requiredWithCondition{
		code:  code,
		paths: copied,
	}
}

func (c requiredWithCondition) Apply(context *engine.Context, root ast.Value, _ ast.Value, childPresent bool) bool {
	if childPresent {
		return false
	}

	required := false
	for _, path := range c.paths {
		actual := ast.Query(root, path)
		if !actual.IsMissing() && !actual.IsNull() {
			required = true
			break
		}
	}

	if !required {
		return false
	}

	return context.AddIssueWithMeta(c.code, "required", map[string]any{
		"paths": append([]string(nil), c.paths...),
	})
}
