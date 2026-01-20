// schema/text/rule/sha384.go
package rule

import "github.com/leandroluk/go/validate/internal/ruleset"

func SHA384(code string) ruleset.Rule[string] {
	return digestRule(code, "invalid sha384", 48)
}
