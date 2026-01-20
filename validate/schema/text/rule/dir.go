// schema/text/rule/dir.go
package rule

import "github.com/leandroluk/go/validate/internal/ruleset"

func Dir(code string) ruleset.Rule[string] {
	return newRule(code, "invalid dir", func(actual string) (bool, map[string]any) {
		return isDir(actual), map[string]any{"actual": actual}
	})
}
