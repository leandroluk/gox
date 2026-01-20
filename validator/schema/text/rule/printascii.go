// schema/text/rule/printascii.go
package rule

import "github.com/leandroluk/go/validator/internal/ruleset"

func PrintASCII(code string) ruleset.Rule[string] {
	return newRule(code, "invalid printascii", func(actual string) (bool, map[string]any) {
		return isPrintASCII(actual), map[string]any{"actual": actual}
	})
}
