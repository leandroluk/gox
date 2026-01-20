// schema/record/validate.go
package record

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/leandroluk/go/validate/internal/ast"
	"github.com/leandroluk/go/validate/internal/defaults"
	"github.com/leandroluk/go/validate/internal/engine"
	"github.com/leandroluk/go/validate/internal/issues"
	"github.com/leandroluk/go/validate/internal/reflection"
	"github.com/leandroluk/go/validate/schema"
)

func (schemaValue *Schema[V]) validateWithOptions(input any, options schema.Options) (map[string]V, error) {
	context := engine.NewContext(options)

	value, err := engine.InputToASTWithOptions(input, options)
	if err != nil {
		return nil, err
	}

	context.SetRoot(value)

	output, _ := schemaValue.validateAST(context, value)
	return output, context.Error()
}

func (schemaValue *Schema[V]) validateAST(context *engine.Context, value ast.Value) (map[string]V, bool) {
	if defaultValue, ok := defaults.Apply(value.Presence, context.Options, schemaValue.defaultProvider); ok {
		return defaultValue, false
	}

	if value.IsMissing() || value.IsNull() {
		if schemaValue.required {
			stop := context.AddIssue(CodeRequired, "required")
			return nil, stop
		}
		return nil, false
	}

	if value.Kind != ast.KindObject {
		stop := addIssueAtPath(context, context.PathString(), CodeType, "expected record", map[string]any{
			"expected": "record",
			"actual":   value.Kind.String(),
		})
		return nil, stop
	}

	if schemaValue.isDefault && reflection.IsDefault(value.Object) {
		return map[string]V{}, false
	}

	length := len(value.Object)
	_, stop := schemaValue.lengthRules.ApplyAll(length, context)
	if stop {
		return nil, true
	}

	recordBasePath := context.PathString()
	output := make(map[string]V, len(value.Object))
	seen := make(map[string]struct{}, len(value.Object))

	for key, child := range value.Object {
		entryPath := joinPaths(recordBasePath, keyPathSegment(key))

		keyOK, stop := schemaValue.validateKey(context, recordBasePath, entryPath, key)
		if stop {
			return output, true
		}
		if !keyOK {
			continue
		}

		if schemaValue.unique {
			hash := ast.Hash(child)
			if _, exists := seen[hash]; exists {
				stopUnique := addIssueAtPath(context, entryPath, CodeUnique, "duplicate value", nil)
				if stopUnique {
					return output, true
				}
				continue
			}
			seen[hash] = struct{}{}
		}

		if schemaValue.valueFunc != nil {
			typed, stopValue := schemaValue.validateValueFunc(context, recordBasePath, entryPath, child)
			output[key] = typed
			if stopValue {
				return output, true
			}
			continue
		}

		if schemaValue.valueSchema != nil {
			valueAny, err := schemaValue.valueSchema.ValidateAny(child, context.Options)
			if err != nil {
				stopValue := commitChildError(context, err, entryPath)
				if stopValue {
					return output, true
				}
				continue
			}

			typed, ok := valueAny.(V)
			if !ok {
				stopType := addIssueAtPath(context, entryPath, CodeType, fmt.Sprintf("value schema returned %T", valueAny), map[string]any{
					"expected": reflect.TypeOf((*V)(nil)).Elem().String(),
					"actual":   reflect.TypeOf(valueAny).String(),
				})
				if stopType {
					return output, true
				}
				continue
			}

			output[key] = typed
		}
	}

	return output, false
}

func (schemaValue *Schema[V]) validateKey(context *engine.Context, basePath string, entryPath string, key string) (bool, bool) {
	if schemaValue.keySchema != nil {
		_, err := schemaValue.keySchema.ValidateAny(key, context.Options)
		if err == nil {
			return true, false
		}
		return false, schemaValue.commitKeySchemaError(context, err, entryPath)
	}

	if schemaValue.keyFunc != nil {
		return schemaValue.validateKeyFunc(context, basePath, entryPath, key)
	}

	return true, false
}

func (schemaValue *Schema[V]) validateKeyFunc(context *engine.Context, basePath string, entryPath string, key string) (bool, bool) {
	snapshot := engine.TakeSnapshot(context)
	startCount := engine.IssuesCount(context)

	stopFromFunc := schemaValue.keyFunc(context, key)

	all := context.Issues.Items()
	added := append([]issues.Issue(nil), all[startCount:]...)

	engine.RestoreSnapshot(context, snapshot)

	if len(added) == 0 {
		return true, stopFromFunc
	}

	committedStop := false
	for _, issue := range added {
		relative := makeRelative(issue.Path, basePath)
		path := joinPaths(entryPath, relative)
		code := normalizeKeyCode(issue.Code)

		if addIssueAtPath(context, path, code, issue.Message, issue.Meta) {
			committedStop = true
		}
	}

	return false, committedStop || stopFromFunc
}

func (schemaValue *Schema[V]) validateValueFunc(context *engine.Context, basePath string, entryPath string, value ast.Value) (V, bool) {
	snapshot := engine.TakeSnapshot(context)
	startCount := engine.IssuesCount(context)

	output, stopFromFunc := schemaValue.valueFunc(context, value)

	all := context.Issues.Items()
	added := append([]issues.Issue(nil), all[startCount:]...)

	engine.RestoreSnapshot(context, snapshot)

	if len(added) == 0 {
		return output, stopFromFunc
	}

	committedStop := false
	for _, issue := range added {
		relative := makeRelative(issue.Path, basePath)
		path := joinPaths(entryPath, relative)

		if addIssueAtPath(context, path, issue.Code, issue.Message, issue.Meta) {
			committedStop = true
		}
	}

	return output, committedStop || stopFromFunc
}

func (schemaValue *Schema[V]) commitKeySchemaError(context *engine.Context, err error, entryPath string) bool {
	var validationError issues.ValidationError
	if !errors.As(err, &validationError) {
		return addIssueAtPath(context, entryPath, CodeKeyInvalid, "invalid key", map[string]any{
			"error": err.Error(),
		})
	}

	committedStop := false
	for _, issue := range validationError.Issues {
		path := joinPaths(entryPath, issue.Path)
		code := normalizeKeyCode(issue.Code)
		if addIssueAtPath(context, path, code, issue.Message, issue.Meta) {
			committedStop = true
		}
	}

	return committedStop
}

func commitChildError(context *engine.Context, err error, basePath string) bool {
	var validationError issues.ValidationError
	if !errors.As(err, &validationError) {
		return addIssueAtPath(context, basePath, CodeType, "invalid value", map[string]any{
			"error": err.Error(),
		})
	}

	committedStop := false
	for _, issue := range validationError.Issues {
		path := joinPaths(basePath, issue.Path)
		if addIssueAtPath(context, path, issue.Code, issue.Message, issue.Meta) {
			committedStop = true
		}
	}

	return committedStop
}

func addIssueAtPath(context *engine.Context, path string, code string, message string, meta map[string]any) bool {
	issue := issues.NewIssue(path, code, message)
	if meta != nil {
		issue = issue.WithMetaMap(meta)
	}

	limitReached := context.Issues.AddWithLimit(issue, context.Options.MaxIssues)
	if context.Options.FailFast {
		return true
	}
	return limitReached
}
