// schema/duration/rule/max.go
package rule

import (
	"time"

	"github.com/leandroluk/go/validator/internal/engine"
	"github.com/leandroluk/go/validator/internal/ruleset"
)

func Max(code string, maximum time.Duration) ruleset.Rule[time.Duration] {
	return ruleset.New("max", func(actual time.Duration, context *engine.Context) (time.Duration, bool) {
		if actual > maximum {
			stop := context.AddIssueWithMeta(code, "too large", map[string]any{
				"max":    maximum.String(),
				"actual": actual.String(),
			})
			return actual, stop
		}
		return actual, false
	})
}
