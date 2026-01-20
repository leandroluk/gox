// schema/text/rule/hexadecimal.go
package rule

import "github.com/leandroluk/go/validator/internal/ruleset"

func Hexadecimal(code string) ruleset.Rule[string] {
	return newRule(code, "invalid hexadecimal", func(actual string) (bool, map[string]any) {
		return isHexadecimal(actual), map[string]any{"actual": actual}
	})
}
