// schema/text/rule/semver.go
package rule

import "github.com/leandroluk/go/validator/internal/ruleset"

func SemVer(code string) ruleset.Rule[string] {
	return newRule(code, "invalid semver", func(actual string) (bool, map[string]any) {
		return isSemVer(actual), map[string]any{"actual": actual}
	})
}
