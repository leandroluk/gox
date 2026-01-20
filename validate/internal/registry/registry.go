// internal/registry/registry.go
package registry

import (
	"reflect"
	"sync"

	"github.com/leandroluk/go/validate/schema"
)

type Registry struct {
	mutex  sync.RWMutex
	byType map[reflect.Type]schema.AnySchema
}

func New() *Registry {
	return &Registry{
		byType: make(map[reflect.Type]schema.AnySchema),
	}
}

func (registry *Registry) Register(schemaValue schema.AnySchema) {
	if schemaValue == nil {
		return
	}
	outputType := schemaValue.OutputType()
	if outputType == nil {
		return
	}

	registry.mutex.Lock()
	registry.byType[outputType] = schemaValue
	registry.mutex.Unlock()
}

func (registry *Registry) Lookup(outputType reflect.Type) (schema.AnySchema, bool) {
	if outputType == nil {
		return nil, false
	}

	registry.mutex.RLock()
	value, ok := registry.byType[outputType]
	registry.mutex.RUnlock()

	return value, ok
}

func (registry *Registry) Reset() {
	registry.mutex.Lock()
	registry.byType = make(map[reflect.Type]schema.AnySchema)
	registry.mutex.Unlock()
}

var global = New()

func Register(schemaValue schema.AnySchema) {
	global.Register(schemaValue)
}

func Lookup(outputType reflect.Type) (schema.AnySchema, bool) {
	return global.Lookup(outputType)
}

func LookupTyped[T any]() (schema.AnySchema, bool) {
	var pointer *T
	outputType := reflect.TypeOf(pointer).Elem()
	return global.Lookup(outputType)
}

func Reset() {
	global.Reset()
}
