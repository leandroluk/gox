// schema/text/rule/ascii.go
package rule

import "github.com/leandroluk/go/validator/internal/ruleset"

func ASCII(code string) ruleset.Rule[string] {
	return newRule(code, "invalid ascii", func(actual string) (bool, map[string]any) {
		return isASCII(actual), map[string]any{"actual": actual}
	})
}
