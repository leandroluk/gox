// schema/text/rule/isbn13.go
package rule

import "github.com/leandroluk/go/validator/internal/ruleset"

func ISBN13(code string) ruleset.Rule[string] {
	return newRule(code, "invalid isbn13", func(actual string) (bool, map[string]any) {
		return isISBN13(actual), map[string]any{"actual": actual}
	})
}
