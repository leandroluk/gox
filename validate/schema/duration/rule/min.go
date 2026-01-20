// schema/duration/rule/min.go
package rule

import (
	"time"

	"github.com/leandroluk/gox/validate/internal/engine"
	"github.com/leandroluk/gox/validate/internal/ruleset"
)

func Min(code string, minimum time.Duration) ruleset.Rule[time.Duration] {
	return ruleset.New("min", func(actual time.Duration, context *engine.Context) (time.Duration, bool) {
		if actual < minimum {
			stop := context.AddIssueWithMeta(code, "too small", map[string]any{
				"min":    minimum.String(),
				"actual": actual.String(),
			})
			return actual, stop
		}
		return actual, false
	})
}
