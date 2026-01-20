// schema/text/rule/sha224.go
package rule

import "github.com/leandroluk/gox/validate/internal/ruleset"

func SHA224(code string) ruleset.Rule[string] {
	return digestRule(code, "invalid sha224", 28)
}
