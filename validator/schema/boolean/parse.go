// schema/boolean/parse.go
package boolean

import (
	"strconv"

	"github.com/leandroluk/go/validator/internal/ast"
	"github.com/leandroluk/go/validator/schema"
)

func parseBooleanWithOptions(options schema.Options, value ast.Value) (bool, bool) {
	if value.Kind == ast.KindBoolean {
		return value.Boolean, true
	}

	if !options.Coerce {
		return false, false
	}

	switch value.Kind {
	case ast.KindString:
		parsed, err := strconv.ParseBool(value.String)
		if err != nil {
			return false, false
		}
		return parsed, true

	case ast.KindNumber:
		if value.Number == "0" {
			return false, true
		}
		if value.Number == "1" {
			return true, true
		}
		return false, false

	default:
		return false, false
	}
}
