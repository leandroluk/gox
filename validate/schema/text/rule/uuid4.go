// schema/text/rule/uuid4.go
package rule

import "github.com/leandroluk/go/validate/internal/ruleset"

func UUID4(code string) ruleset.Rule[string] {
	return newRule(code, "invalid uuid4", func(actual string) (bool, map[string]any) {
		return uuid4Regex.MatchString(actual), map[string]any{"actual": actual}
	})
}
