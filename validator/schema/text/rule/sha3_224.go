// schema/text/rule/sha3_224.go
package rule

import "github.com/leandroluk/go/validator/internal/ruleset"

func SHA3_224(code string) ruleset.Rule[string] {
	return digestRule(code, "invalid sha3-224", 28)
}
