// schema/text/rule/file_path.go
package rule

import "github.com/leandroluk/go/validate/internal/ruleset"

func FilePath(code string) ruleset.Rule[string] {
	return newRule(code, "invalid filepath", func(actual string) (bool, map[string]any) {
		return isFilePath(actual), map[string]any{"actual": actual}
	})
}
