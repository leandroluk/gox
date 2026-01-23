package di

import (
	"fmt"
	"reflect"
)

// resolveByType finds the registered provider and triggers the build process.
func resolveByType(targetType reflect.Type) reflect.Value {
	logDebug("Resolving %v", targetType)

	registryMutex.RLock()
	providers := providerRegistry[targetType]
	registryMutex.RUnlock()

	if len(providers) == 0 {
		registryMutex.RLock()
		handlers := make([]func(reflect.Type) (any, bool), 0)
		registryMutex.RUnlock()

		for _, handler := range handlers {
			if instance, ok := handler(targetType); ok {
				return reflect.ValueOf(instance)
			}
		}

		fail(fmt.Sprintf("di: no provider registered for type %v", targetType))
	}

	return buildInstance(providers[0])
}

// buildInstance manages the lifecycle of the instance (Transient vs Singleton).
func buildInstance(providerInstance *Provider) reflect.Value {
	if providerInstance.IsSingleton {
		providerInstance.initOnce.Do(func() {
			providerInstance.CachedInstance = callFactoryWithDependencies(providerInstance)
		})
		return providerInstance.CachedInstance
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
