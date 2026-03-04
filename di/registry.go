package di

import (
	"fmt"
	"reflect"
)

// registerProvider handles the low-level logic of adding a factory to the registry.
// It performs basic validation on return types and assignability.
// Empty name ("") indicates an unnamed (default) provider.
func registerProvider(factoryFN any, isSingleton bool, asType reflect.Type, name string) {
	registerProviderWithLifecycle(factoryFN, isSingleton, asType, name, nil)
}

// registerProviderWithLifecycle registers a provider with optional lifecycle hooks.
func registerProviderWithLifecycle(factoryFN any, isSingleton bool, asType reflect.Type, name string, hooks *LifecycleHooks) {
	if factoryFN == nil {
		Fail("di: nil factory function provided")
	}

	factoryValue := reflect.ValueOf(factoryFN)
	factoryType := factoryValue.Type()

	if factoryType.Kind() != reflect.Func {
		Fail("di: factory must be a function")
	}

	if factoryType.NumOut() != 1 {
		Fail("di: factory function must return exactly one value")
	}

	outputType := factoryType.Out(0)

	// If an explicit type (like an interface) is provided, check assignability.
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

	hasLifecycle := hooks != nil && (hooks.OnStart != nil || hooks.OnStop != nil)
	lifecycleStr := ""
	if hasLifecycle {
		lifecycleStr = " with lifecycle"
	}

	LogDebug("Registering %v (%s, Singleton: %v%s)", outputType, nameStr, isSingleton, lifecycleStr)

	providerInstance := &Provider{
		Name:            name,
		FactoryFunction: factoryValue,
		OutputType:      outputType,
		IsSingleton:     isSingleton,
		hasLifecycle:    hasLifecycle,
	}

	// Wrap lifecycle hooks to work with reflect.Value
	if hooks != nil {
		if hooks.OnStart != nil {
			providerInstance.OnStartHook = func(v reflect.Value) error {
				return hooks.OnStart(v.Interface())
			}
		}
		if hooks.OnStop != nil {
			providerInstance.OnStopHook = func(v reflect.Value) error {
				return hooks.OnStop(v.Interface())
			}
		}
	}

	RegistryMutex.Lock()
	defer RegistryMutex.Unlock()

	// Initialize map for this type if doesn't exist
	if ProviderRegistry[outputType] == nil {
		ProviderRegistry[outputType] = make(map[string]*Provider)
	}

	// Check if name already exists
	if _, exists := ProviderRegistry[outputType][name]; exists {
		nameDesc := "unnamed provider"
		if name != "" {
			nameDesc = fmt.Sprintf("provider with name %q", name)
		}
		Fail(fmt.Sprintf("di: %s for type %v is already registered", nameDesc, outputType))
	}

	ProviderRegistry[outputType][name] = providerInstance

	// Add to lifecycle management if has hooks
	if hasLifecycle {
		addLifecycleProvider(providerInstance)
	}
}
