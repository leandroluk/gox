// schema/text/rule/uppercase.go
package rule

import (
	"strings"

	"github.com/leandroluk/go/validator/internal/ruleset"
)

func Uppercase(code string) ruleset.Rule[string] {
	return newRule(code, "must be uppercase", func(actual string) (bool, map[string]any) {
		if actual == strings.ToUpper(actual) {
			return true, nil
		}
		return false, map[string]any{"actual": actual}
	})
}
