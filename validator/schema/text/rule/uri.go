// schema/text/rule/uri.go
package rule

import "github.com/leandroluk/go/validator/internal/ruleset"

func URI(code string) ruleset.Rule[string] {
	return newRule(code, "invalid uri", func(actual string) (bool, map[string]any) {
		return isURI(actual), map[string]any{"actual": actual}
	})
}
