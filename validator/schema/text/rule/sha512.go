// schema/text/rule/sha512.go
package rule

import "github.com/leandroluk/go/validator/internal/ruleset"

func SHA512(code string) ruleset.Rule[string] {
	return digestRule(code, "invalid sha512", 64)
}
