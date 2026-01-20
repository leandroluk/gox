// schema/record/rule/gte.go
package rule

import (
	"github.com/leandroluk/go/validator/internal/engine"
	"github.com/leandroluk/go/validator/internal/ruleset"
)

func Gte(code string, expectedGreaterOrEqual int) ruleset.Rule[int] {
	return ruleset.New(code, func(actual int, context *engine.Context) (int, bool) {
		if actual >= expectedGreaterOrEqual {
			return actual, false
		}

		stop := context.AddIssueWithMeta(code, "must be greater or equal", map[string]any{
			"expected": expectedGreaterOrEqual,
			"actual":   actual,
		})
		return actual, stop
	})
}
