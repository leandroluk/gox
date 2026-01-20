// internal/engine/context.go
package engine

import (
	"github.com/leandroluk/gox/validate/internal/ast"
	"github.com/leandroluk/gox/validate/internal/issues"
	"github.com/leandroluk/gox/validate/internal/path"
	"github.com/leandroluk/gox/validate/schema"
)

type Context struct {
	Options schema.Options
	Path    path.Builder
	Issues  issues.List
	Root    ast.Value
}

func NewContext(options schema.Options) *Context {
	return &Context{
		Options: options,
		Path:    path.NewBuilder(),
		Issues:  issues.NewList(),
		Root:    ast.NullValue(),
	}
}

func (context *Context) SetRoot(value ast.Value) {
	context.Root = value
}

func (context *Context) PushField(name string) {
	context.Path.PushField(name)
}

func (context *Context) PushIndex(index int) {
	context.Path.PushIndex(index)
}

func (context *Context) PushKey(key string) {
	context.Path.PushKey(key)
}

func (context *Context) Pop() {
	context.Path.Pop()
}

func (context *Context) PathString() string {
	return context.Path.String()
}

func (context *Context) AddIssue(code string, message string) bool {
	issue := issues.NewIssue(context.PathString(), code, message)
	limitReached := context.Issues.AddWithLimit(issue, context.Options.MaxIssues)
	if context.Options.FailFast {
		return true
	}
	return limitReached
}

func (context *Context) AddIssueWithMeta(code string, message string, meta map[string]any) bool {
	issue := issues.NewIssue(context.PathString(), code, message).WithMetaMap(meta)
	limitReached := context.Issues.AddWithLimit(issue, context.Options.MaxIssues)
	if context.Options.FailFast {
		return true
	}
	return limitReached
}

func (context *Context) Error() error {
	if context.Options.Formatter != nil {
		return issues.NewValidationErrorWithFormatter(context.Issues.Items(), context.Options.Formatter)
	}
	return issues.NewValidationError(context.Issues.Items())
}
