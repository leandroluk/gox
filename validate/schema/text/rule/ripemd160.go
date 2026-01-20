// schema/text/rule/ripemd160.go
package rule

import "github.com/leandroluk/gox/validate/internal/ruleset"

func RIPEMD160(code string) ruleset.Rule[string] {
	return digestRule(code, "invalid ripemd160", 20)
}
