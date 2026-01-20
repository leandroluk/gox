// schema/text/rule/hsla.go
package rule

import "github.com/leandroluk/go/validate/internal/ruleset"

func HSLA(code string) ruleset.Rule[string] {
	return newRule(code, "invalid hsla", func(actual string) (bool, map[string]any) {
		return isHSLA(actual), map[string]any{"actual": actual}
	})
}
