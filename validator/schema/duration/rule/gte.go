// schema/duration/rule/gte.go
package rule

import (
	"time"

	"github.com/leandroluk/go/validator/internal/engine"
	"github.com/leandroluk/go/validator/internal/ruleset"
)

func Gte(code string, expected time.Duration) ruleset.Rule[time.Duration] {
	return ruleset.New("", func(actual time.Duration, context *engine.Context) (time.Duration, bool) {
		if actual >= expected {
			return actual, false
		}
		stop := context.AddIssueWithMeta(code, "must be greater or equal", map[string]any{
			"expected": expected.String(),
			"actual":   actual.String(),
		})
		return actual, stop
	})
}
