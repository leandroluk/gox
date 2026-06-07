package di_v1_test

import (
	"testing"

	"github.com/leandroluk/gox/di"
)

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

func TestDI_RegisterAndResolve(t *testing.T) {
	di.Reset()

	di.Register[*BasicConfig](func(o *di.Options[*BasicConfig]) {
		o.Constructor = func() (*BasicConfig, error) {
			return &BasicConfig{Factor: 10}, nil
		}
	})

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

	di.Singleton[*BasicConfig](func(o *di.Options[*BasicConfig]) {
		o.Constructor = func() (*BasicConfig, error) {
			return &BasicConfig{Factor: 10}, nil
		}
	})

	inst1 := di.Resolve[*BasicConfig]()
	inst2 := di.Resolve[*BasicConfig]()

	if inst1 != inst2 {
		t.Error("Singleton failed: instances are different, memory addresses do not match")
	}
}

func TestDI_Transient(t *testing.T) {
	di.Reset()

	di.Register[*BasicConfig](func(o *di.Options[*BasicConfig]) {
		o.Constructor = func() (*BasicConfig, error) {
			return &BasicConfig{Factor: 10}, nil
		}
	})

	inst1 := di.Resolve[*BasicConfig]()
	inst2 := di.Resolve[*BasicConfig]()

	if inst1 == inst2 {
		t.Error("Transient failed: instances should have different memory addresses")
	}
}

func TestDI_RegisterAs(t *testing.T) {
	di.Reset()

	di.RegisterAs[BasicShape](func(o *di.Options[BasicShape]) {
		o.Constructor = func() (BasicShape, error) {
			return &BasicCircle{Radius: 5}, nil
		}
	})

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

	di.SingletonAs[BasicShape](func(o *di.Options[BasicShape]) {
		o.Constructor = func() (BasicShape, error) {
			return &BasicCircle{Radius: 5}, nil
		}
	})

	shape1 := di.Resolve[BasicShape]()
	shape2 := di.Resolve[BasicShape]()

	circle1 := shape1.(*BasicCircle)
	circle2 := shape2.(*BasicCircle)

	if circle1 != circle2 {
		t.Error("SingletonAs failed: instances should be the same")
	}
}

func TestDI_SingletonInstance(t *testing.T) {
	di.Reset()

	specificConfig := &BasicConfig{Factor: 42}
	di.SingletonInstance(specificConfig, nil)

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

	di.Singleton[*BasicConfig](func(o *di.Options[*BasicConfig]) {
		o.Constructor = func() (*BasicConfig, error) {
			return &BasicConfig{Factor: 10}, nil
		}
	})

	di.RegisterFrom[*BasicCalculator](func() (*BasicCalculator, error) {
		cfg := di.Resolve[*BasicConfig]()
		return &BasicCalculator{Config: cfg}, nil
	})

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

	di.RegisterNamed[BasicShape]("circle1", func(o *di.Options[BasicShape]) {
		o.Constructor = func() (BasicShape, error) { return &BasicCircle{Radius: 1}, nil }
	})
	di.RegisterNamed[BasicShape]("circle2", func(o *di.Options[BasicShape]) {
		o.Constructor = func() (BasicShape, error) { return &BasicCircle{Radius: 2}, nil }
	})
	di.RegisterNamed[BasicShape]("rectangle", func(o *di.Options[BasicShape]) {
		o.Constructor = func() (BasicShape, error) { return &BasicRectangle{Width: 3, Height: 4}, nil }
	})

	shapes := di.ResolveAll[BasicShape]()

	if len(shapes) != 3 {
		t.Errorf("Expected 3 registered shapes, got %d", len(shapes))
	}
}

func TestDI_ResolveAll_NoProviders(t *testing.T) {
	di.Reset()

	type UnregisteredType struct{}

	results := di.ResolveAll[UnregisteredType]()

	if results != nil {
		t.Errorf("Expected nil for unregistered type, got %v", results)
	}
}

func TestDI_Reset(t *testing.T) {
	di.Register[*BasicConfig](func(o *di.Options[*BasicConfig]) {
		o.Constructor = func() (*BasicConfig, error) { return &BasicConfig{}, nil }
	})
	di.Register[*BasicCircle](func(o *di.Options[*BasicCircle]) {
		o.Constructor = func() (*BasicCircle, error) { return &BasicCircle{}, nil }
	})

	di.Reset()

	di.RegistryMutex.RLock()
	count := len(di.ProviderRegistry)
	di.RegistryMutex.RUnlock()

	if count != 0 {
		t.Errorf("Reset failed: expected 0 providers, got %d", count)
	}
}
