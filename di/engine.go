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
		fail(fmt.Sprintf("di: no provider registered for type %v", targetType))
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
