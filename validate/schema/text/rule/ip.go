// schema/text/rule/ip.go
package rule

import "github.com/leandroluk/go/validate/internal/ruleset"

func IP(code string) ruleset.Rule[string] {
	return newRule(code, "invalid ip", func(actual string) (bool, map[string]any) {
		return isIP(actual), map[string]any{"actual": actual}
	})
}
