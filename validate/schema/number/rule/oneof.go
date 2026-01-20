// schema/number/rule/oneof.go
package rule

import (
	"github.com/leandroluk/gox/validate/internal/engine"
	"github.com/leandroluk/gox/validate/internal/ruleset"
	"github.com/leandroluk/gox/validate/internal/types"
	"github.com/leandroluk/gox/validate/schema/number/util"
)

func OneOf[N types.Number](code string, values ...N) ruleset.Rule[N] {
	allowedMeta := make([]any, 0, len(values))
	allowedMap := make(map[N]struct{}, len(values))
	allowNaN := false

	for _, value := range values {
		allowedMeta = append(allowedMeta, value)

		if util.IsNaN(value) {
			allowNaN = true
			continue
		}
		allowedMap[value] = struct{}{}
	}

	return ruleset.New("oneof", func(actual N, context *engine.Context) (N, bool) {
		allowed := false

		if util.IsNaN(actual) {
			allowed = allowNaN
		} else {
			_, allowed = allowedMap[actual]
		}

		if allowed {
			return actual, false
		}

		stop := context.AddIssueWithMeta(code, "not allowed", map[string]any{
			"allowed": allowedMeta,
			"actual":  actual,
		})
		return actual, stop
	})
}
