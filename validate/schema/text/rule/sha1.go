// schema/text/rule/sha1.go
package rule

import "github.com/leandroluk/go/validate/internal/ruleset"

func SHA1(code string) ruleset.Rule[string] {
	return digestRule(code, "invalid sha1", 20)
}
