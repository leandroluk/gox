package di_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/leandroluk/gox/di"
)

// --- helpers ---

type counter struct{ n int }

type greeter interface{ Greet() string }
type greeterImpl struct{ msg string }

func (g *greeterImpl) Greet() string { return g.msg }

type connectable interface {
	Connect() error
	Close() error
}

type service interface{ connectable }

type serviceImpl struct{ connected bool }

func (s *serviceImpl) Connect() error { s.connected = true; return nil }
func (s *serviceImpl) Close() error   { s.connected = false; return nil }

// --- singleton ---

func TestSingleton_SameInstance(t *testing.T) {
	defer di.Reset()
	di.Register[*counter](func(b di.Builder[*counter]) {
		b.New(func() (*counter, error) { return &counter{}, nil })
	})
	a := di.Resolve[*counter]()
	b := di.Resolve[*counter]()
	if a != b {
		t.Fatal("singleton: expected same pointer")
	}
}

func TestSingleton_DefaultScope(t *testing.T) {
	defer di.Reset()
	di.Register[*counter](func(b di.Builder[*counter]) {
		b.New(func() (*counter, error) { return &counter{}, nil })
	})
	a := di.Resolve[*counter]()
	a.n = 42
	b := di.Resolve[*counter]()
	if b.n != 42 {
		t.Fatal("singleton: mutation not shared")
	}
}

// --- transient ---

func TestTransient_NewInstanceEachTime(t *testing.T) {
	defer di.Reset()
	di.Register[*counter](func(b di.Builder[*counter]) {
		b.New(func() (*counter, error) { return &counter{}, nil }).
			Scope(di.ScopeTransient)
	})
	a := di.Resolve[*counter]()
	b := di.Resolve[*counter]()
	if a == b {
		t.Fatal("transient: expected different pointers")
	}
}

// --- named ---

func TestNamed_ResolvesCorrectImpl(t *testing.T) {
	defer di.Reset()
	di.Register[greeter](func(b di.Builder[greeter]) {
		b.Named("hello", func() (greeter, error) { return &greeterImpl{msg: "hello"}, nil })
		b.Named("world", func() (greeter, error) { return &greeterImpl{msg: "world"}, nil })
	})
	if got := di.ResolveNamed[greeter]("hello").Greet(); got != "hello" {
		t.Fatalf("named hello: got %q", got)
	}
	if got := di.ResolveNamed[greeter]("world").Greet(); got != "world" {
		t.Fatalf("named world: got %q", got)
	}
}

func TestNamed_Singleton(t *testing.T) {
	defer di.Reset()
	di.Register[*counter](func(b di.Builder[*counter]) {
		b.Named("x", func() (*counter, error) { return &counter{}, nil })
	})
	a := di.ResolveNamed[*counter]("x")
	b := di.ResolveNamed[*counter]("x")
	if a != b {
		t.Fatal("named singleton: expected same pointer")
	}
}

// --- instance ---

func TestInstance_ReturnsSameValue(t *testing.T) {
	defer di.Reset()
	val := &counter{n: 7}
	di.Register[*counter](func(b di.Builder[*counter]) {
		b.Instance(val)
	})
	got := di.Resolve[*counter]()
	if got != val {
		t.Fatal("instance: expected registered pointer")
	}
}

func TestInstance_ScopeLockedToSingleton(t *testing.T) {
	defer di.Reset()
	val := &counter{n: 1}
	di.Register[*counter](func(b di.Builder[*counter]) {
		b.Instance(val).Scope(di.ScopeTransient) // scope override ignored
	})
	a := di.Resolve[*counter]()
	b := di.Resolve[*counter]()
	if a != b {
		t.Fatal("instance: Scope override should be ignored")
	}
}

// --- extend + multi ---

func TestExtend_ResolvesSourceType(t *testing.T) {
	defer di.Reset()
	di.Register[service](func(b di.Builder[service]) {
		b.New(func() (service, error) { return &serviceImpl{}, nil })
	})
	di.Register[connectable](func(b di.Builder[connectable]) {
		var svc service
		b.Extend(&svc)
	})
	svc := di.Resolve[service]()
	conn := di.Resolve[connectable]()
	if svc != conn {
		t.Fatal("extend: expected same underlying instance")
	}
}

func TestMulti_ResolveAllOnlyReturnsMarked(t *testing.T) {
	defer di.Reset()
	di.Register[service](func(b di.Builder[service]) {
		b.New(func() (service, error) { return &serviceImpl{}, nil })
	})
	// connectable extends service with Multi
	di.Register[connectable](func(b di.Builder[connectable]) {
		var svc service
		b.Extend(&svc).Multi()
	})
	// second connectable via named, not Multi
	di.Register[connectable](func(b di.Builder[connectable]) {
		b.Named("extra", func() (connectable, error) { return &serviceImpl{}, nil })
	})

	all := di.ResolveAll[connectable]()
	if len(all) != 1 {
		t.Fatalf("ResolveAll: expected 1 Multi entry, got %d", len(all))
	}
}

