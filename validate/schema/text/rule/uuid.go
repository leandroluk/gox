// schema/text/rule/uuid.go
package rule

import "github.com/leandroluk/go/validate/internal/ruleset"

func UUID(code string) ruleset.Rule[string] {
	return newRule(code, "invalid uuid", func(actual string) (bool, map[string]any) {
		return uuidRegex.MatchString(actual), map[string]any{"actual": actual}
	})
}
