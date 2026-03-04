package di_test

import (
	"reflect"
	"testing"

	"github.com/leandroluk/gox/di"
)

// --- Mocks and Types for Testing ---

type Shape interface {
	Area() float64
}

type Circle struct {
	Radius float64
}

func (c *Circle) Area() float64 {
	return 3.14 * c.Radius * c.Radius
}

type Rectangle struct {
	Width  float64
	Height float64
}

func (r *Rectangle) Area() float64 {
	return r.Width * r.Height
}

type Calculator struct {
	Config *Config
}

type Config struct {
	Factor int
}

// Interface incompatível para teste de assignability
type InvalidShape interface {
	Volume() float64
}

// --- Factory Functions ---

func NewConfig() *Config {
	return &Config{Factor: 10}
}

func NewCircle() *Circle {
	return &Circle{Radius: 5}
}

func NewRectangle() *Rectangle {
	return &Rectangle{Width: 10, Height: 5}
}

func NewCalculator(cfg *Config) *Calculator {
	return &Calculator{Config: cfg}
}

// --- Helper: Internal Reset ---
func resetRegistry() {
	di.RegistryMutex.Lock()
	defer di.RegistryMutex.Unlock()
	di.ProviderRegistry = make(map[reflect.Type][]*di.Provider)
}

// --- Test Cases ---

func TestDI_RegisterAndResolve(t *testing.T) {
	resetRegistry()

	di.Register(NewConfig)

	cfg := di.Resolve[*Config]()
	if cfg == nil {
		t.Fatal("Expected *Config, got nil")
	}
	if cfg.Factor != 10 {
		t.Errorf("Expected factor 10, got %d", cfg.Factor)
	}
}

func TestDI_Singleton(t *testing.T) {
	resetRegistry()

	di.Singleton(NewConfig)

	inst1 := di.Resolve[*Config]()
	inst2 := di.Resolve[*Config]()

	// Compara endereços de memória para garantir que é a mesma instância
	if inst1 != inst2 {
		t.Error("Singleton failed: instances are different, memory addresses do not match")
	}
}

func TestDI_Transient(t *testing.T) {
	resetRegistry()

	di.Register(NewConfig)

	inst1 := di.Resolve[*Config]()
	inst2 := di.Resolve[*Config]()

	// Devem ser instâncias diferentes (endereços diferentes)
	if inst1 == inst2 {
		t.Error("Transient failed: instances should have different memory addresses")
	}
}

func TestDI_RegisterAs(t *testing.T) {
	resetRegistry()

	// Registers *Circle bound to Shape interface
	di.RegisterAs[Shape](NewCircle)

	shape := di.Resolve[Shape]()
	if shape == nil {
		t.Fatal("Failed to resolve Shape interface")
	}

	expectedArea := 78.5
	if shape.Area() != expectedArea {
		t.Errorf("Expected area %f, got %f", expectedArea, shape.Area())
	}
}

func TestDI_SingletonAs(t *testing.T) {
	resetRegistry()

	di.SingletonAs[Shape](NewCircle)

	shape1 := di.Resolve[Shape]()
	shape2 := di.Resolve[Shape]()

	// Verifica se são a mesma instância
	circle1 := shape1.(*Circle)
	circle2 := shape2.(*Circle)

	if circle1 != circle2 {
		t.Error("SingletonAs failed: instances should be the same")
	}
}

func TestDI_SingletonInstance(t *testing.T) {
	resetRegistry()

	// Cria uma instância específica
	specificConfig := &Config{Factor: 42}
	di.SingletonInstance(specificConfig)

	// Resolve e verifica se é exatamente a mesma instância
	resolved := di.Resolve[*Config]()

	if resolved != specificConfig {
		t.Error("SingletonInstance failed: should return the exact same instance")
	}

	if resolved.Factor != 42 {
		t.Errorf("Expected factor 42, got %d", resolved.Factor)
	}
}

func TestDI_NestedDependencies(t *testing.T) {
	resetRegistry()

	// Register dependencies
	di.Singleton(NewConfig)
	di.Register(NewCalculator) // NewCalculator depends on *Config

	calc := di.Resolve[*Calculator]()

	if calc.Config == nil {
		t.Fatal("Dependency *Config was not automatically injected into Calculator")
	}

	if calc.Config.Factor != 10 {
		t.Errorf("Injected *Config has incorrect value: %d", calc.Config.Factor)
	}
}

func TestDI_ResolveAll(t *testing.T) {
	resetRegistry()

	// Register multiple shapes
	di.RegisterAs[Shape](func() *Circle { return &Circle{Radius: 1} })
	di.RegisterAs[Shape](func() *Circle { return &Circle{Radius: 2} })
	di.RegisterAs[Shape](func() *Rectangle { return &Rectangle{Width: 3, Height: 4} })

	shapes := di.ResolveAll[Shape]()

	if len(shapes) != 3 {
		t.Errorf("Expected 3 registered shapes, got %d", len(shapes))
	}
}

func TestDI_ResolveAll_NoProviders(t *testing.T) {
	resetRegistry()

	// Tenta resolver um tipo que não foi registrado
	type UnregisteredType struct{}

	results := di.ResolveAll[UnregisteredType]()

	if results != nil {
		t.Errorf("Expected nil for unregistered type, got %v", results)
	}
}

func TestDI_Panics(t *testing.T) {
	t.Run("Panic on non-function factory", func(t *testing.T) {
		resetRegistry()
		defer func() {
			if r := recover(); r == nil {
				t.Error("Should have panicked when registering a string instead of a function")
			}
		}()
		di.Register("not a function")
	})

	t.Run("Panic on multi-return factory", func(t *testing.T) {
		resetRegistry()
		defer func() {
			if r := recover(); r == nil {
				t.Error("Should have panicked when factory returns more than one value")
			}
		}()
		multiReturnFactory := func() (*Config, error) { return &Config{}, nil }
		di.Register(multiReturnFactory)
	})

	t.Run("Panic on nil factory", func(t *testing.T) {
		resetRegistry()
		defer func() {
			if r := recover(); r == nil {
				t.Error("Should have panicked when factory is nil")
			}
		}()
		di.Register(nil)
	})

	t.Run("Panic on type not assignable", func(t *testing.T) {
		resetRegistry()
		defer func() {
			if r := recover(); r == nil {
				t.Error("Should have panicked when return type is not assignable to specified interface")
			}
		}()
		// Circle não implementa InvalidShape (que tem Volume())
		di.RegisterAs[InvalidShape](NewCircle)
	})

	t.Run("Panic on unresolved dependency", func(t *testing.T) {
		resetRegistry()
		defer func() {
			if r := recover(); r == nil {
				t.Error("Should have panicked when trying to resolve unregistered type")
			}
		}()
		type UnregisteredType struct{}
		di.Resolve[UnregisteredType]()
	})
}

func TestDI_Reset(t *testing.T) {
	// Registra alguns providers
	di.Register(NewConfig)
	di.Register(NewCircle)

	// Usa Reset explicitamente
	di.Reset()

	// Verifica que o registry está vazio
	di.RegistryMutex.RLock()
	count := len(di.ProviderRegistry)
	di.RegistryMutex.RUnlock()

	if count != 0 {
		t.Errorf("Reset failed: expected 0 providers, got %d", count)
	}
}
