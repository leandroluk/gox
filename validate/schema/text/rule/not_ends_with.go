// schema/text/rule/not_ends_with.go
package rule

import (
	"strings"

	"github.com/leandroluk/gox/validate/internal/ruleset"
)

func NotEndsWith(code string, suffix string) ruleset.Rule[string] {
	return newRule(code, "must not end with", func(actual string) (bool, map[string]any) {
		if !strings.HasSuffix(actual, suffix) {
			return true, nil
		}
		return false, map[string]any{
			"expected": suffix,
			"actual":   actual,
		}
	})
}
