// schema/text/rule/data_uri.go
package rule

import "github.com/leandroluk/gox/validate/internal/ruleset"

func DataURI(code string) ruleset.Rule[string] {
	return newRule(code, "invalid data uri", func(actual string) (bool, map[string]any) {
		return isDataURI(actual), map[string]any{"actual": actual}
	})
}
