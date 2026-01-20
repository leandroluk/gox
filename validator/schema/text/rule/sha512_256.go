// schema/text/rule/sha512_256.go
package rule

import "github.com/leandroluk/go/validator/internal/ruleset"

func SHA512_256(code string) ruleset.Rule[string] {
	return digestRule(code, "invalid sha512/256", 32)
}
