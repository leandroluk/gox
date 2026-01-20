// schema/text/rule/fqdn.go
package rule

import "github.com/leandroluk/go/validator/internal/ruleset"

func FQDN(code string) ruleset.Rule[string] {
	return newRule(code, "invalid fqdn", func(actual string) (bool, map[string]any) {
		return isHostname(actual, true), map[string]any{"actual": actual}
	})
}
