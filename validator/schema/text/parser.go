// schema/text/parser.go
package text

import (
	"github.com/leandroluk/go/validator/internal/ast"
	"github.com/leandroluk/go/validator/internal/engine"
)

func parseTextValue(context *engine.Context, value ast.Value) (string, bool) {
	switch value.Kind {
	case ast.KindString:
		return value.String, false

	case ast.KindNumber:
		if context.Options.Coerce {
			return value.Number, false
		}
		stop := context.AddIssueWithMeta(CodeType, "expected string", map[string]any{
			"expected": "string",
			"actual":   value.Kind.String(),
		})
		return "", stop

	case ast.KindBoolean:
		if context.Options.Coerce {
			if value.Boolean {
				return "true", false
			}
			return "false", false
		}
		stop := context.AddIssueWithMeta(CodeType, "expected string", map[string]any{
			"expected": "string",
			"actual":   value.Kind.String(),
		})
		return "", stop

	default:
		stop := context.AddIssueWithMeta(CodeType, "expected string", map[string]any{
			"expected": "string",
			"actual":   value.Kind.String(),
		})
		return "", stop
	}
}
