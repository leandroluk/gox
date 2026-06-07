package di_v1_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/leandroluk/gox/di"
)

type PanicInvalidShape interface {
	Volume() float64
}

type PanicConfig struct {
	Value int
}

func TestDI_Panics(t *testing.T) {
	t.Run("Panic on nil constructor", func(t *testing.T) {
		di.Reset()
		defer func() {
			if r := recover(); r == nil {
				t.Error("Should have panicked when constructor is nil")
			}
		}()
		di.Register[*PanicConfig](func(o *di.Options[*PanicConfig]) {
			// Constructor is nil
		})
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

		if !strings.Contains(errMsg, "UnregisteredService") {
			t.Errorf("Error should mention missing type, got: %s", errMsg)
		}

		if !strings.Contains(errMsg, "dependency chain") {
			t.Errorf("Error should show dependency chain, got: %s", errMsg)
		}

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

	di.RegisterFrom[*PanicService](func() (*PanicService, error) {
		db := di.Resolve[*PanicDatabase]()
		return &PanicService{DB: db}, nil
	})
	di.RegisterFrom[*PanicDatabase](func() (*PanicDatabase, error) {
		cfg := di.Resolve[*PanicConfig]()
		return &PanicDatabase{Config: cfg}, nil
	})

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Expected panic for missing nested dependency, got none")
		}

		errMsg := r.(string)

		if !strings.Contains(errMsg, "dependency chain") {
			t.Errorf("Error should show dependency chain, got: %s", errMsg)
		}

		if !strings.Contains(errMsg, "PanicConfig") {
			t.Errorf("Error should mention missing PanicConfig, got: %s", errMsg)
		}
	}()

	di.Resolve[*PanicService]()
}

func TestDI_Panic_FactoryError(t *testing.T) {
	di.Reset()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Should panic when factory returns error")
		}

		errMsg := r.(string)
		if !strings.Contains(errMsg, "factory function returned error") {
			t.Errorf("Error should mention factory error, got: %s", errMsg)
		}
	}()

	di.RegisterFrom[*PanicConfig](func() (*PanicConfig, error) {
		return nil, errors.New("config load failed")
	})

	di.Resolve[*PanicConfig]()
}
