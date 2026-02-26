// github.com/leandroluk/gox/oas/adapter/fiber/params.go
package adapter

import "regexp"

// ExtractPathParams extracts path parameter names in Fiber format.
// Supports normal parameters (:param) and optional ones (:param?).
//
// Examples:
//
//	/users/:id           → ["id"]
//	/users/:id/:field    → ["id", "field"]
//	/files/:name?        → ["name"]
func (g *Group) ExtractPathParams(routePath string) []string {
	re := regexp.MustCompile(`:([a-zA-Z0-9_]+)\??`)
	matches := re.FindAllStringSubmatch(routePath, -1)

	params := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			params = append(params, match[1])
		}
	}
	return params
}
