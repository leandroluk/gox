// schema/text/rule/cve.go
package rule

import "github.com/leandroluk/go/validator/internal/ruleset"

func CVE(code string) ruleset.Rule[string] {
	return newRule(code, "invalid cve", func(actual string) (bool, map[string]any) {
		return isCVE(actual), map[string]any{"actual": actual}
	})
}
