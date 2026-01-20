// internal/ruleset/rule.go
package ruleset

import "github.com/leandroluk/gox/validate/internal/engine"

type Reporter interface {
	AddIssue(code string, message string) bool
	AddIssueWithMeta(code string, message string, meta map[string]any) bool
}

type RuleFn[T any] func(value T, reporter Reporter) bool

type ApplyFunction[T any] func(value T, context *engine.Context) (T, bool)

type Rule[T any] struct {
	Key   string
	Apply ApplyFunction[T]
}

func New[T any](key string, apply ApplyFunction[T]) Rule[T] {
	return Rule[T]{Key: key, Apply: apply}
}

func Apply[T any](value T, reporter Reporter, rules ...RuleFn[T]) bool {
	for _, rule := range rules {
		if rule == nil {
			continue
		}
		if rule(value, reporter) {
			return true
		}
	}
	return false
}
