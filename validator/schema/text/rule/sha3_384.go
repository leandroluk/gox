// schema/text/rule/sha3_384.go
package rule

import "github.com/leandroluk/go/validator/internal/ruleset"

func SHA3_384(code string) ruleset.Rule[string] {
	return digestRule(code, "invalid sha3-384", 48)
}
