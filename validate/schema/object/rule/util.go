// schema/object/rule/util.go
package rule

import (
	"strconv"

	"github.com/leandroluk/go/validate/internal/ast"
)

func conditionMet(actual ast.Value, op ConditionOp, expected ast.Value) bool {
	switch op {
	case OpPresent:
		return !actual.IsMissing() && !actual.IsNull()

	case OpMissing:
		return actual.IsMissing()

	case OpNull:
		return actual.IsNull()

	case OpNotNull:
		return !actual.IsNull()

	case OpEq:
		return astValueEqual(actual, expected)

	case OpNeq:
		return !astValueEqual(actual, expected)

	default:
		return false
	}
}

func astValueEqual(actual ast.Value, expected ast.Value) bool {
	if expected.IsNull() {
		return actual.IsNull()
	}

	if actual.IsMissing() || actual.IsNull() {
		return false
	}

	if actual.Kind != expected.Kind {
		return false
	}

	switch expected.Kind {
	case ast.KindBoolean:
		return actual.Boolean == expected.Boolean
	case ast.KindString:
		return actual.String == expected.String
	case ast.KindNumber:
		return normalizeNumberText(actual.Number) == normalizeNumberText(expected.Number)
	default:
		return false
	}
}

func normalizeNumberText(text string) string {
	if text == "" {
		return text
	}

	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return text
	}

	return strconv.FormatFloat(value, 'g', -1, 64)
}

func astValueToMeta(value ast.Value) any {
	if value.IsMissing() {
		return "(missing)"
	}
	if value.IsNull() {
		return nil
	}

	switch value.Kind {
	case ast.KindBoolean:
		return value.Boolean
	case ast.KindString:
		return value.String
	case ast.KindNumber:
		return value.Number
	default:
		return value.Kind.String()
	}
}
