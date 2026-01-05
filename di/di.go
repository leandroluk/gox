package di

import (
	"fmt"
	"reflect"
	"sync"
)

var (
	registryMutex    sync.RWMutex
	providerRegistry = map[reflect.Type][]*Provider{}
)

// Provider holds the necessary information to create and manage an instance.
type Provider struct {
	FactoryFunction reflect.Value // The function used to create the instance.
	OutputType      reflect.Type  // The reflected type of the result.
	IsSingleton     bool          // Indicates if it should return the same instance every time.
	CachedInstance  reflect.Value // Stores the instance if it's a singleton.
}

// resolveByType finds the registered provider and triggers the build process.
func resolveByType(targetType reflect.Type) reflect.Value {
	registryMutex.RLock()
	providers := providerRegistry[targetType]
	registryMutex.RUnlock()

	if len(providers) == 0 {
		panic(fmt.Sprintf("di: no provider registered for type %v", targetType))
	}

	return buildInstance(providers[0])
}

// buildInstance manages the lifecycle of the instance (Transient vs Singleton).
func buildInstance(providerInstance *Provider) reflect.Value {
	if providerInstance.IsSingleton {
		registryMutex.Lock()
		defer registryMutex.Unlock()

		if providerInstance.CachedInstance.IsValid() {
			return providerInstance.CachedInstance
		}

		instance := callFactoryWithDependencies(providerInstance)
		providerInstance.CachedInstance = instance
		return instance
	}

	return callFactoryWithDependencies(providerInstance)
}

// callFactoryWithDependencies recursively resolves all inputs of a factory function.
func callFactoryWithDependencies(providerInstance *Provider) reflect.Value {
	factoryType := providerInstance.FactoryFunction.Type()
	numberOfInputs := factoryType.NumIn()

	arguments := make([]reflect.Value, numberOfInputs)

	for index := 0; index < numberOfInputs; index++ {
		dependencyType := factoryType.In(index)
		arguments[index] = resolveByType(dependencyType)
	}

	outputValues := providerInstance.FactoryFunction.Call(arguments)
	return outputValues[0]
}

// registerProvider handles the low-level logic of adding a factory to the registry.
// It performs basic validation on return types and assignability.
func registerProvider(factoryFN any, isSingleton bool, asType reflect.Type) {
	if factoryFN == nil {
		panic("di: nil factory function provided")
	}

	factoryValue := reflect.ValueOf(factoryFN)
	factoryType := factoryValue.Type()

	if factoryType.Kind() != reflect.Func {
		panic("di: factory must be a function")
	}

	if factoryType.NumOut() != 1 {
		panic("di: factory function must return exactly one value")
	}

	outputType := factoryType.Out(0)

	// If an explicit type (like an interface) is provided, check assignability.
	if asType != nil {
		if !outputType.AssignableTo(asType) {
			panic(fmt.Sprintf("di: factory return type %v is not assignable to %v", outputType, asType))
		}
		outputType = asType
	}

	providerInstance := &Provider{
		FactoryFunction: factoryValue,
		OutputType:      outputType,
		IsSingleton:     isSingleton,
	}

	registryMutex.Lock()
	defer registryMutex.Unlock()
	providerRegistry[outputType] = append(providerRegistry[outputType], providerInstance)
}

// Register adds a transient provider (a new instance is created every time it's resolved).
func Register(factoryFN any) {
	registerProvider(factoryFN, false, nil)
}

// RegisterAs adds a transient provider bound to a specific interface or type T.
func RegisterAs[T any](factoryFN any) {
	registerProvider(factoryFN, false, reflect.TypeFor[T]())
}

// Singleton adds a provider that caches its instance after the first resolution.
func Singleton(factoryFN any) {
	registerProvider(factoryFN, true, nil)
}

// SingletonAs adds a singleton provider bound to a specific interface or type T.
func SingletonAs[T any](factoryFN any) {
	registerProvider(factoryFN, true, reflect.TypeFor[T]())
}

// Resolve retrieves the primary instance for type T. Panics if no provider is found.
func Resolve[T any]() T {
	targetType := reflect.TypeFor[T]()
	value := resolveByType(targetType)
	return value.Interface().(T)
}

// ResolveAll retrieves all registered providers for type T as a slice.
func ResolveAll[T any]() []T {
	targetType := reflect.TypeFor[T]()

	registryMutex.RLock()
	providers := providerRegistry[targetType]
	registryMutex.RUnlock()

	if len(providers) == 0 {
		return nil
	}

	results := make([]T, 0, len(providers))
	for _, providerInstance := range providers {
		value := buildInstance(providerInstance)
		results = append(results, value.Interface().(T))
	}

	return results
}

// Reset clears all registered providers.
// Primarily used for unit tests to ensure isolation.
func Reset() {
	registryMutex.Lock()
	defer registryMutex.Unlock()
	providerRegistry = make(map[reflect.Type][]*Provider)
}
