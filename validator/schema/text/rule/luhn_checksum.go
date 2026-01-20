// schema/text/rule/luhn_checksum.go
package rule

import "github.com/leandroluk/go/validator/internal/ruleset"

func LuhnChecksum(code string) ruleset.Rule[string] {
	return newRule(code, "invalid luhn checksum", func(actual string) (bool, map[string]any) {
		return isLuhnChecksum(actual), map[string]any{"actual": actual}
	})
}
