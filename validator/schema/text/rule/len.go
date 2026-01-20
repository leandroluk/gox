// schema/text/rule/len.go
package rule

import "github.com/leandroluk/go/validator/internal/ruleset"

func Len(code string, expected int) ruleset.Rule[string] {
	return newRule(code, "invalid length", func(actual string) (bool, map[string]any) {
		actualLen := len(actual)
		if actualLen == expected {
			return true, nil
		}
		return false, map[string]any{
			"expected": expected,
			"actual":   actualLen,
		}
	})
}
