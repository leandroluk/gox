// schema/array/rule/lt.go
package rule

import (
	"github.com/leandroluk/go/validator/internal/engine"
	"github.com/leandroluk/go/validator/internal/ruleset"
)

func Lt(code string, expected int) ruleset.Rule[int] {
	return ruleset.New("", func(actual int, context *engine.Context) (int, bool) {
		if actual < expected {
			return actual, false
		}
		stop := context.AddIssueWithMeta(code, "must be lower", map[string]any{
			"expected": expected,
			"actual":   actual,
		})
		return actual, stop
	})
}
