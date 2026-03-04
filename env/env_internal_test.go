package env

import (
	"io"
	"os"
	"testing"
)

type errReader struct{}

func (e *errReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func TestEnv_Internal_ScannerError(t *testing.T) {
	// Covering parseReader error
	err := parseReader(&errReader{})
	if err == nil {
		t.Error("Expected scanner error, got nil")
	}

	// Covering loadEnvFile error return (line 13 in parser.go)
	err = loadEnvFile("non-existent-file-path-that-really-should-not-exist")
	if err == nil {
		t.Error("Expected loadEnvFile error, got nil")
	}
}

func TestEnv_SystemLookup(t *testing.T) {
	os.Setenv("SYS_VAR", "VAL")
	defer os.Unsetenv("SYS_VAR")
	if v, found := lookupEnv("SYS_VAR"); !found || v != "VAL" {
		t.Error("lookupEnv fail")
	}
}

func TestEnv_ExpandValue(t *testing.T) {
	os.Setenv("V1", "foo")
	if v := expandValue("${V1}-bar"); v != "foo-bar" {
		t.Errorf("Expected foo-bar, got %q", v)
	}
}

func TestEnv_ExpandValue_NotFound(t *testing.T) {
	// Variable not found in expansion (covers lookupEnv return "" path)
	if v := expandValue("${NOT_FOUND_FOR_REAL}"); v != "" {
		t.Errorf("Expected empty string for not found var, got %q", v)
	}
}

