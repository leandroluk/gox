// schema/text/rule/port.go
package rule

import "github.com/leandroluk/gox/validate/internal/ruleset"

func Port(code string) ruleset.Rule[string] {
	return newRule(code, "invalid port", func(actual string) (bool, map[string]any) {
		return isPort(actual), map[string]any{"actual": actual}
	})
}
