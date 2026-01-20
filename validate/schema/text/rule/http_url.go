// schema/text/rule/http_url.go
package rule

import "github.com/leandroluk/go/validate/internal/ruleset"

func HTTPURL(code string) ruleset.Rule[string] {
	return newRule(code, "invalid http url", func(actual string) (bool, map[string]any) {
		return isHTTPURL(actual), map[string]any{"actual": actual}
	})
}
