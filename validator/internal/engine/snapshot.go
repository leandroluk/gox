// internal/engine/snapshot.go
package engine

import (
	"github.com/leandroluk/go/validator/internal/issues"
	"github.com/leandroluk/go/validator/internal/path"
)

type Snapshot struct {
	Path   path.Builder
	Issues issues.List
}

func TakeSnapshot(context *Context) Snapshot {
	return Snapshot{
		Path:   context.Path,
		Issues: context.Issues,
	}
}

func RestoreSnapshot(context *Context, snapshot Snapshot) {
	context.Path = snapshot.Path
	context.Issues = snapshot.Issues
}

func IssuesCount(context *Context) int {
	return len(context.Issues.Items())
}

type SandboxResult[T any] struct {
	Output T
	Stop   bool
	Passed bool
	Issues []issues.Issue
}

func Sandbox[T any](context *Context, run func() (T, bool)) SandboxResult[T] {
	snapshot := TakeSnapshot(context)
	before := IssuesCount(context)

	output, stop := run()

	afterItems := context.Issues.Items()
	var delta []issues.Issue
	if len(afterItems) > before {
		delta = append([]issues.Issue(nil), afterItems[before:]...)
	}

	RestoreSnapshot(context, snapshot)

	return SandboxResult[T]{
		Output: output,
		Stop:   stop,
		Passed: len(delta) == 0,
		Issues: delta,
	}
}

func (result SandboxResult[T]) Commit(context *Context) bool {
	for _, issue := range result.Issues {
		limitReached := context.Issues.AddWithLimit(issue, context.Options.MaxIssues)
		if context.Options.FailFast {
			return true
		}
		if limitReached {
			return true
		}
	}
	return false
}
