// schema/text/rule/dir_path.go
package rule

import "github.com/leandroluk/go/validate/internal/ruleset"

func DirPath(code string) ruleset.Rule[string] {
	return newRule(code, "invalid dirpath", func(actual string) (bool, map[string]any) {
		return isDirPath(actual), map[string]any{"actual": actual}
	})
}
