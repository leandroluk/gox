# di

Minimal, type-safe dependency injection for Go.

## API

### Registration

```go
di.Register[T](func(b di.Builder[T]))
```

One call = one provider. Builder methods:

| Method                             | Description                           |
| ---------------------------------- | ------------------------------------- |
| `b.New(func() (T, error))`         | Unnamed provider via constructor      |
| `b.Named(name, func() (T, error))` | Named provider via constructor        |
| `b.Instance(val T)`                | Pre-built instance (always singleton) |
| `b.Extend(ptr *I)`                 | Alias another registered type         |

Each returns `*Registration[T]` for chaining:

| Chain                                    | Description               |
| ---------------------------------------- | ------------------------- |
| `.Scope(ScopeSingleton\|ScopeTransient)` | Default: `ScopeSingleton` |
| `.Multi()`                               | Include in `ResolveAll`   |
| `.OnStart(func(T) error)`                | Lifecycle start hook      |
| `.OnStop(func(T) error)`                 | Lifecycle stop hook       |

### Resolution

```go
di.Resolve[T]() T                       // panics if not registered
di.ResolveNamed[T](name) T              // panics if not registered
di.TryResolve[T]() (T, bool)            // safe — returns false if missing
di.TryResolveNamed[T](name) (T, bool)   // safe — returns false if missing
di.ResolveAll[T]() []T                  // only Multi-marked entries
```

### Lifecycle

```go
di.StartAll() error
di.StartAllWithContext(ctx) error
di.StartAllWithTimeout(d) error
di.StopAll() error
di.StopAllWithContext(ctx) error
di.StopAllWithTimeout(d) error
```

`StartAll` runs `OnStart` hooks in registration order. On failure, already-started providers are rolled back. `StopAll` runs `OnStop` hooks in reverse order.

### Testing

```go
di.Reset() // clears all registrations and lifecycle state
```

---

## Patterns

### Singleton (default)

```go
di.Register[Logger](func(b di.Builder[Logger]) {
    b.New(func() (Logger, error) { return &StdLogger{}, nil })
})

log := di.Resolve[Logger]() // same instance every call
```

### Transient

```go
di.Register[*Request](func(b di.Builder[*Request]) {
    b.New(func() (*Request, error) { return &Request{}, nil }).
        Scope(di.ScopeTransient)
})

r1 := di.Resolve[*Request]() // new instance
r2 := di.Resolve[*Request]() // new instance
```

### Pre-built instance

```go
cfg := &Config{Addr: ":8080"}
di.Register[*Config](func(b di.Builder[*Config]) {
    b.Instance(cfg)
})
```

### Named + selector via env

Each implementation self-registers by name in `init()`. A selector promotes the configured one to the unnamed default:

```go
// nats/nats.go
func init() {
    di.Register[Broker](func(b di.Builder[Broker]) {
        b.Named("[broker/nats]", func() (Broker, error) {
            return NewNatsBroker(config), nil
        }).OnStart(func(br Broker) error { return br.Connect() }).
           OnStop(func(br Broker) error { return br.Close() })
    })
}

// broker/broker.go
func Register() {
    switch env.Get("BROKER_PROVIDER", "nats") {
    case "nats":
        di.Register[Broker](func(b di.Builder[Broker]) {
            b.New(func() (Broker, error) {
                return di.ResolveNamed[Broker]("[broker/nats]"), nil
            })
        })
    }
}

// usage
broker := di.Resolve[Broker]()
```

### Multi — capability aggregation

Register a type alias marked with `Multi()` to aggregate all implementations under a common interface:

```go
// nats/nats.go
func init() {
    di.Register[Broker](func(b di.Builder[Broker]) {
        b.Named("[broker/nats]", ctor).OnStart(...).OnStop(...)
    })
    di.Register[Connectable](func(b di.Builder[Connectable]) {
        var broker Broker
        b.Extend(&broker).Multi()
    })
}

// db/postgres.go
func init() {
    di.Register[Database](func(b di.Builder[Database]) {
        b.Named("[db/postgres]", ctor).OnStart(...).OnStop(...)
    })
    di.Register[Connectable](func(b di.Builder[Connectable]) {
        var db Database
        b.Extend(&db).Multi()
    })
}

// startup health check — iterates all Multi-marked Connectables
for _, c := range di.ResolveAll[Connectable]() {
    if err := c.Ping(); err != nil { ... }
}
```

