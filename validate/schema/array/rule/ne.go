// schema/array/rule/ne.go
package rule

import (
	"github.com/leandroluk/gox/validate/internal/engine"
	"github.com/leandroluk/gox/validate/internal/ruleset"
)

func Ne(code string, expected int) ruleset.Rule[int] {
	return ruleset.New("", func(actual int, context *engine.Context) (int, bool) {
		if actual != expected {
			return actual, false
		}
		stop := context.AddIssueWithMeta(code, "must not be equal", map[string]any{
			"expected": expected,
			"actual":   actual,
		})
		return actual, stop
	})
}
