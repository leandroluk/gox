// schema/text/rule/rgb.go
package rule

import "github.com/leandroluk/gox/validate/internal/ruleset"

func RGB(code string) ruleset.Rule[string] {
	return newRule(code, "invalid rgb", func(actual string) (bool, map[string]any) {
		return isRGB(actual), map[string]any{"actual": actual}
	})
}
