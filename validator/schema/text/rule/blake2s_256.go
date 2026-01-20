// schema/text/rule/blake2s_256.go
package rule

import "github.com/leandroluk/go/validator/internal/ruleset"

func BLAKE2S_256(code string) ruleset.Rule[string] {
	return digestRule(code, "invalid blake2s-256", 32)
}
