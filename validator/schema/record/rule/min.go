// schema/record/rule/min.go
package rule

import (
	"github.com/leandroluk/go/validator/internal/engine"
	"github.com/leandroluk/go/validator/internal/ruleset"
)

func Min(code string, min int) ruleset.Rule[int] {
	return ruleset.New(code, func(actual int, context *engine.Context) (int, bool) {
		if actual >= min {
			return actual, false
		}

		stop := context.AddIssueWithMeta(code, "too small", map[string]any{
			"min":    min,
			"actual": actual,
		})
		return actual, stop
	})
}
