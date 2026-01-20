// schema/text/rule/starts_with.go
package rule

import (
	"strings"

	"github.com/leandroluk/go/validate/internal/ruleset"
)

func StartsWith(code string, prefix string) ruleset.Rule[string] {
	return newRule(code, "must start with", func(actual string) (bool, map[string]any) {
		if strings.HasPrefix(actual, prefix) {
			return true, nil
		}
		return false, map[string]any{
			"expected": prefix,
			"actual":   actual,
		}
	})
}
