// schema/text/rule/blake2b_512.go
package rule

import "github.com/leandroluk/go/validator/internal/ruleset"

func BLAKE2B_512(code string) ruleset.Rule[string] {
	return digestRule(code, "invalid blake2b-512", 64)
}
