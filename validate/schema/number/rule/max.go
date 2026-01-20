// schema/number/rule/max.go
package rule

import (
	"github.com/leandroluk/gox/validate/internal/engine"
	"github.com/leandroluk/gox/validate/internal/ruleset"
	"github.com/leandroluk/gox/validate/internal/types"
	"github.com/leandroluk/gox/validate/schema/number/util"
)

func Max[N types.Number](code string, maximum N) ruleset.Rule[N] {
	return ruleset.New("max", func(actual N, context *engine.Context) (N, bool) {
		if util.IsNaN(actual) {
			return actual, false
		}

		if actual > maximum {
			stop := context.AddIssueWithMeta(code, "too large", map[string]any{
				"max":    maximum,
				"actual": actual,
			})
			return actual, stop
		}

		return actual, false
	})
}
