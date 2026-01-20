// schema/date/rule/lt.go
package rule

import (
	"time"

	"github.com/leandroluk/go/validator/internal/engine"
	"github.com/leandroluk/go/validator/internal/ruleset"
)

func Lt(code string, expected time.Time) ruleset.Rule[time.Time] {
	return ruleset.New("", func(actual time.Time, context *engine.Context) (time.Time, bool) {
		if actual.Before(expected) {
			return actual, false
		}
		stop := context.AddIssueWithMeta(code, "must be lower", map[string]any{
			"expected": expected.Format(time.RFC3339Nano),
			"actual":   actual.Format(time.RFC3339Nano),
		})
		return actual, stop
	})
}
