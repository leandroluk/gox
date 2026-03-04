package env_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/leandroluk/gox/env"
)

func TestEnv_Load(t *testing.T) {
	// 0. No paths
	env.Load()

	// 1. Empty path (covers line 18)
	env.Load("")

	// 2. Path is a directory (covers line 21 info.IsDir())
	tmpDir := t.TempDir()
	env.Load(tmpDir)

	// 3. Path doesn't exist (covers line 21 err != nil)
	env.Load(filepath.Join(tmpDir, "non-existent.env"))

	// 4. Valid file
	content1 := "KEY1=VAL1"
	file1 := filepath.Join(tmpDir, "file1.env")
	os.WriteFile(file1, []byte(content1), 0644)
	env.Load(file1)
	if v := env.Get[string]("KEY1"); v != "VAL1" {
		t.Errorf("Expected VAL1, got %q", v)
	}

	// 5. Multiple files
	content2 := "KEY1=VAL2"
	file2 := filepath.Join(tmpDir, "file2.env")
	os.WriteFile(file2, []byte(content2), 0644)
	env.Load(file2, file1)
	if v := env.Get[string]("KEY1"); v != "VAL2" {
		t.Errorf("Expected VAL2, got %q", v)
	}
}

func TestEnv_Get_EdgeCases(t *testing.T) {
	os.Setenv("EMPTY_VAL", "  ")
	t.Run("Key not found", func(t *testing.T) {
		if v := env.Get[string]("NOT_FOUND"); v != "" {
			t.Errorf("Expected empty string, got %q", v)
		}
		if v := env.Get[int]("NOT_FOUND", 42); v != 42 {
			t.Errorf("Expected 42, got %d", v)
		}
		var zero int
		if v := env.Get[int]("NOT_FOUND"); v != zero {
			t.Errorf("Expected %d, got %d", zero, v)
		}
		// Multiple defaults
		if v := env.Get[int]("NOT_FOUND", 42, 43); v != 42 {
			t.Errorf("Expected 42, got %d", v)
		}
	})
	t.Run("Empty/whitespace value", func(t *testing.T) {
		if v := env.Get[string]("EMPTY_VAL", "default"); v != "default" {
			t.Errorf("Expected 'default', got %q", v)
		}
	})
	t.Run("Type conversion failure", func(t *testing.T) {
		os.Setenv("INVALID_INT", "not-an-int")
		if v := env.Get[int]("INVALID_INT", 123); v != 123 {
			t.Errorf("Expected 123, got %d", v)
		}
		if v := env.Get[int]("INVALID_INT"); v != 0 {
			t.Errorf("Expected 0, got %d", v)
		}
	})
}

