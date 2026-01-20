// schema/text/rule/ipv6.go
package rule

import "github.com/leandroluk/gox/validate/internal/ruleset"

func IPv6(code string) ruleset.Rule[string] {
	return newRule(code, "invalid ipv6", func(actual string) (bool, map[string]any) {
		return isIPv6(actual), map[string]any{"actual": actual}
	})
}
