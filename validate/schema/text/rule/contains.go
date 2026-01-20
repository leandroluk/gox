// schema/text/rule/contains.go
package rule

import (
	"strings"

	"github.com/leandroluk/gox/validate/internal/ruleset"
)

func Contains(code string, needle string) ruleset.Rule[string] {
	return newRule(code, "must contain", func(actual string) (bool, map[string]any) {
		if strings.Contains(actual, needle) {
			return true, nil
		}
		return false, map[string]any{
			"expected": needle,
			"actual":   actual,
		}
	})
}
