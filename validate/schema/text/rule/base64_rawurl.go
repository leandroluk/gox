// schema/text/rule/base64_rawurl.go
package rule

import "github.com/leandroluk/gox/validate/internal/ruleset"

func Base64RawURL(code string) ruleset.Rule[string] {
	return newRule(code, "invalid base64rawurl", func(actual string) (bool, map[string]any) {
		return isBase64RawURL(actual), map[string]any{"actual": actual}
	})
}
