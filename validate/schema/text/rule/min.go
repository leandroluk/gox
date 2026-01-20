// schema/text/rule/min.go
package rule

import "github.com/leandroluk/go/validate/internal/ruleset"

func Min(code string, min int) ruleset.Rule[string] {
	return newRule(code, "too short", func(actual string) (bool, map[string]any) {
		actualLen := len(actual)
		if actualLen >= min {
			return true, nil
		}
		return false, map[string]any{
			"min":    min,
			"actual": actualLen,
		}
	})
}
