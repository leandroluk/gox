package di_test

import (
	"testing"

	"github.com/leandroluk/gox/di"
)

// --- Basic Test Types (v1.0) ---

type BasicShape interface {
	Area() float64
}

type BasicCircle struct {
	Radius float64
}

func (c *BasicCircle) Area() float64 {
	return 3.14 * c.Radius * c.Radius
}

type BasicRectangle struct {
	Width  float64
	Height float64
}

func (r *BasicRectangle) Area() float64 {
	return r.Width * r.Height
}

type BasicCalculator struct {
	Config *BasicConfig
}

type BasicConfig struct {
	Factor int
}

// --- Factory Functions ---

func NewBasicConfig() *BasicConfig {
	return &BasicConfig{Factor: 10}
}

func NewBasicCircle() *BasicCircle {
	return &BasicCircle{Radius: 5}
}

func NewBasicRectangle() *BasicRectangle {
	return &BasicRectangle{Width: 10, Height: 5}
}

func NewBasicCalculator(cfg *BasicConfig) *BasicCalculator {
	return &BasicCalculator{Config: cfg}
}

// --- Core Functionality Tests ---

func TestDI_RegisterAndResolve(t *testing.T) {
	di.Reset()

	di.Register(NewBasicConfig)

	cfg := di.Resolve[*BasicConfig]()
	if cfg == nil {
		t.Fatal("Expected *BasicConfig, got nil")
	}
	if cfg.Factor != 10 {
		t.Errorf("Expected factor 10, got %d", cfg.Factor)
	}
}

func TestDI_Singleton(t *testing.T) {
	di.Reset()

	di.Singleton(NewBasicConfig)

	inst1 := di.Resolve[*BasicConfig]()
	inst2 := di.Resolve[*BasicConfig]()

	// Compara endereços de memória para garantir que é a mesma instância
	if inst1 != inst2 {
		t.Error("Singleton failed: instances are different, memory addresses do not match")
	}
}

func TestDI_Transient(t *testing.T) {
	di.Reset()

	di.Register(NewBasicConfig)

	inst1 := di.Resolve[*BasicConfig]()
	inst2 := di.Resolve[*BasicConfig]()

	// Devem ser instâncias diferentes (endereços diferentes)
	if inst1 == inst2 {
		t.Error("Transient failed: instances should have different memory addresses")
	}
}

func TestDI_RegisterAs(t *testing.T) {
	di.Reset()

	// Registers *BasicCircle bound to BasicShape interface
	di.RegisterAs[BasicShape](NewBasicCircle)

	shape := di.Resolve[BasicShape]()
	if shape == nil {
		t.Fatal("Failed to resolve BasicShape interface")
	}

	expectedArea := 78.5
	if shape.Area() != expectedArea {
		t.Errorf("Expected area %f, got %f", expectedArea, shape.Area())
	}
}

func TestDI_SingletonAs(t *testing.T) {
	di.Reset()

	di.SingletonAs[BasicShape](NewBasicCircle)

	shape1 := di.Resolve[BasicShape]()
	shape2 := di.Resolve[BasicShape]()

	// Verifica se são a mesma instância
	circle1 := shape1.(*BasicCircle)
	circle2 := shape2.(*BasicCircle)

	if circle1 != circle2 {
		t.Error("SingletonAs failed: instances should be the same")
	}
}

func TestDI_SingletonInstance(t *testing.T) {
	di.Reset()

	// Cria uma instância específica
	specificConfig := &BasicConfig{Factor: 42}
	di.SingletonInstance(specificConfig)

	// Resolve e verifica se é exatamente a mesma instância
	resolved := di.Resolve[*BasicConfig]()

	if resolved != specificConfig {
		t.Error("SingletonInstance failed: should return the exact same instance")
	}

	if resolved.Factor != 42 {
		t.Errorf("Expected factor 42, got %d", resolved.Factor)
	}
}

func TestDI_NestedDependencies(t *testing.T) {
	di.Reset()

	// Register dependencies
	di.Singleton(NewBasicConfig)
	di.Register(NewBasicCalculator) // NewBasicCalculator depends on *BasicConfig

	calc := di.Resolve[*BasicCalculator]()

	if calc.Config == nil {
		t.Fatal("Dependency *BasicConfig was not automatically injected into BasicCalculator")
	}

	if calc.Config.Factor != 10 {
		t.Errorf("Injected *BasicConfig has incorrect value: %d", calc.Config.Factor)
	}
}

func TestDI_ResolveAll(t *testing.T) {
	di.Reset()

	// Register multiple shapes using named instances (v1.3)
	// Note: In v1.3, you cannot register multiple unnamed providers of the same type
	di.RegisterNamed[BasicShape]("circle1", func() *BasicCircle { return &BasicCircle{Radius: 1} })
	di.RegisterNamed[BasicShape]("circle2", func() *BasicCircle { return &BasicCircle{Radius: 2} })
	di.RegisterNamed[BasicShape]("rectangle", func() *BasicRectangle { return &BasicRectangle{Width: 3, Height: 4} })

	shapes := di.ResolveAll[BasicShape]()

	if len(shapes) != 3 {
		t.Errorf("Expected 3 registered shapes, got %d", len(shapes))
	}
}

func TestDI_ResolveAll_NoProviders(t *testing.T) {
	di.Reset()

	// Tenta resolver um tipo que não foi registrado
	type UnregisteredType struct{}

	results := di.ResolveAll[UnregisteredType]()

	if results != nil {
		t.Errorf("Expected nil for unregistered type, got %v", results)
	}
}

func TestDI_Reset(t *testing.T) {
	// Registra alguns providers
	di.Register(NewBasicConfig)
	di.Register(NewBasicCircle)

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
