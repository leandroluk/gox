// schema/text/rule/numeric.go
package rule

import "github.com/leandroluk/go/validator/internal/ruleset"

func Numeric(code string) ruleset.Rule[string] {
	return newRule(code, "invalid numeric", func(actual string) (bool, map[string]any) {
		return isNumeric(actual), map[string]any{"actual": actual}
	})
}
