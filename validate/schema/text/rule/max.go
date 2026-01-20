// schema/text/rule/max.go
package rule

import "github.com/leandroluk/gox/validate/internal/ruleset"

func Max(code string, max int) ruleset.Rule[string] {
	return newRule(code, "too long", func(actual string) (bool, map[string]any) {
		actualLen := len(actual)
		if actualLen <= max {
			return true, nil
		}
		return false, map[string]any{
			"max":    max,
			"actual": actualLen,
		}
	})
}
