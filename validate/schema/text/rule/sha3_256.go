// schema/text/rule/sha3_256.go
package rule

import "github.com/leandroluk/gox/validate/internal/ruleset"

func SHA3_256(code string) ruleset.Rule[string] {
	return digestRule(code, "invalid sha3-256", 32)
}
