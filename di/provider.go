package di

import (
	"reflect"
	"sync"
	"sync/atomic"
)

type Options[T any] struct {
	Constructor            func() (T, error)
	OnApplicationBootstrap func(T) error
	OnApplicationShutdown  func(T) error
}

type Provider struct {
	Name            string
	FactoryFunction reflect.Value
	OutputType      reflect.Type
	IsSingleton     bool
	CachedInstance  reflect.Value
	initOnce        sync.Once
	resolving       atomic.Bool

	OnStartHook  func(reflect.Value) error
	OnStopHook   func(reflect.Value) error
	hasLifecycle bool
	started      atomic.Bool
}

func Reset() {
	RegistryMutex.Lock()
	ProviderRegistry = make(map[reflect.Type]map[string]*Provider)
	RegistryMutex.Unlock()

	lifecycleMutex.Lock()
	lifecycleProviders = nil
	lifecycleMutex.Unlock()
}
