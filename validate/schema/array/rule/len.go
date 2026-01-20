// schema/array/rule/len.go
package rule

import (
	"github.com/leandroluk/go/validate/internal/engine"
	"github.com/leandroluk/go/validate/internal/ruleset"
)

func Len(code string, expected int) ruleset.Rule[int] {
	return ruleset.New("", func(actual int, context *engine.Context) (int, bool) {
		if actual == expected {
			return actual, false
		}
		stop := context.AddIssueWithMeta(code, "invalid length", map[string]any{
			"expected": expected,
			"actual":   actual,
		})
		return actual, stop
	})
}
