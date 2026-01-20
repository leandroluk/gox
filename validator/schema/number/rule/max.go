// schema/number/rule/max.go
package rule

import (
	"github.com/leandroluk/go/validator/internal/engine"
	"github.com/leandroluk/go/validator/internal/ruleset"
	"github.com/leandroluk/go/validator/internal/types"
	"github.com/leandroluk/go/validator/schema/number/util"
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
