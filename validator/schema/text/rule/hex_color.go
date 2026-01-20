// schema/text/rule/hex_color.go
package rule

import "github.com/leandroluk/go/validator/internal/ruleset"

func HexColor(code string) ruleset.Rule[string] {
	return newRule(code, "invalid hex color", func(actual string) (bool, map[string]any) {
		return isHexColor(actual), map[string]any{"actual": actual}
	})
}
