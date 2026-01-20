// schema/text/rule/uuid3.go
package rule

import "github.com/leandroluk/gox/validate/internal/ruleset"

func UUID3(code string) ruleset.Rule[string] {
	return newRule(code, "invalid uuid3", func(actual string) (bool, map[string]any) {
		return uuid3Regex.MatchString(actual), map[string]any{"actual": actual}
	})
}
