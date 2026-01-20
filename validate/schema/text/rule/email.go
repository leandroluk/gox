// schema/text/rule/email.go
package rule

import "github.com/leandroluk/gox/validate/internal/ruleset"

func Email(code string) ruleset.Rule[string] {
	return newRule(code, "invalid email", func(actual string) (bool, map[string]any) {
		return isEmail(actual), map[string]any{"actual": actual}
	})
}