func TestEnv_Converters(t *testing.T) {
	t.Run("Bool", func(t *testing.T) {
		os.Setenv("B_TRUE", "true")
		os.Setenv("B_FALSE", "false")
		os.Setenv("B_ERR", "not-bool")
		if !env.Get[bool]("B_TRUE") {
			t.Error("Expected true")
		}
		if env.Get[bool]("B_FALSE") {
			t.Error("Expected false")
		}
		if env.Get[bool]("B_ERR") {
			t.Error("Expected false on error")
		}
	})
	t.Run("Ints", func(t *testing.T) {
		os.Setenv("I8", "127")
		os.Setenv("I16", "32767")
		os.Setenv("I32", "2147483647")
		os.Setenv("I64", "9223372036854775807")
		os.Setenv("I_NEG", "-1")
		if env.Get[int8]("I8") != 127 {
			t.Error("int8 fail")
		}
		if env.Get[int16]("I16") != 32767 {
			t.Error("int16 fail")
		}
		if env.Get[int32]("I32") != 2147483647 {
			t.Error("int32 fail")
		}
		if env.Get[int64]("I64") != 9223372036854775807 {
			t.Error("int64 fail")
		}
		if env.Get[int]("I_NEG") != -1 {
			t.Error("int fail")
		}

		os.Setenv("I_OVER", "99999999999999999999999999999999999999")
		if env.Get[int64]("I_OVER") != 0 {
			t.Error("overflow fail")
		}
	})

	t.Run("Floats", func(t *testing.T) {
		os.Setenv("F32", "3.14")
		os.Setenv("F64", "3.1415926535")
		os.Setenv("F_ERR", "not-a-float")
		if env.Get[float32]("F32") != 3.14 {
			t.Error("float32 fail")
		}
		if env.Get[float64]("F64") != 3.1415926535 {
			t.Error("float64 fail")
		}
		if env.Get[float64]("F_ERR") != 0 {
			t.Error("float error fail")
		}
	})
	t.Run("Time", func(t *testing.T) {
		os.Setenv("T_RFC", "2023-10-27T10:00:00Z")
		os.Setenv("T_FULL", "2023-10-27 10:00:00")
		os.Setenv("T_DATE", "2023-10-27")
		os.Setenv("T_ERR", "invalid-time")
		if env.Get[time.Time]("T_RFC").IsZero() {
			t.Error("RFC3339 fail")
		}
		if env.Get[time.Time]("T_FULL").IsZero() {
			t.Error("Full fail")
		}
		if env.Get[time.Time]("T_DATE").IsZero() {
			t.Error("Date fail")
		}
		if !env.Get[time.Time]("T_ERR").IsZero() {
			t.Error("Should be zero on error")
		}
	})
	t.Run("Duration", func(t *testing.T) {
		os.Setenv("DUR", "1h30m")
		os.Setenv("DUR_ERR", "invalid")
		if env.Get[time.Duration]("DUR") != 90*time.Minute {
			t.Error("Duration fail")
		}
		if env.Get[time.Duration]("DUR_ERR") != 0 {
			t.Error("Duration error fail")
		}
	})
	t.Run("JSON", func(t *testing.T) {
		os.Setenv("JS", `{"foo":"bar"}`)
		os.Setenv("JS_ERR", `{"foo":`)
		if string(env.Get[json.RawMessage]("JS")) != `{"foo":"bar"}` {
			t.Error("JSON fail")
		}
		if len(env.Get[json.RawMessage]("JS_ERR")) != 0 {
			t.Error("JSON error fail")
		}
	})
	t.Run("Unsupported", func(t *testing.T) {
		type custom struct{}
		os.Setenv("CUSTOM_KEY", "some-value")
		result := env.Get[custom]("CUSTOM_KEY")
		// Should return zero value for unsupported type
		if result != (custom{}) {
			t.Error("Expected zero value for unsupported type")
		}
	})
}

func TestEnv_Parser_EdgeCases(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "edge.env")
	content := `
# Comment
EXPORT_VAR=val1
export EXPORTED_VAR=val2
export   SPACE_EXPORT   =   val3
EMPTY_KEY=
=VAL_NO_KEY
INVALID_LINE_NO_EQUALS
SINGLE_QUOTE='quoted value'
DOUBLE_QUOTE="quoted value"
MIXED_QUOTES="'mixed'"
INLINE_COMMENT=val4 # comment
QUOTE_COMMENT="val # with hash" # real comment
UNCLOSED="unclosed
	`
	os.WriteFile(tmpFile, []byte(content), 0644)
	env.Load(tmpFile)
	tests := []struct {
		key      string
		expected string
	}{
		{"EXPORT_VAR", "val1"},
		{"EXPORTED_VAR", "val2"},
		{"SPACE_EXPORT", "val3"},
		{"EMPTY_KEY", ""},
		{"SINGLE_QUOTE", "quoted value"},
		{"DOUBLE_QUOTE", "quoted value"},
		{"MIXED_QUOTES", "'mixed'"},
		{"INLINE_COMMENT", "val4"},
		{"QUOTE_COMMENT", "val # with hash"},
		{"UNCLOSED", "\"unclosed"},
	}
	for _, tt := range tests {
		if v := env.Get[string](tt.key); v != tt.expected {
			t.Errorf("Key %s: expected %q, got %q", tt.key, tt.expected, v)
		}
	}
}

func TestEnv_MoreEdgeCases(t *testing.T) {
	// Length 1 value (covers parser.go line 60 skip)
	tmpFile := filepath.Join(t.TempDir(), "short.env")
	os.WriteFile(tmpFile, []byte("K=V\nexport "), 0644)
	env.Load(tmpFile)
	if env.Get[string]("K") != "V" {
		t.Error("short value fail")
	}
}
