// schema/date/rule/eq.go
package rule

import (
	"time"

	"github.com/leandroluk/go/validator/internal/engine"
	"github.com/leandroluk/go/validator/internal/ruleset"
)

func Eq(code string, expected time.Time) ruleset.Rule[time.Time] {
	return ruleset.New("", func(actual time.Time, context *engine.Context) (time.Time, bool) {
		if actual.Equal(expected) {
			return actual, false
		}
		stop := context.AddIssueWithMeta(code, "must be equal", map[string]any{
			"expected": expected.Format(time.RFC3339Nano),
			"actual":   actual.Format(time.RFC3339Nano),
		})
		return actual, stop
	})
}
