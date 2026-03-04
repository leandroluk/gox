package di_test

import (
	"strings"
	"testing"

	"github.com/leandroluk/gox/di"
)

// --- Panic Test Types ---

type PanicInvalidShape interface {
	Volume() float64
}

type PanicConfig struct {
	Value int
}

// --- Panic Tests ---

func TestDI_Panics(t *testing.T) {
	t.Run("Panic on non-function factory", func(t *testing.T) {
		di.Reset()
		defer func() {
			if r := recover(); r == nil {
				t.Error("Should have panicked when registering a string instead of a function")
			}
		}()
		di.Register("not a function")
	})

	t.Run("Panic on multi-return factory", func(t *testing.T) {
		di.Reset()
		defer func() {
			if r := recover(); r == nil {
				t.Error("Should have panicked when factory returns more than one value")
			}
		}()
		multiReturnFactory := func() (*PanicConfig, error) { return &PanicConfig{}, nil }
		di.Register(multiReturnFactory)
	})

	t.Run("Panic on nil factory", func(t *testing.T) {
		di.Reset()
		defer func() {
			if r := recover(); r == nil {
				t.Error("Should have panicked when factory is nil")
			}
		}()
		di.Register(nil)
	})

	t.Run("Panic on type not assignable", func(t *testing.T) {
		di.Reset()
		defer func() {
			if r := recover(); r == nil {
				t.Error("Should have panicked when return type is not assignable to specified interface")
			}
		}()
		// BasicCircle não implementa PanicInvalidShape (que tem Volume())
		di.RegisterAs[PanicInvalidShape](NewBasicCircle)
	})

	t.Run("Panic on unresolved dependency", func(t *testing.T) {
		di.Reset()
		defer func() {
			if r := recover(); r == nil {
				t.Error("Should have panicked when trying to resolve unregistered type")
			}
		}()
		type UnregisteredType struct{}
		di.Resolve[UnregisteredType]()
	})
}

func TestDI_Panic_ImprovedErrorMessage_NoProvider(t *testing.T) {
	di.Reset()

	type UnregisteredService struct{}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Expected panic for unregistered type, got none")
		}

		errMsg := r.(string)

		// Should mention the missing type
		if !strings.Contains(errMsg, "UnregisteredService") {
			t.Errorf("Error should mention missing type, got: %s", errMsg)
		}

		// Should show dependency chain
		if !strings.Contains(errMsg, "dependency chain") {
			t.Errorf("Error should show dependency chain, got: %s", errMsg)
		}

		// Should provide helpful hint
		if !strings.Contains(errMsg, "hint:") {
			t.Errorf("Error should provide a hint, got: %s", errMsg)
		}

		if !strings.Contains(errMsg, "register") {
			t.Errorf("Hint should mention registering, got: %s", errMsg)
		}
	}()

	di.Resolve[*UnregisteredService]()
}

func TestDI_Panic_NestedMissingDependency(t *testing.T) {
	di.Reset()

	type PanicDatabase struct {
		Config *PanicConfig
	}
	type PanicService struct {
		DB *PanicDatabase
	}

	// Register Service and Database, but NOT Config
	di.Register(func(db *PanicDatabase) *PanicService {
		return &PanicService{DB: db}
	})
	di.Register(func(cfg *PanicConfig) *PanicDatabase {
		return &PanicDatabase{Config: cfg}
	})
	// PanicConfig is NOT registered!

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Expected panic for missing nested dependency, got none")
		}

		errMsg := r.(string)

		// Should show the full chain
		if !strings.Contains(errMsg, "dependency chain") {
			t.Errorf("Error should show full dependency chain, got: %s", errMsg)
		}

		// Should mention Config (the missing dependency)
		if !strings.Contains(errMsg, "PanicConfig") {
			t.Errorf("Error should mention missing PanicConfig, got: %s", errMsg)
		}

		// Ideally should show: Service -> Database -> Config
		// At minimum should show Config in the chain
		if !strings.Contains(errMsg, "->") {
			t.Errorf("Chain should show arrow notation, got: %s", errMsg)
		}
	}()

	di.Resolve[*PanicService]()
}
