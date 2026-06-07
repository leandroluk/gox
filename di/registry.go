package di

import (
	"fmt"
	"reflect"
	"sync"
)

var (
	RegistryMutex    sync.RWMutex
	ProviderRegistry = map[reflect.Type]map[string]*Provider{}
)

func registerProvider[T any](opts *Options[T], isSingleton bool, asType reflect.Type, name string) {
	if opts.Constructor == nil {
		Fail("di: nil factory function provided")
	}

	factoryValue := reflect.ValueOf(opts.Constructor)
	factoryType := factoryValue.Type()

	if factoryType.Kind() != reflect.Func {
		Fail("di: factory must be a function")
	}

	numOut := factoryType.NumOut()
	if numOut != 2 {
		Fail("di: factory function must return exactly two values (T, error)")
	}

	errorType := reflect.TypeFor[error]()
	if !factoryType.Out(1).AssignableTo(errorType) {
		Fail("di: second return value of factory function must be an error")
	}

	outputType := factoryType.Out(0)

	if asType != nil {
		if !outputType.AssignableTo(asType) {
			Fail(fmt.Sprintf("di: factory return type %v is not assignable to %v", outputType, asType))
		}
		outputType = asType
	}

	nameStr := "unnamed"
	if name != "" {
		nameStr = fmt.Sprintf("named: %q", name)
	}

	providerInstance := &Provider{
		Name:            name,
		FactoryFunction: factoryValue,
		OutputType:      outputType,
		IsSingleton:     isSingleton,
	}

	if opts.OnApplicationBootstrap != nil {
		hook := opts.OnApplicationBootstrap
		providerInstance.OnStartHook = func(instance reflect.Value) error {
			val := instance.Interface()
			typedVal, ok := val.(T)
			if !ok {
				return fmt.Errorf("di: cannot apply OnApplicationBootstrap hook for type %T to instance of type %T", typedVal, val)
			}
			return hook(typedVal)
		}
		providerInstance.hasLifecycle = true
	}

	if opts.OnApplicationShutdown != nil {
		hook := opts.OnApplicationShutdown
		providerInstance.OnStopHook = func(instance reflect.Value) error {
			val := instance.Interface()
			typedVal, ok := val.(T)
			if !ok {
				return fmt.Errorf("di: cannot apply OnApplicationShutdown hook for type %T to instance of type %T", typedVal, val)
			}
			return hook(typedVal)
		}
		providerInstance.hasLifecycle = true
	}

	lifecycleStr := ""
	if providerInstance.hasLifecycle {
		lifecycleStr = " with lifecycle"
	}

	LogDebug("Registering %v (%s, Singleton: %v%s)", outputType, nameStr, isSingleton, lifecycleStr)

	RegistryMutex.Lock()
	defer RegistryMutex.Unlock()

	if ProviderRegistry[outputType] == nil {
		ProviderRegistry[outputType] = make(map[string]*Provider)
	}

	if _, exists := ProviderRegistry[outputType][name]; exists {
		nameDesc := "unnamed provider"
		if name != "" {
			nameDesc = fmt.Sprintf("provider with name %q", name)
		}
		Fail(fmt.Sprintf("di: %s for type %v is already registered", nameDesc, outputType))
	}

	ProviderRegistry[outputType][name] = providerInstance

	if providerInstance.hasLifecycle {
		addLifecycleProvider(providerInstance)
	}
}
