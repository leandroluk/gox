// internal/engine/snapshot_test.go
package engine_test

import (
	"testing"

	"github.com/leandroluk/go/validate/internal/engine"
	"github.com/leandroluk/go/validate/internal/testkit"
	"github.com/leandroluk/go/validate/schema"
)

func TestTakeAndRestoreSnapshot_RestoresPathAndIssues(t *testing.T) {
	context := engine.NewContext(schema.DefaultOptions())
	context.PushField("a")
	context.AddIssue("code", "msg")

	snapshot := engine.TakeSnapshot(context)

	context.PushField("b")
	context.AddIssue("code2", "msg2")

	engine.RestoreSnapshot(context, snapshot)

	if context.PathString() != "a" {
		t.Fatalf("expected path %q, got %q", "a", context.PathString())
	}
	if engine.IssuesCount(context) != 1 {
		t.Fatalf("expected 1 issue, got %d", engine.IssuesCount(context))
	}

	validationError := testkit.RequireValidationError(t, context.Error())
	if len(validationError.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(validationError.Issues))
	}
	if validationError.Issues[0].Code != "code" {
		t.Fatalf("expected code %q, got %q", "code", validationError.Issues[0].Code)
	}
}

func TestSandbox_RollsBackContextAndCanCommitIssues(t *testing.T) {
	context := engine.NewContext(schema.DefaultOptions())
	context.PushField("root")

	result := engine.Sandbox(context, func() (int, bool) {
		context.PushField("child")
		stop := context.AddIssue("code", "msg")
		return 10, stop
	})

	if result.Passed {
		t.Fatalf("expected passed=false")
	}
	if context.PathString() != "root" {
		t.Fatalf("expected path %q, got %q", "root", context.PathString())
	}
	if engine.IssuesCount(context) != 0 {
		t.Fatalf("expected 0 issues, got %d", engine.IssuesCount(context))
	}
	if len(result.Issues) != 1 {
		t.Fatalf("expected 1 sandbox issue, got %d", len(result.Issues))
	}

	result.Commit(context)

	validationError := testkit.RequireValidationError(t, context.Error())
	if len(validationError.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(validationError.Issues))
	}
	if validationError.Issues[0].Code != "code" {
		t.Fatalf("expected code %q, got %q", "code", validationError.Issues[0].Code)
	}
	if validationError.Issues[0].Path != "root.child" {
		t.Fatalf("expected path %q, got %q", "root.child", validationError.Issues[0].Path)
	}
}