func TestMulti_EmptyWhenNoneMarked(t *testing.T) {
	defer di.Reset()
	di.Register[*counter](func(b di.Builder[*counter]) {
		b.New(func() (*counter, error) { return &counter{}, nil })
	})
	all := di.ResolveAll[*counter]()
	if len(all) != 0 {
		t.Fatalf("ResolveAll: expected 0, got %d", len(all))
	}
}

// --- TryResolve ---

func TestTryResolve_Found(t *testing.T) {
	defer di.Reset()
	di.Register[*counter](func(b di.Builder[*counter]) {
		b.New(func() (*counter, error) { return &counter{n: 3}, nil })
	})
	v, ok := di.TryResolve[*counter]()
	if !ok || v == nil || v.n != 3 {
		t.Fatalf("TryResolve: expected found, got ok=%v v=%v", ok, v)
	}
}

func TestTryResolve_NotFound(t *testing.T) {
	defer di.Reset()
	v, ok := di.TryResolve[*counter]()
	if ok || v != nil {
		t.Fatalf("TryResolve: expected not found, got ok=%v v=%v", ok, v)
	}
}

func TestTryResolveNamed_Found(t *testing.T) {
	defer di.Reset()
	di.Register[greeter](func(b di.Builder[greeter]) {
		b.Named("hi", func() (greeter, error) { return &greeterImpl{msg: "hi"}, nil })
	})
	v, ok := di.TryResolveNamed[greeter]("hi")
	if !ok || v == nil {
		t.Fatalf("TryResolveNamed: expected found, got ok=%v", ok)
	}
}

func TestTryResolveNamed_NotFound(t *testing.T) {
	defer di.Reset()
	v, ok := di.TryResolveNamed[greeter]("missing")
	if ok || v != nil {
		t.Fatalf("TryResolveNamed: expected not found, got ok=%v", ok)
	}
}

// --- interface resolution ---

func TestInterfaceResolution(t *testing.T) {
	defer di.Reset()
	di.Register[greeter](func(b di.Builder[greeter]) {
		b.New(func() (greeter, error) { return &greeterImpl{msg: "ok"}, nil })
	})
	g := di.Resolve[greeter]()
	if g.Greet() != "ok" {
		t.Fatalf("interface: got %q", g.Greet())
	}
}

// --- lifecycle ---

func TestLifecycle_OnStartOnStop(t *testing.T) {
	defer di.Reset()
	var log []string
	di.Register[*serviceImpl](func(b di.Builder[*serviceImpl]) {
		b.New(func() (*serviceImpl, error) { return &serviceImpl{}, nil }).
			OnStart(func(s *serviceImpl) error { log = append(log, "start"); return s.Connect() }).
			OnStop(func(s *serviceImpl) error { log = append(log, "stop"); return s.Close() })
	})
	if err := di.StartAll(); err != nil {
		t.Fatal(err)
	}
	if err := di.StopAll(); err != nil {
		t.Fatal(err)
	}
	if len(log) != 2 || log[0] != "start" || log[1] != "stop" {
		t.Fatalf("lifecycle: unexpected log %v", log)
	}
}

func TestLifecycle_StopReverseOrder(t *testing.T) {
	defer di.Reset()
	var log []string
	for _, name := range []string{"a", "b", "c"} {
		n := name
		di.Register[greeter](func(b di.Builder[greeter]) {
			b.Named(n, func() (greeter, error) { return &greeterImpl{msg: n}, nil }).
				OnStart(func(g greeter) error { log = append(log, "start:"+n); return nil }).
				OnStop(func(g greeter) error { log = append(log, "stop:"+n); return nil })
		})
	}
	if err := di.StartAll(); err != nil {
		t.Fatal(err)
	}
	if err := di.StopAll(); err != nil {
		t.Fatal(err)
	}
	want := []string{"start:a", "start:b", "start:c", "stop:c", "stop:b", "stop:a"}
	for i, w := range want {
		if log[i] != w {
			t.Fatalf("order: log[%d]=%q want %q", i, log[i], w)
		}
	}
}

