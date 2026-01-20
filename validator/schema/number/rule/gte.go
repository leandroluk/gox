// schema/number/rule/gte.go
package rule

import (
	"github.com/leandroluk/go/validator/internal/engine"
	"github.com/leandroluk/go/validator/internal/ruleset"
	"github.com/leandroluk/go/validator/internal/types"
	"github.com/leandroluk/go/validator/schema/number/util"
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
