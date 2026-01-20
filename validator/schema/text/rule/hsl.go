// schema/text/rule/hsl.go
package rule

import "github.com/leandroluk/go/validator/internal/ruleset"

func HSL(code string) ruleset.Rule[string] {
	return newRule(code, "invalid hsl", func(actual string) (bool, map[string]any) {
		return isHSL(actual), map[string]any{"actual": actual}
	})
}
