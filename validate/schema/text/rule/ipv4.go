// schema/text/rule/ipv4.go
package rule

import "github.com/leandroluk/gox/validate/internal/ruleset"

func IPv4(code string) ruleset.Rule[string] {
	return newRule(code, "invalid ipv4", func(actual string) (bool, map[string]any) {
		return isIPv4(actual), map[string]any{"actual": actual}
	})
}
