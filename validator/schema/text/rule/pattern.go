// schema/text/rule/pattern.go
package rule

import (
	"regexp"

	"github.com/leandroluk/go/validator/internal/ruleset"
)

func Pattern(code string, pattern *regexp.Regexp) ruleset.Rule[string] {
	patternString := ""
	if pattern != nil {
		patternString = pattern.String()
	}

	return newRule(code, "does not match pattern", func(actual string) (bool, map[string]any) {
		if pattern != nil && pattern.MatchString(actual) {
			return true, nil
		}
		return false, map[string]any{
			"pattern": patternString,
			"actual":  actual,
		}
	})
}
