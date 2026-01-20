// internal/reflection/is_default_test.go
package reflection_test

import (
	"testing"
	"time"

	"github.com/leandroluk/go/validator/internal/reflection"
)

func TestIsDefault_Primitives(t *testing.T) {
	if !reflection.IsDefault(0) {
		t.Fatalf("expected int(0) to be default")
	}
	if !reflection.IsDefault("") {
		t.Fatalf("expected empty string to be default")
	}
	if !reflection.IsDefault(false) {
		t.Fatalf("expected false to be default")
	}

	if reflection.IsDefault(1) {
		t.Fatalf("expected int(1) to not be default")
	}
	if reflection.IsDefault("x") {
		t.Fatalf("expected 'x' to not be default")
	}
	if reflection.IsDefault(true) {
		t.Fatalf("expected true to not be default")
	}
}

func TestIsDefault_InterfaceWrapped(t *testing.T) {
	var a any = 0
	var b any = ""
	var c any = false

	if !reflection.IsDefault(a) || !reflection.IsDefault(b) || !reflection.IsDefault(c) {
		t.Fatalf("expected interface-wrapped zero values to be default")
	}
}

func TestIsDefault_Time(t *testing.T) {
	if !reflection.IsDefault(time.Time{}) {
		t.Fatalf("expected time.Time{} to be default")
	}
	if reflection.IsDefault(time.Unix(1, 0).UTC()) {
		t.Fatalf("expected non-zero time to not be default")
	}
}

func TestIsDefault_SlicesAndMaps(t *testing.T) {
	var sNil []int
	if !reflection.IsDefault(sNil) {
		t.Fatalf("expected nil slice to be default")
	}
	if reflection.IsDefault([]int{}) {
		t.Fatalf("expected empty but non-nil slice to not be default")
	}

	var mNil map[string]int
	if !reflection.IsDefault(mNil) {
		t.Fatalf("expected nil map to be default")
	}
	if reflection.IsDefault(map[string]int{}) {
		t.Fatalf("expected empty but non-nil map to not be default")
	}
}

func TestIsDefault_Pointers(t *testing.T) {
	var pNil *int
	if !reflection.IsDefault(pNil) {
		t.Fatalf("expected nil pointer to be default")
	}

	x := 0
	p := &x
	if reflection.IsDefault(p) {
		t.Fatalf("expected non-nil pointer to not be default")
	}
}
