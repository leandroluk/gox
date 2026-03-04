package di

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

// Helper para capturar stdout
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestDebugMode_Enabled(t *testing.T) {
	// Salva estado anterior
	oldDebugMode := debugMode
	defer func() { debugMode = oldDebugMode }()

	// Habilita debug mode
	Debug()

	if !debugMode {
		t.Error("Debug() did not set debugMode to true")
	}

	// Testa logDebug com debug ativado
	output := captureOutput(func() {
		logDebug("Test message: %s", "value")
	})

	if !strings.Contains(output, "[DI] Test message: value") {
		t.Errorf("Expected debug output, got: %q", output)
	}
}

func TestDebugMode_Disabled(t *testing.T) {
	// Salva estado anterior
	oldDebugMode := debugMode
	defer func() { debugMode = oldDebugMode }()

	// Garante que debug está desabilitado
	debugMode = false

	// Testa logDebug com debug desativado (não deve imprimir nada)
	output := captureOutput(func() {
		logDebug("This should not appear")
	})

	if output != "" {
		t.Errorf("Expected no output with debug disabled, got: %q", output)
	}
}

func TestFail_WithDebug(t *testing.T) {
	// Salva estado anterior
	oldDebugMode := debugMode
	defer func() { debugMode = oldDebugMode }()

	debugMode = true

	// Captura panic e output
	output := captureOutput(func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("fail() should have panicked")
			}
		}()
		fail("test error message")
	})

	if !strings.Contains(output, "[DI] ERROR: test error message") {
		t.Errorf("Expected error message in output, got: %q", output)
	}
}

func TestFail_WithoutDebug(t *testing.T) {
	// Salva estado anterior
	oldDebugMode := debugMode
	defer func() { debugMode = oldDebugMode }()

	debugMode = false

	// Captura panic sem output de debug
	output := captureOutput(func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("fail() should have panicked")
			} else {
				// Verifica que a mensagem do panic está correta
				if r != "test error message" {
					t.Errorf("Expected panic message 'test error message', got: %v", r)
				}
			}
		}()
		fail("test error message")
	})

	// Não deve ter output quando debug está desabilitado
	if output != "" {
		t.Errorf("Expected no output with debug disabled, got: %q", output)
	}
}

func TestDebugMode_Integration(t *testing.T) {
	// Salva estado anterior
	oldDebugMode := debugMode
	defer func() { debugMode = oldDebugMode }()

	// Reseta registry
	Reset()

	// Habilita debug
	Debug()

	// Captura output durante registro e resolução
	output := captureOutput(func() {
		RegisterAs[string](func() string {
			return "test value"
		})

		val := Resolve[string]()
		if val != "test value" {
			t.Errorf("Expected 'test value', got '%s'", val)
		}
	})

	// Verifica que mensagens de debug foram impressas
	if !strings.Contains(output, "[DI]") {
		t.Errorf("Expected debug output during registration/resolution, got: %q", output)
	}

	if !strings.Contains(output, "Registering") {
		t.Error("Expected 'Registering' in debug output")
	}

	if !strings.Contains(output, "Resolving") {
		t.Error("Expected 'Resolving' in debug output")
	}
}

func TestLogDebug_MultipleArgs(t *testing.T) {
	// Salva estado anterior
	oldDebugMode := debugMode
	defer func() { debugMode = oldDebugMode }()

	debugMode = true

	output := captureOutput(func() {
		logDebug("Values: %d, %s, %v", 42, "test", true)
	})

	expected := "[DI] Values: 42, test, true"
	if !strings.Contains(output, expected) {
		t.Errorf("Expected %q in output, got: %q", expected, output)
	}
}

func TestLogDebug_NoArgs(t *testing.T) {
	// Salva estado anterior
	oldDebugMode := debugMode
	defer func() { debugMode = oldDebugMode }()

	debugMode = true

	output := captureOutput(func() {
		logDebug("Simple message")
	})

	if !strings.Contains(output, "[DI] Simple message") {
		t.Errorf("Expected simple message in output, got: %q", output)
	}
}

func TestDebug_MultipleCalls(t *testing.T) {
	// Salva estado anterior
	oldDebugMode := debugMode
	defer func() { debugMode = oldDebugMode }()

	// Debug pode ser chamado múltiplas vezes
	Debug()
	Debug()
	Debug()

	if !debugMode {
		t.Error("Debug mode should remain enabled")
	}
}

// Teste adicional: verifica que panic funciona corretamente em ambos os modos
func TestFail_PanicRecovery(t *testing.T) {
	testCases := []struct {
		name      string
		debugMode bool
	}{
		{"with debug", true},
		{"without debug", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Salva e restaura estado
			oldDebugMode := debugMode
			defer func() { debugMode = oldDebugMode }()

			debugMode = tc.debugMode

			// Verifica que panic acontece
			defer func() {
				if r := recover(); r == nil {
					t.Error("Expected panic")
				} else {
					msg := fmt.Sprintf("%v", r)
					if msg != "panic message" {
						t.Errorf("Expected 'panic message', got: %q", msg)
					}
				}
			}()

			fail("panic message")
		})
	}
}
