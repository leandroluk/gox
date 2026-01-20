// schema/object/rule/types.go
package rule

import (
	"fmt"

	"github.com/leandroluk/go/validate/internal/ast"
	"github.com/leandroluk/go/validate/internal/engine"
)

var ErrUnsupportedExpectedValue = fmt.Errorf("unsupported expected value")

type ConditionOp string

const (
	OpEq      ConditionOp = "eq"
	OpNeq     ConditionOp = "neq"
	OpPresent ConditionOp = "present"
	OpMissing ConditionOp = "missing"
	OpNull    ConditionOp = "null"
	OpNotNull ConditionOp = "notnull"
)

type Comparator interface {
	Apply(context *engine.Context, root ast.Value, child ast.Value) bool
}

type RequiredCondition interface {
	Apply(context *engine.Context, root ast.Value, child ast.Value, childPresent bool) bool
}

type ExcludedIfCondition struct {
	code     string
	path     string
	op       ConditionOp
	expected ast.Value
}

type SkipUnlessCondition struct {
	path     string
	op       ConditionOp
	expected ast.Value
}
