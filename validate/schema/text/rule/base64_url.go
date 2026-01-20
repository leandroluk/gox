// schema/text/rule/base64_url.go
package rule

import "github.com/leandroluk/gox/validate/internal/ruleset"

func Base64URL(code string) ruleset.Rule[string] {
	return newRule(code, "invalid base64url", func(actual string) (bool, map[string]any) {
		return isBase64URL(actual), map[string]any{"actual": actual}
	})
}
