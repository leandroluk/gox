// schema/record/rule/eq.go
package rule

import (
	"github.com/leandroluk/go/validator/internal/engine"
	"github.com/leandroluk/go/validator/internal/ruleset"
)

func Eq(code string, expected int) ruleset.Rule[int] {
	return ruleset.New(code, func(actual int, context *engine.Context) (int, bool) {
		if actual == expected {
			return actual, false
		}

		stop := context.AddIssueWithMeta(code, "must be equal", map[string]any{
			"expected": expected,
			"actual":   actual,
		})
		return actual, stop
	})
}
