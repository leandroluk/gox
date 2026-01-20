// internal/defaults/apply.go
package defaults

import (
	"github.com/leandroluk/gox/validate/internal/ast"
	"github.com/leandroluk/gox/validate/schema"
)

type Provider[T any] struct {
	set   bool
	value T
	fn    func() T
}

func None[T any]() Provider[T] {
	return Provider[T]{}
}

func Value[T any](value T) Provider[T] {
	return Provider[T]{set: true, value: value}
}

func Func[T any](fn func() T) Provider[T] {
	if fn == nil {
		return Provider[T]{}
	}
	return Provider[T]{set: true, fn: fn}
}

func (provider Provider[T]) IsSet() bool {
	return provider.set
}

func (provider Provider[T]) Provide() T {
	if provider.fn != nil {
		return provider.fn()
	}
	return provider.value
}

func ShouldApply(presence ast.Presence, options schema.Options) bool {
	if presence == ast.Missing {
		return true
	}
	if presence == ast.Null && options.DefaultOnNull {
		return true
	}
	return false
}

func Apply[T any](presence ast.Presence, options schema.Options, provider Provider[T]) (T, bool) {
	if !provider.set {
		var zero T
		return zero, false
	}
	if !ShouldApply(presence, options) {
		var zero T
		return zero, false
	}
	return provider.Provide(), true
}
