// schema/date/rule/max.go
package rule

import (
	"time"

	"github.com/leandroluk/go/validate/internal/engine"
	"github.com/leandroluk/go/validate/internal/ruleset"
)

func Max(code string, maximum time.Time) ruleset.Rule[time.Time] {
	return ruleset.New("max", func(actual time.Time, context *engine.Context) (time.Time, bool) {
		if actual.After(maximum) {
			stop := context.AddIssueWithMeta(code, "too late", map[string]any{
				"max":    maximum.Format(time.RFC3339Nano),
				"actual": actual.Format(time.RFC3339Nano),
			})
			return actual, stop
		}
		return actual, false
	})
}
