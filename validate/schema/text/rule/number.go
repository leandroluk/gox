// schema/text/rule/number.go
package rule

import "github.com/leandroluk/gox/validate/internal/ruleset"

func Number(code string) ruleset.Rule[string] {
	return newRule(code, "invalid number", func(actual string) (bool, map[string]any) {
		return isNumber(actual), map[string]any{"actual": actual}
	})
}
