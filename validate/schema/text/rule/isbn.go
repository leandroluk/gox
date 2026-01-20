// schema/text/rule/isbn.go
package rule

import "github.com/leandroluk/go/validate/internal/ruleset"

func ISBN(code string) ruleset.Rule[string] {
	return newRule(code, "invalid isbn", func(actual string) (bool, map[string]any) {
		return isISBN(actual), map[string]any{"actual": actual}
	})
}
