// schema/text/rule/blake2b_256.go
package rule

import "github.com/leandroluk/gox/validate/internal/ruleset"

func BLAKE2B_256(code string) ruleset.Rule[string] {
	return digestRule(code, "invalid blake2b-256", 32)
}
