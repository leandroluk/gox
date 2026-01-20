// schema/text/rule/hostname.go
package rule

import "github.com/leandroluk/go/validator/internal/ruleset"

func Hostname(code string) ruleset.Rule[string] {
	return newRule(code, "invalid hostname", func(actual string) (bool, map[string]any) {
		return isHostname(actual, false), map[string]any{"actual": actual}
	})
}
