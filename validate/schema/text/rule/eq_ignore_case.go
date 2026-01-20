// schema/text/rule/eq_ignore_case.go
package rule

import (
	"strings"

	"github.com/leandroluk/go/validate/internal/ruleset"
)

func EqIgnoreCase(code string, expected string) ruleset.Rule[string] {
	return newRule(code, "must be equal (ignore case)", func(actual string) (bool, map[string]any) {
		if strings.EqualFold(actual, expected) {
			return true, nil
		}
		return false, map[string]any{
			"expected": expected,
			"actual":   actual,
		}
	})
}
