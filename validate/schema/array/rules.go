// schema/array/rules.go
package array

const (
	CodeRequired = "array.required"
	CodeType     = "array.type"
	CodeMin      = "array.min"
	CodeMax      = "array.max"
	CodeItem     = "array.item"

	CodeLen = "array.len"
	CodeEq  = "array.eq"
	CodeNe  = "array.ne"
	CodeGt  = "array.gt"
	CodeGte = "array.gte"
	CodeLt  = "array.lt"
	CodeLte = "array.lte"

	CodeUnique = "array.unique"
)

func normalizeLimit(value int) (int, bool) {
	if value < 0 {
		return 0, false
	}
	return value, true
}
