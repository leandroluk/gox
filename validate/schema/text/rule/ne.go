// schema/text/rule/ne.go
package rule

import "github.com/leandroluk/gox/validate/internal/ruleset"

func Ne(code string, disallowed string) ruleset.Rule[string] {
	return newRule(code, "must not be equal", func(actual string) (bool, map[string]any) {
		if actual != disallowed {
			return true, nil
		}
		return false, map[string]any{
			"expected": disallowed,
			"actual":   actual,
		}
	})
}
