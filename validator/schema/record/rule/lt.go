// schema/record/rule/lt.go
package rule

import (
	"github.com/leandroluk/go/validator/internal/engine"
	"github.com/leandroluk/go/validator/internal/ruleset"
)

func Lt(code string, expectedLowerThan int) ruleset.Rule[int] {
	return ruleset.New(code, func(actual int, context *engine.Context) (int, bool) {
		if actual < expectedLowerThan {
			return actual, false
		}

		stop := context.AddIssueWithMeta(code, "must be lower", map[string]any{
			"expected": expectedLowerThan,
			"actual":   actual,
		})
		return actual, stop
	})
}
