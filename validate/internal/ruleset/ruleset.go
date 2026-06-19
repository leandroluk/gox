// internal/ruleset/ruleset.go
package ruleset

import "github.com/leandroluk/gox/validate/internal/engine"

type Reporter interface {
	AddIssue(code string, message string, meta ...map[string]any) bool
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

type Set[T any] struct {
	order []string
	byKey map[string]Rule[T]
	list  []Rule[T]
}

func NewSet[T any]() *Set[T] {
	return &Set[T]{
		order: make([]string, 0),
		byKey: make(map[string]Rule[T]),
		list:  make([]Rule[T], 0),
	}
}

func (set *Set[T]) Put(ruleValue Rule[T]) {
	if ruleValue.Key == "" {
		set.list = append(set.list, ruleValue)
		return
	}

	if _, exists := set.byKey[ruleValue.Key]; !exists {
		set.order = append(set.order, ruleValue.Key)
	}

	set.byKey[ruleValue.Key] = ruleValue
}

func (set *Set[T]) Remove(key string) {
	if key == "" {
		return
	}

	if _, exists := set.byKey[key]; !exists {
		return
	}

	delete(set.byKey, key)

	out := set.order[:0]
	for _, existingKey := range set.order {
		if existingKey != key {
			out = append(out, existingKey)
		}
	}
	set.order = out
}

func (set *Set[T]) ApplyAll(value T, context *engine.Context) (T, bool) {
	for _, key := range set.order {
		ruleValue := set.byKey[key]
		if ruleValue.Apply == nil {
			continue
		}

		var stopped bool
		value, stopped = ruleValue.Apply(value, context)
		if stopped {
			return value, true
		}
	}

	for _, ruleValue := range set.list {
		if ruleValue.Apply == nil {
			continue
		}

		var stopped bool
		value, stopped = ruleValue.Apply(value, context)
		if stopped {
			return value, true
		}
	}

	return value, false
}
