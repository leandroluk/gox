// schema/text/rule/md4.go
package rule

import "github.com/leandroluk/gox/validate/internal/ruleset"

func MD4(code string) ruleset.Rule[string] {
	return digestRule(code, "invalid md4", 16)
}
