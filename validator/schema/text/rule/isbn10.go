// schema/text/rule/isbn10.go
package rule

import "github.com/leandroluk/go/validator/internal/ruleset"

func ISBN10(code string) ruleset.Rule[string] {
	return newRule(code, "invalid isbn10", func(actual string) (bool, map[string]any) {
		return isISBN10(actual), map[string]any{"actual": actual}
	})
}
