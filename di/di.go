package di

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

// Scope controls how instances are created.
type Scope int

const (
	ScopeSingleton Scope = iota // one shared instance (default)
	ScopeTransient              // new instance per resolution
)

// Builder[T] configures providers for type T.
type Builder[T any] interface {
	New(ctor func() (T, error)) *Registration[T]
	Named(name string, ctor func() (T, error)) *Registration[T]
	Instance(val T) *Registration[T]
	Extend(ptr any) *Registration[T]
}

// Registration is the fluent chain returned by builder methods.
type Registration[T any] struct{ e *entry }

func (r *Registration[T]) Scope(s Scope) *Registration[T] {
	if !r.e.scopeLocked {
		r.e.scope = s
	}
	return r
}

func (r *Registration[T]) Multi() *Registration[T] {
	r.e.multi = true
	return r
}

func (r *Registration[T]) OnStart(fn func(T) error) *Registration[T] {
	r.e.onStart = func(v any) error { return fn(v.(T)) }
	addToLifecycle(r.e)
	return r
}

func (r *Registration[T]) OnStop(fn func(T) error) *Registration[T] {
	r.e.onStop = func(v any) error { return fn(v.(T)) }
	addToLifecycle(r.e)
	return r
}

// --- internal ---

type entry struct {
	key         string
	typ         reflect.Type
	factory     func() (any, error)
	scope       Scope
	scopeLocked bool
	multi       bool

	once      sync.Once
	cached    any
	resolving atomic.Bool

	onStart     func(any) error
	onStop      func(any) error
	started     atomic.Bool
	inLifecycle bool
}

var (
	mu    sync.RWMutex
	store = map[reflect.Type]map[string]*entry{}

	lcMu  sync.RWMutex
	lcAll []*entry
)

func addToLifecycle(e *entry) {
	if e.inLifecycle {
		return
	}
	e.inLifecycle = true
	lcMu.Lock()
	lcAll = append(lcAll, e)
	lcMu.Unlock()
}

type builderImpl[T any] struct{ typ reflect.Type }

func (b *builderImpl[T]) New(ctor func() (T, error)) *Registration[T] {
	return b.add("", func() (any, error) { return ctor() }, ScopeSingleton, false)
}

func (b *builderImpl[T]) Named(name string, ctor func() (T, error)) *Registration[T] {
	if name == "" {
		panic("di: Named requires non-empty name")
	}
	return b.add(name, func() (any, error) { return ctor() }, ScopeSingleton, false)
}

func (b *builderImpl[T]) Instance(val T) *Registration[T] {
	return b.add("", func() (any, error) { return val, nil }, ScopeSingleton, true)
}

func (b *builderImpl[T]) Extend(ptr any) *Registration[T] {
	if ptr == nil {
		panic("di: Extend requires a non-nil pointer")
	}
	pv := reflect.ValueOf(ptr)
	if pv.Kind() != reflect.Ptr {
		panic("di: Extend requires a pointer to a type variable (e.g. var x MyInterface; b.Extend(&x))")
	}
	srcType := pv.Type().Elem()
	return b.add("", func() (any, error) {
		return resolveType(srcType), nil
	}, ScopeSingleton, false)
}

func (b *builderImpl[T]) add(key string, factory func() (any, error), scope Scope, locked bool) *Registration[T] {
	e := &entry{
		key:         key,
		typ:         b.typ,
		factory:     factory,
		scope:       scope,
		scopeLocked: locked,
	}
	mu.Lock()
	if store[b.typ] == nil {
		store[b.typ] = make(map[string]*entry)
	}
	if _, exists := store[b.typ][key]; exists {
		mu.Unlock()
		if key == "" {
			panic(fmt.Sprintf("di: unnamed provider for %v already registered", b.typ))
		}
		panic(fmt.Sprintf("di: provider named %q for %v already registered", key, b.typ))
	}
	store[b.typ][key] = e
	mu.Unlock()
	return &Registration[T]{e: e}
}

// Register configures one provider for type T via the builder.
func Register[T any](configurator func(Builder[T])) {
	if configurator == nil {
		return
	}
	configurator(&builderImpl[T]{typ: reflect.TypeFor[T]()})
}

// Resolve returns the default (unnamed) instance of T. Panics if not registered.
func Resolve[T any]() T {
	return resolveType(reflect.TypeFor[T]()).(T)
}

// ResolveNamed returns the named instance of T. Panics if not registered.
func ResolveNamed[T any](name string) T {
	return resolveNamed(reflect.TypeFor[T](), name).(T)
}

// TryResolve returns the default instance and true, or zero value and false if not registered.
func TryResolve[T any]() (T, bool) {
	var zero T
	mu.RLock()
	m := store[reflect.TypeFor[T]()]
	mu.RUnlock()
	if m == nil || m[""] == nil {
		return zero, false
	}
	return buildEntry(m[""]).(T), true
}

