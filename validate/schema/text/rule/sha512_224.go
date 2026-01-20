// schema/text/rule/sha512_224.go
package rule

import "github.com/leandroluk/go/validate/internal/ruleset"

func SHA512_224(code string) ruleset.Rule[string] {
	return digestRule(code, "invalid sha512/224", 28)
}
