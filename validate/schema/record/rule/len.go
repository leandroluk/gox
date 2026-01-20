// schema/record/rule/len.go
package rule

import (
	"github.com/leandroluk/gox/validate/internal/engine"
	"github.com/leandroluk/gox/validate/internal/ruleset"
)

func Len(code string, expected int) ruleset.Rule[int] {
	return ruleset.New(code, func(actual int, context *engine.Context) (int, bool) {
		if actual == expected {
			return actual, false
		}

		stop := context.AddIssueWithMeta(code, "invalid length", map[string]any{
			"expected": expected,
			"actual":   actual,
		})
		return actual, stop
	})
}
