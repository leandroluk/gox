// schema/text/rule/md5.go
package rule

import "github.com/leandroluk/go/validate/internal/ruleset"

func MD5(code string) ruleset.Rule[string] {
	return digestRule(code, "invalid md5", 16)
}
