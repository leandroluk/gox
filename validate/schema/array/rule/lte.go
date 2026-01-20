// schema/array/rule/lte.go
package rule

import (
	"github.com/leandroluk/go/validate/internal/engine"
	"github.com/leandroluk/go/validate/internal/ruleset"
)

func Lte(code string, expected int) ruleset.Rule[int] {
	return ruleset.New("", func(actual int, context *engine.Context) (int, bool) {
		if actual <= expected {
			return actual, false
		}
		stop := context.AddIssueWithMeta(code, "must be lower or equal", map[string]any{
			"expected": expected,
			"actual":   actual,
		})
		return actual, stop
	})
}
