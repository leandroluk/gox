// schema/text/rule/urn_rfc2141.go
package rule

import "github.com/leandroluk/go/validator/internal/ruleset"

func URNRFC2141(code string) ruleset.Rule[string] {
	return newRule(code, "invalid urn", func(actual string) (bool, map[string]any) {
		return isURNRFC2141(actual), map[string]any{"actual": actual}
	})
}