### Optional dependency

```go
if cache, ok := di.TryResolve[Cache](); ok {
    cache.Set(key, val)
}
```

### Interface → concrete

```go
di.Register[port.Broker](func(b di.Builder[port.Broker]) {
    b.New(func() (port.Broker, error) { return &NatsBroker{}, nil })
})

// resolves as interface
broker := di.Resolve[port.Broker]()
```

---

## Examples

Runnable examples in [`examples/`](./examples/). All use `//go:build ignore` — run with `go run`:

### [basic.go](./examples/basic.go) — registration patterns

```go
// singleton via interface
di.Register[Logger](func(b di.Builder[Logger]) {
    b.New(func() (Logger, error) { return &StdLogger{}, nil })
})

// pre-built instance — Logger injected manually
di.Register[Repo](func(b di.Builder[Repo]) {
    b.Instance(&MemRepo{
        data: map[int]string{1: "Alice", 2: "Bob"},
        log:  di.Resolve[Logger](),
    })
})

// named transient variant
di.Register[Logger](func(b di.Builder[Logger]) {
    b.Named("stderr", func() (Logger, error) { return &StdLogger{}, nil }).
        Scope(di.ScopeTransient)
})

log    := di.Resolve[Logger]()
repo   := di.Resolve[Repo]()
stderr := di.ResolveNamed[Logger]("stderr")

if cache, ok := di.TryResolve[Repo](); ok {
    fmt.Println(cache.Find(2))
}
```

```sh
go run ./examples/basic.go
```

### [lifecycle.go](./examples/lifecycle.go) — OnStart/OnStop + graceful shutdown

```go
di.Register[DB](func(b di.Builder[DB]) {
    b.New(func() (DB, error) { return &MemDB{}, nil }).
        OnStart(func(d DB) error { return d.(*MemDB).Connect() }).
        OnStop(func(d DB) error { return d.(*MemDB).Close() })
})

di.Register[Cache](func(b di.Builder[Cache]) {
    b.New(func() (Cache, error) { return &MemCache{}, nil }).
        OnStart(func(c Cache) error { return c.(*MemCache).Connect() }).
        OnStop(func(c Cache) error { return c.(*MemCache).Close() })
})

di.StartAll()
// ... app runs ...
di.StopAll() // reverse order: Cache stops before DB
```

```sh
go run ./examples/lifecycle.go
```

### [multi.go](./examples/multi.go) — init() + selector + capability aggregation

```go
// nats/nats.go — self-registers by name
func init() {
    di.Register[Broker](func(b di.Builder[Broker]) {
        b.Named("[broker/nats]", ctor).OnStart(...).OnStop(...)
    })
    // opts in to Connectable aggregation
    di.Register[Connectable](func(b di.Builder[Connectable]) {
        var broker Broker
        b.Extend(&broker).Multi()
    })
}

// selector — promotes named to default
func registerBroker(provider string) {
    switch provider {
    case "nats":
        di.Register[Broker](func(b di.Builder[Broker]) {
            b.New(func() (Broker, error) {
                return di.ResolveNamed[Broker]("[broker/nats]"), nil
            })
        })
    }
}

// health check — only Multi-marked entries
for _, c := range di.ResolveAll[Connectable]() {
    c.Ping()
}
```

```sh
go run ./examples/multi.go
```

---

## Notes

- `Instance` ignores `.Scope()` — always singleton.
- `OnStart`/`OnStop` only apply to singleton providers (transients have no cached instance to stop).
- Circular singleton dependencies panic with a clear message.
- `Reset()` is intended for tests only — do not call in production code.
