// schema/text/rule/one_of.go
package rule

import "github.com/leandroluk/gox/validate/internal/ruleset"

func OneOf(code string, values ...string) ruleset.Rule[string] {
	allowed := make(map[string]struct{}, len(values))
	for _, v := range values {
		allowed[v] = struct{}{}
	}

	return newRule(code, "not allowed", func(actual string) (bool, map[string]any) {
		if _, ok := allowed[actual]; ok {
			return true, nil
		}
		return false, map[string]any{
			"allowed": values,
			"actual":  actual,
		}
	})
}
