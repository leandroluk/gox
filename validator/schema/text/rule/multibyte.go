// schema/text/rule/multibyte.go
package rule

import "github.com/leandroluk/go/validator/internal/ruleset"

func Multibyte(code string) ruleset.Rule[string] {
	return newRule(code, "invalid multibyte", func(actual string) (bool, map[string]any) {
		return isMultibyte(actual), map[string]any{"actual": actual}
	})
}
