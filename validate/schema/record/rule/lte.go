// schema/record/rule/lte.go
package rule

import (
	"github.com/leandroluk/gox/validate/internal/engine"
	"github.com/leandroluk/gox/validate/internal/ruleset"
)

func Lte(code string, expectedLowerOrEqual int) ruleset.Rule[int] {
	return ruleset.New(code, func(actual int, context *engine.Context) (int, bool) {
		if actual <= expectedLowerOrEqual {
			return actual, false
		}

		stop := context.AddIssueWithMeta(code, "must be lower or equal", map[string]any{
			"expected": expectedLowerOrEqual,
			"actual":   actual,
		})
		return actual, stop
	})
}
