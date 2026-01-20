// schema/object/rule/comparator_helpers.go
package rule

import (
	"strconv"

	"github.com/leandroluk/gox/validate/internal/ast"
)

func compareOrder(left ast.Value, right ast.Value) (int, bool) {
	if left.IsMissing() || left.IsNull() || right.IsMissing() || right.IsNull() {
		return 0, false
	}

	if left.Kind != right.Kind {
		return 0, false
	}

	switch left.Kind {
	case ast.KindNumber:
		lf, err1 := strconv.ParseFloat(left.Number, 64)
		rf, err2 := strconv.ParseFloat(right.Number, 64)
		if err1 != nil || err2 != nil {
			return 0, false
		}
		if lf < rf {
			return -1, true
		}
		if lf > rf {
			return 1, true
		}
		return 0, true

	case ast.KindString:
		if left.String < right.String {
			return -1, true
		}
		if left.String > right.String {
			return 1, true
		}
		return 0, true

	default:
		return 0, false
	}
}

func valuesEqual(left ast.Value, right ast.Value) bool {
	return astValueEqual(left, right)
}
