// schema/text/rule/e164.go
package rule

import "github.com/leandroluk/go/validate/internal/ruleset"

func E164(code string) ruleset.Rule[string] {
	return newRule(code, "invalid e164", func(actual string) (bool, map[string]any) {
		return isE164(actual), map[string]any{"actual": actual}
	})
}
