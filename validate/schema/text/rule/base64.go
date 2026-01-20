// schema/text/rule/base64.go
package rule

import "github.com/leandroluk/go/validate/internal/ruleset"

func Base64(code string) ruleset.Rule[string] {
	return newRule(code, "invalid base64", func(actual string) (bool, map[string]any) {
		return isBase64(actual), map[string]any{"actual": actual}
	})
}
