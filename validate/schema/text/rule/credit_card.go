// schema/text/rule/credit_card.go
package rule

import "github.com/leandroluk/go/validate/internal/ruleset"

func CreditCard(code string) ruleset.Rule[string] {
	return newRule(code, "invalid credit card", func(actual string) (bool, map[string]any) {
		return isCreditCard(actual), map[string]any{"actual": actual}
	})
}
