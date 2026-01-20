// schema/text/rule/issn.go
package rule

import "github.com/leandroluk/gox/validate/internal/ruleset"

func ISSN(code string) ruleset.Rule[string] {
	return newRule(code, "invalid issn", func(actual string) (bool, map[string]any) {
		return isISSN(actual), map[string]any{"actual": actual}
	})
}
