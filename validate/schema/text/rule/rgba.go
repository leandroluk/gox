// schema/text/rule/rgba.go
package rule

import "github.com/leandroluk/gox/validate/internal/ruleset"

func RGBA(code string) ruleset.Rule[string] {
	return newRule(code, "invalid rgba", func(actual string) (bool, map[string]any) {
		return isRGBA(actual), map[string]any{"actual": actual}
	})
}
