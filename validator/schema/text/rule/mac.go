// schema/text/rule/mac.go
package rule

import "github.com/leandroluk/go/validator/internal/ruleset"

func MAC(code string) ruleset.Rule[string] {
	return newRule(code, "invalid mac", func(actual string) (bool, map[string]any) {
		return isMAC(actual), map[string]any{"actual": actual}
	})
}