// TryResolveNamed returns the named instance and true, or zero value and false if not registered.
func TryResolveNamed[T any](name string) (T, bool) {
	var zero T
	mu.RLock()
	m := store[reflect.TypeFor[T]()]
	mu.RUnlock()
	if m == nil || m[name] == nil {
		return zero, false
	}
	return buildEntry(m[name]).(T), true
}

// ResolveAll returns all instances of T marked with Multi().
func ResolveAll[T any]() []T {
	mu.RLock()
	m := store[reflect.TypeFor[T]()]
	mu.RUnlock()
	var out []T
	for _, e := range m {
		if e.multi {
			out = append(out, buildEntry(e).(T))
		}
	}
	return out
}

// Reset clears all registrations and lifecycle state. Use in tests.
func Reset() {
	mu.Lock()
	store = make(map[reflect.Type]map[string]*entry)
	mu.Unlock()
	lcMu.Lock()
	lcAll = nil
	lcMu.Unlock()
}

// --- internal resolution ---

func resolveType(typ reflect.Type) any {
	mu.RLock()
	m := store[typ]
	mu.RUnlock()
	if m == nil || m[""] == nil {
		panic(fmt.Sprintf("di: no provider registered for %v", typ))
	}
	return buildEntry(m[""])
}

func resolveNamed(typ reflect.Type, name string) any {
	mu.RLock()
	m := store[typ]
	mu.RUnlock()
	if m == nil || m[name] == nil {
		panic(fmt.Sprintf("di: no provider named %q for %v", name, typ))
	}
	return buildEntry(m[name])
}

func buildEntry(e *entry) any {
	if e.scope == ScopeSingleton {
		if e.resolving.Load() {
			panic(fmt.Sprintf("di: circular dependency detected for %v", e.typ))
		}
		e.once.Do(func() {
			e.resolving.Store(true)
			val, err := e.factory()
			if err != nil {
				panic(fmt.Sprintf("di: factory error for %v: %v", e.typ, err))
			}
			e.cached = val
			e.resolving.Store(false)
		})
		return e.cached
	}
	val, err := e.factory()
	if err != nil {
		panic(fmt.Sprintf("di: factory error for %v: %v", e.typ, err))
	}
	return val
}

// --- lifecycle ---

// StartAll runs OnStart hooks in registration order.
func StartAll() error {
	return StartAllWithContext(context.Background())
}

// StartAllWithContext runs OnStart hooks with context cancellation support.
func StartAllWithContext(ctx context.Context) error {
	lcMu.RLock()
	list := make([]*entry, len(lcAll))
	copy(list, lcAll)
	lcMu.RUnlock()

	var started []*entry
	for _, e := range list {
		if e.onStart == nil || e.started.Load() {
			continue
		}
		select {
		case <-ctx.Done():
			_ = doStop(started, context.Background())
			return ctx.Err()
		default:
		}
		inst := buildEntry(e)
		if err := e.onStart(inst); err != nil {
			_ = doStop(started, context.Background())
			return fmt.Errorf("di: start %v: %w", e.typ, err)
		}
		e.started.Store(true)
		started = append(started, e)
	}
	return nil
}

// StartAllWithTimeout runs StartAllWithContext with a deadline.
func StartAllWithTimeout(d time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()
	return StartAllWithContext(ctx)
}

// StopAll runs OnStop hooks in reverse registration order.
func StopAll() error {
	return StopAllWithContext(context.Background())
}

// StopAllWithContext runs OnStop hooks with context cancellation support.
func StopAllWithContext(ctx context.Context) error {
	lcMu.RLock()
	list := make([]*entry, len(lcAll))
	copy(list, lcAll)
	lcMu.RUnlock()
	return doStop(list, ctx)
}

// StopAllWithTimeout runs StopAllWithContext with a deadline.
func StopAllWithTimeout(d time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()
	return StopAllWithContext(ctx)
}

func doStop(list []*entry, ctx context.Context) error {
	var errs []error
	for i := len(list) - 1; i >= 0; i-- {
		e := list[i]
		if e.onStop == nil || !e.started.Load() || e.cached == nil {
			continue
		}
		select {
		case <-ctx.Done():
			if len(errs) > 0 {
				return fmt.Errorf("di: stop cancelled with %d error(s): %w", len(errs), errs[0])
			}
			return ctx.Err()
		default:
		}
		if err := e.onStop(e.cached); err != nil {
			errs = append(errs, fmt.Errorf("%v: %w", e.typ, err))
		}
		e.started.Store(false)
	}
	if len(errs) > 0 {
		return fmt.Errorf("di: %d stop error(s): %w", len(errs), errs[0])
	}
	return nil
}
