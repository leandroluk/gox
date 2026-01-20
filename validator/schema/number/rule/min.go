// schema/number/rule/min.go
package rule

import (
	"github.com/leandroluk/go/validator/internal/engine"
	"github.com/leandroluk/go/validator/internal/ruleset"
	"github.com/leandroluk/go/validator/internal/types"
	"github.com/leandroluk/go/validator/schema/number/util"
)

func Min[N types.Number](code string, minimum N) ruleset.Rule[N] {
	return ruleset.New("min", func(actual N, context *engine.Context) (N, bool) {
		if util.IsNaN(actual) {
			return actual, false
		}

		if actual < minimum {
			stop := context.AddIssueWithMeta(code, "too small", map[string]any{
				"min":    minimum,
				"actual": actual,
			})
			return actual, stop
		}

		return actual, false
	})
}
