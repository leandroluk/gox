// schema/text/rule/url.go
package rule

import "github.com/leandroluk/go/validator/internal/ruleset"

func URL(code string) ruleset.Rule[string] {
	return newRule(code, "invalid url", func(actual string) (bool, map[string]any) {
		return isURL(actual), map[string]any{"actual": actual}
	})
}
