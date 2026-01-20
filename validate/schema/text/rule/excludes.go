// schema/text/rule/excludes.go
package rule

import (
	"strings"

	"github.com/leandroluk/gox/validate/internal/ruleset"
)

func Excludes(code string, needle string) ruleset.Rule[string] {
	return newRule(code, "must not contain", func(actual string) (bool, map[string]any) {
		if !strings.Contains(actual, needle) {
			return true, nil
		}
		return false, map[string]any{
			"expected": needle,
			"actual":   actual,
		}
	})
}
