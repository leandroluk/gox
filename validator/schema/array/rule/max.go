// schema/array/rule/max.go
package rule

import (
	"github.com/leandroluk/go/validator/internal/engine"
	"github.com/leandroluk/go/validator/internal/ruleset"
)

func Max(code string, maximum int) ruleset.Rule[int] {
	return ruleset.New("max", func(actual int, context *engine.Context) (int, bool) {
		if actual > maximum {
			stop := context.AddIssueWithMeta(code, "too long", map[string]any{
				"max":    maximum,
				"actual": actual,
			})
			return actual, stop
		}
		return actual, false
	})
}
