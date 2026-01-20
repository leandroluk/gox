// schema/text/rule/cidr.go
package rule

import "github.com/leandroluk/go/validate/internal/ruleset"

func CIDR(code string) ruleset.Rule[string] {
	return newRule(code, "invalid cidr", func(actual string) (bool, map[string]any) {
		return isCIDR(actual), map[string]any{"actual": actual}
	})
}
