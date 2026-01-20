// schema/text/rule/ne_ignore_case.go
package rule

import (
	"strings"

	"github.com/leandroluk/gox/validate/internal/ruleset"
)

func NeIgnoreCase(code string, disallowed string) ruleset.Rule[string] {
	return newRule(code, "must not be equal (ignore case)", func(actual string) (bool, map[string]any) {
		if !strings.EqualFold(actual, disallowed) {
			return true, nil
		}
		return false, map[string]any{
			"expected": disallowed,
			"actual":   actual,
		}
	})
}
