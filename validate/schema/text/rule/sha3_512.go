// schema/text/rule/sha3_512.go
package rule

import "github.com/leandroluk/go/validate/internal/ruleset"

func SHA3_512(code string) ruleset.Rule[string] {
	return digestRule(code, "invalid sha3-512", 64)
}
