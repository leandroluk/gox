// schema/number/util/number_equal.go
package util

import "github.com/leandroluk/gox/validate/internal/types"

func NumberEqual[N types.Number](a N, b N) bool {
	if IsNaN(a) || IsNaN(b) {
		return IsNaN(a) && IsNaN(b)
	}
	return a == b
}
