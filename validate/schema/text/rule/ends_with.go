// schema/text/rule/ends_with.go
package rule

import (
	"strings"

	"github.com/leandroluk/go/validate/internal/ruleset"
)

func EndsWith(code string, suffix string) ruleset.Rule[string] {
	return newRule(code, "must end with", func(actual string) (bool, map[string]any) {
		if strings.HasSuffix(actual, suffix) {
			return true, nil
		}
		return false, map[string]any{
			"expected": suffix,
			"actual":   actual,
		}
	})
}
