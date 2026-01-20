// schema/text/rule/file.go
package rule

import "github.com/leandroluk/gox/validate/internal/ruleset"

func File(code string) ruleset.Rule[string] {
	return newRule(code, "invalid file", func(actual string) (bool, map[string]any) {
		return isFile(actual), map[string]any{"actual": actual}
	})
}
