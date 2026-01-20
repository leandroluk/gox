// schema/text/rule/image.go
package rule

import "github.com/leandroluk/go/validator/internal/ruleset"

func Image(code string) ruleset.Rule[string] {
	return newRule(code, "invalid image", func(actual string) (bool, map[string]any) {
		return isImage(actual), map[string]any{"actual": actual}
	})
}
