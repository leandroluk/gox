// schema/record/rule/max.go
package rule

import (
	"github.com/leandroluk/go/validator/internal/engine"
	"github.com/leandroluk/go/validator/internal/ruleset"
)

func Max(code string, max int) ruleset.Rule[int] {
	return ruleset.New(code, func(actual int, context *engine.Context) (int, bool) {
		if actual <= max {
			return actual, false
		}

		stop := context.AddIssueWithMeta(code, "too large", map[string]any{
			"max":    max,
			"actual": actual,
		})
		return actual, stop
	})
}