func TestLifecycle_StartFail_Rollback(t *testing.T) {
	defer di.Reset()
	var stopped []string
	di.Register[greeter](func(b di.Builder[greeter]) {
		b.Named("a", func() (greeter, error) { return &greeterImpl{msg: "a"}, nil }).
			OnStart(func(g greeter) error { return nil }).
			OnStop(func(g greeter) error { stopped = append(stopped, "a"); return nil })
	})
	di.Register[greeter](func(b di.Builder[greeter]) {
		b.Named("b", func() (greeter, error) { return &greeterImpl{msg: "b"}, nil }).
			OnStart(func(g greeter) error { return errors.New("b failed") }).
			OnStop(func(g greeter) error { stopped = append(stopped, "b"); return nil })
	})
	err := di.StartAll()
	if err == nil {
		t.Fatal("expected start error")
	}
	if len(stopped) != 1 || stopped[0] != "a" {
		t.Fatalf("rollback: expected [a], got %v", stopped)
	}
}

func TestLifecycle_OnlyStartedOncePer_StartAll(t *testing.T) {
	defer di.Reset()
	starts := 0
	di.Register[*counter](func(b di.Builder[*counter]) {
		b.New(func() (*counter, error) { return &counter{}, nil }).
			OnStart(func(c *counter) error { starts++; return nil })
	})
	_ = di.StartAll()
	_ = di.StartAll()
	if starts != 1 {
		t.Fatalf("expected 1 start, got %d", starts)
	}
}

// --- init + selector pattern ---

func TestInitSelectorPattern(t *testing.T) {
	defer di.Reset()

	// simulates init() in nats package
	di.Register[service](func(b di.Builder[service]) {
		b.Named("[svc/nats]", func() (service, error) { return &serviceImpl{}, nil })
	})

	// simulates selector based on env
	provider := "nats"
	switch provider {
	case "nats":
		di.Register[service](func(b di.Builder[service]) {
			b.New(func() (service, error) {
				return di.ResolveNamed[service]("[svc/nats]"), nil
			})
		})
	}

	svc := di.Resolve[service]()
	named := di.ResolveNamed[service]("[svc/nats]")
	if svc != named {
		t.Fatal("selector pattern: default should resolve to named impl")
	}
}

// --- reset ---

func TestReset_ClearsAll(t *testing.T) {
	di.Register[*counter](func(b di.Builder[*counter]) {
		b.New(func() (*counter, error) { return &counter{}, nil })
	})
	di.Reset()
	_, ok := di.TryResolve[*counter]()
	if ok {
		t.Fatal("Reset: registry not cleared")
	}
}

// --- panic cases ---

func TestPanic_NotRegistered(t *testing.T) {
	defer di.Reset()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for unregistered type")
		}
	}()
	di.Resolve[*counter]()
}

func TestPanic_NamedNotRegistered(t *testing.T) {
	defer di.Reset()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for unregistered named type")
		}
	}()
	di.ResolveNamed[greeter]("missing")
}

func TestPanic_DuplicateUnnamed(t *testing.T) {
	defer di.Reset()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on duplicate unnamed registration")
		}
	}()
	di.Register[*counter](func(b di.Builder[*counter]) {
		b.New(func() (*counter, error) { return &counter{}, nil })
	})
	di.Register[*counter](func(b di.Builder[*counter]) {
		b.New(func() (*counter, error) { return &counter{}, nil })
	})
}

func TestPanic_DuplicateNamed(t *testing.T) {
	defer di.Reset()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on duplicate named registration")
		}
	}()
	di.Register[greeter](func(b di.Builder[greeter]) {
		b.Named("dup", func() (greeter, error) { return &greeterImpl{}, nil })
		b.Named("dup", func() (greeter, error) { return &greeterImpl{}, nil })
	})
}

func TestPanic_NamedEmptyString(t *testing.T) {
	defer di.Reset()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for empty name")
		}
	}()
	di.Register[greeter](func(b di.Builder[greeter]) {
		b.Named("", func() (greeter, error) { return &greeterImpl{}, nil })
	})
}

func TestPanic_FactoryError(t *testing.T) {
	defer di.Reset()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on factory error")
		}
	}()
	di.Register[*counter](func(b di.Builder[*counter]) {
		b.New(func() (*counter, error) { return nil, fmt.Errorf("boom") })
	})
	di.Resolve[*counter]()
}

func TestPanic_CircularDependency(t *testing.T) {
	defer di.Reset()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for circular dependency")
		}
	}()
	di.Register[*counter](func(b di.Builder[*counter]) {
		b.New(func() (*counter, error) {
			di.Resolve[*counter]() // circular
			return &counter{}, nil
		})
	})
	di.Resolve[*counter]()
}

func TestPanic_ExtendNonPointer(t *testing.T) {
	defer di.Reset()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for non-pointer Extend")
		}
	}()
	di.Register[connectable](func(b di.Builder[connectable]) {
		b.Extend("not a pointer")
	})
}
