// schema/text/rule/eq.go
package rule

import "github.com/leandroluk/gox/validate/internal/ruleset"

func Eq(code string, expected string) ruleset.Rule[string] {
	return newRule(code, "must be equal", func(actual string) (bool, map[string]any) {
		if actual == expected {
			return true, nil
		}
		return false, map[string]any{
			"expected": expected,
			"actual":   actual,
		}
	})
}
