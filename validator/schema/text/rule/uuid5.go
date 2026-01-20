// schema/text/rule/uuid5.go
package rule

import "github.com/leandroluk/go/validator/internal/ruleset"

func UUID5(code string) ruleset.Rule[string] {
	return newRule(code, "invalid uuid5", func(actual string) (bool, map[string]any) {
		return uuid5Regex.MatchString(actual), map[string]any{"actual": actual}
	})
}
