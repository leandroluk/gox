// schema/date/rule/gt.go
package rule

import (
	"time"

	"github.com/leandroluk/gox/validate/internal/engine"
	"github.com/leandroluk/gox/validate/internal/ruleset"
)

func Gt(code string, expected time.Time) ruleset.Rule[time.Time] {
	return ruleset.New("", func(actual time.Time, context *engine.Context) (time.Time, bool) {
		if actual.After(expected) {
			return actual, false
		}
		stop := context.AddIssueWithMeta(code, "must be greater", map[string]any{
			"expected": expected.Format(time.RFC3339Nano),
			"actual":   actual.Format(time.RFC3339Nano),
		})
		return actual, stop
	})
}
