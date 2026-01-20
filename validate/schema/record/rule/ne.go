// schema/record/rule/ne.go
package rule

import (
	"github.com/leandroluk/gox/validate/internal/engine"
	"github.com/leandroluk/gox/validate/internal/ruleset"
)

func Ne(code string, notExpected int) ruleset.Rule[int] {
	return ruleset.New(code, func(actual int, context *engine.Context) (int, bool) {
		if actual != notExpected {
			return actual, false
		}

		stop := context.AddIssueWithMeta(code, "must not be equal", map[string]any{
			"expected": notExpected,
			"actual":   actual,
		})
		return actual, stop
	})
}
