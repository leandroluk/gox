// schema/number/rule/gte.go
package rule

import (
	"github.com/leandroluk/gox/validate/internal/engine"
	"github.com/leandroluk/gox/validate/internal/ruleset"
	"github.com/leandroluk/gox/validate/internal/types"
	"github.com/leandroluk/gox/validate/schema/number/util"
)

func Gte[N types.Number](code string, expected N) ruleset.Rule[N] {
	return ruleset.New("", func(actual N, context *engine.Context) (N, bool) {
		if util.IsNaN(actual) || util.IsNaN(expected) {
			stop := context.AddIssueWithMeta(code, "incomparable", map[string]any{
				"expected": expected,
				"actual":   actual,
			})
			return actual, stop
		}

		if actual >= expected {
			return actual, false
		}

		stop := context.AddIssueWithMeta(code, "must be greater or equal", map[string]any{
			"expected": expected,
			"actual":   actual,
		})
		return actual, stop
	})
}
