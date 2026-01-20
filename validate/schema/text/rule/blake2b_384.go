// schema/text/rule/blake2b_384.go
package rule

import "github.com/leandroluk/go/validate/internal/ruleset"

func BLAKE2B_384(code string) ruleset.Rule[string] {
	return digestRule(code, "invalid blake2b-384", 48)
}
