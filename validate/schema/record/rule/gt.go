// schema/record/rule/gt.go
package rule

import (
	"github.com/leandroluk/go/validate/internal/engine"
	"github.com/leandroluk/go/validate/internal/ruleset"
)

func Gt(code string, expectedGreaterThan int) ruleset.Rule[int] {
	return ruleset.New(code, func(actual int, context *engine.Context) (int, bool) {
		if actual > expectedGreaterThan {
			return actual, false
		}

		stop := context.AddIssueWithMeta(code, "must be greater", map[string]any{
			"expected": expectedGreaterThan,
			"actual":   actual,
		})
		return actual, stop
	})
}
