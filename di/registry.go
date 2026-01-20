package di

import (
	"fmt"
	"reflect"
)

// registerProvider handles the low-level logic of adding a factory to the registry.
// It performs basic validation on return types and assignability.
func registerProvider(factoryFN any, isSingleton bool, asType reflect.Type) {
	if factoryFN == nil {
		fail("di: nil factory function provided")
	}

	factoryValue := reflect.ValueOf(factoryFN)
	factoryType := factoryValue.Type()

	if factoryType.Kind() != reflect.Func {
		fail("di: factory must be a function")
	}

	if factoryType.NumOut() != 1 {
		fail("di: factory function must return exactly one value")
	}

	outputType := factoryType.Out(0)

	// If an explicit type (like an interface) is provided, check assignability.
	if asType != nil {
		if !outputType.AssignableTo(asType) {
			fail(fmt.Sprintf("di: factory return type %v is not assignable to %v", outputType, asType))
		}
		outputType = asType
	}

	logDebug("Registering %v (Singleton: %v)", outputType, isSingleton)

	providerInstance := &Provider{
		FactoryFunction: factoryValue,
		OutputType:      outputType,
		IsSingleton:     isSingleton,
	}

	registryMutex.Lock()
	defer registryMutex.Unlock()
	providerRegistry[outputType] = append(providerRegistry[outputType], providerInstance)
}
