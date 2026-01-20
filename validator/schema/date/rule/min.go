// schema/date/rule/min.go
package rule

import (
	"time"

	"github.com/leandroluk/go/validator/internal/engine"
	"github.com/leandroluk/go/validator/internal/ruleset"
)

func Min(code string, minimum time.Time) ruleset.Rule[time.Time] {
	return ruleset.New("min", func(actual time.Time, context *engine.Context) (time.Time, bool) {
		if actual.Before(minimum) {
			stop := context.AddIssueWithMeta(code, "too early", map[string]any{
				"min":    minimum.Format(time.RFC3339Nano),
				"actual": actual.Format(time.RFC3339Nano),
			})
			return actual, stop
		}
		return actual, false
	})
}
