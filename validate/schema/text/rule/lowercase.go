// schema/text/rule/lowercase.go
package rule

import (
	"strings"

	"github.com/leandroluk/gox/validate/internal/ruleset"
)

func Lowercase(code string) ruleset.Rule[string] {
	return newRule(code, "must be lowercase", func(actual string) (bool, map[string]any) {
		if actual == strings.ToLower(actual) {
			return true, nil
		}
		return false, map[string]any{"actual": actual}
	})
}
