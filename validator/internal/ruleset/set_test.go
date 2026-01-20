// internal/ruleset/set_test.go
package ruleset_test

import (
	"testing"

	"github.com/leandroluk/go/validator/internal/engine"
	"github.com/leandroluk/go/validator/internal/ruleset"
)

func TestSet_Put_OverwriteKeepsOrderAndReplacesRule(t *testing.T) {
	set := ruleset.NewSet[string]()
	context := (*engine.Context)(nil)

	set.Put(ruleset.New("a", func(value string, _ *engine.Context) (string, bool) { return value + "A", false }))
	set.Put(ruleset.New("b", func(value string, _ *engine.Context) (string, bool) { return value + "B", false }))
	set.Put(ruleset.New("a", func(value string, _ *engine.Context) (string, bool) { return value + "X", false }))

	out, stopped := set.ApplyAll("", context)
	if stopped {
		t.Fatalf("expected stopped=false, got true")
	}
	if out != "XB" {
		t.Fatalf("expected %q, got %q", "XB", out)
	}
}

func TestSet_Remove_RemovesRuleAndOrder(t *testing.T) {
	set := ruleset.NewSet[string]()
	context := (*engine.Context)(nil)

	set.Put(ruleset.New("a", func(value string, _ *engine.Context) (string, bool) { return value + "A", false }))
	set.Put(ruleset.New("b", func(value string, _ *engine.Context) (string, bool) { return value + "B", false }))
	set.Put(ruleset.New("c", func(value string, _ *engine.Context) (string, bool) { return value + "C", false }))

	set.Remove("b")

	out, stopped := set.ApplyAll("", context)
	if stopped {
		t.Fatalf("expected stopped=false, got true")
	}
	if out != "AC" {
		t.Fatalf("expected %q, got %q", "AC", out)
	}
}

func TestSet_ApplyAll_StopsWhenRuleReturnsTrue(t *testing.T) {
	set := ruleset.NewSet[string]()
	context := (*engine.Context)(nil)

	set.Put(ruleset.New("a", func(value string, _ *engine.Context) (string, bool) { return value + "A", false }))
	set.Put(ruleset.New("b", func(value string, _ *engine.Context) (string, bool) { return value + "B", true }))
	set.Put(ruleset.New("c", func(value string, _ *engine.Context) (string, bool) { return value + "C", false }))

	out, stopped := set.ApplyAll("", context)
	if !stopped {
		t.Fatalf("expected stopped=true, got false")
	}
	if out != "AB" {
		t.Fatalf("expected %q, got %q", "AB", out)
	}
}

func TestSet_Put_EmptyKey_AppendsWithoutOverwriting(t *testing.T) {
	set := ruleset.NewSet[string]()
	context := (*engine.Context)(nil)

	set.Put(ruleset.New("a", func(value string, _ *engine.Context) (string, bool) { return value + "A", false }))
	set.Put(ruleset.New("", func(value string, _ *engine.Context) (string, bool) { return value + "1", false }))
	set.Put(ruleset.New("", func(value string, _ *engine.Context) (string, bool) { return value + "2", false }))

	out, stopped := set.ApplyAll("", context)
	if stopped {
		t.Fatalf("expected stopped=false, got true")
	}
	if out != "A12" {
		t.Fatalf("expected %q, got %q", "A12", out)
	}
}

func TestSet_Remove_UnknownKeyIsNoop(t *testing.T) {
	set := ruleset.NewSet[string]()
	context := (*engine.Context)(nil)

	set.Put(ruleset.New("a", func(value string, _ *engine.Context) (string, bool) { return value + "A", false }))
	set.Remove("nope")

	out, stopped := set.ApplyAll("", context)
	if stopped {
		t.Fatalf("expected stopped=false, got true")
	}
	if out != "A" {
		t.Fatalf("expected %q, got %q", "A", out)
	}
}
