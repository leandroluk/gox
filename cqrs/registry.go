package cqrs

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/leandroluk/gox/di"
)

type registry struct {
	mutex     sync.RWMutex
	executors map[reflect.Type]func(ctx context.Context, message any) (any, error)
	kindName  string
}

func newRegistry(kindName string) *registry {
	return &registry{
		executors: make(map[reflect.Type]func(context.Context, any) (any, error)),
		kindName:  kindName,
	}
}

func register[TMessage any, TResult any, THandler any](r *registry, factoryFN any) {
	// We use the DI to manage the handler's lifecycle
	di.RegisterAs[THandler](factoryFN)

	messageKey := normalizeType(reflect.TypeFor[TMessage]())

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.executors[messageKey]; exists {
		panic(fmt.Sprintf("cqrs: %s handler already registered for type %v", r.kindName, messageKey))
	}

	r.executors[messageKey] = func(ctx context.Context, message any) (any, error) {
		typedMessage, err := coerce[TMessage](message, r.kindName)
		if err != nil {
			return nil, err
		}

		handlerInstance := di.Resolve[THandler]()

		// Use reflection to call the Handle method
		method := reflect.ValueOf(handlerInstance).MethodByName("Handle")
		results := method.Call([]reflect.Value{
			reflect.ValueOf(ctx),
			reflect.ValueOf(typedMessage),
		})

		errResult := results[1].Interface()
		if errResult != nil {
			return nil, errResult.(error)
		}

		return results[0].Interface(), nil
	}
}

func execute[TResult any](r *registry, ctx context.Context, message any) (TResult, error) {
	var zero TResult
	messageKey, err := normalizedTypeKeyOfValue(message, r.kindName)
	if err != nil {
		return zero, err
	}

	r.mutex.RLock()
	executor, exists := r.executors[messageKey]
	r.mutex.RUnlock()

	if !exists {
		return zero, fmt.Errorf("cqrs: no %s handler registered for type %v", r.kindName, messageKey)
	}

	anyResult, err := executor(ctx, message)
	if err != nil {
		return zero, err
	}

	return coerce[TResult](anyResult, "result")
}
