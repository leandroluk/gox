// schema/record/record.go
package record

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/leandroluk/gox/validate/internal/ast"
	"github.com/leandroluk/gox/validate/internal/defaults"
	"github.com/leandroluk/gox/validate/internal/engine"
	"github.com/leandroluk/gox/validate/internal/issues"
	"github.com/leandroluk/gox/validate/internal/reflection"
	"github.com/leandroluk/gox/validate/internal/ruleset"
	"github.com/leandroluk/gox/validate/internal/types"
	"github.com/leandroluk/gox/validate/schema"
)

const (
	CodeRequired = "record.required"
	CodeType     = "record.type"

	CodeLen = "record.len"
	CodeMin = "record.min"
	CodeMax = "record.max"

	CodeEq  = "record.eq"
	CodeNe  = "record.ne"
	CodeGt  = "record.gt"
	CodeGte = "record.gte"
	CodeLt  = "record.lt"
	CodeLte = "record.lte"

	CodeUnique = "record.unique"

	KeyCodePrefix  = "record.key."
	CodeKeyInvalid = "record.key.invalid"
)

type ruleMap[T any] struct {
	Eq  T
	Gt  T
	Gte T
	Len T
	Lt  T
	Lte T
	Max T
	Min T
	Ne  T
}

var Msg = ruleMap[string]{
	Eq:  "must be equal",
	Gt:  "must be greater",
	Gte: "must be greater or equal",
	Len: "invalid length",
	Lt:  "must be lower",
	Lte: "must be lower or equal",
	Max: "too large",
	Min: "too small",
	Ne:  "must not be equal",
}

var Rule = ruleMap[func(code string, expected int) ruleset.Rule[int]]{
	Eq: func(code string, expected int) ruleset.Rule[int] {
		return ruleset.New(code, func(actual int, context *engine.Context) (int, bool) {
			if actual == expected {
				return actual, false
			}
			stop := context.AddIssue(code, Msg.Eq, types.AnyMap{"expected": expected, "actual": actual})
			return actual, stop
		})
	},
	Gt: func(code string, expected int) ruleset.Rule[int] {
		return ruleset.New(code, func(actual int, context *engine.Context) (int, bool) {
			if actual > expected {
				return actual, false
			}
			stop := context.AddIssue(code, Msg.Gt, types.AnyMap{"expected": expected, "actual": actual})
			return actual, stop
		})
	},
	Gte: func(code string, expected int) ruleset.Rule[int] {
		return ruleset.New(code, func(actual int, context *engine.Context) (int, bool) {
			if actual >= expected {
				return actual, false
			}
			stop := context.AddIssue(code, Msg.Gte, types.AnyMap{"expected": expected, "actual": actual})
			return actual, stop
		})
	},
	Len: func(code string, expected int) ruleset.Rule[int] {
		return ruleset.New(code, func(actual int, context *engine.Context) (int, bool) {
			if actual == expected {
				return actual, false
			}
			stop := context.AddIssue(code, Msg.Len, types.AnyMap{"expected": expected, "actual": actual})
			return actual, stop
		})
	},
	Lt: func(code string, expected int) ruleset.Rule[int] {
		return ruleset.New(code, func(actual int, context *engine.Context) (int, bool) {
			if actual < expected {
				return actual, false
			}
			stop := context.AddIssue(code, Msg.Lt, types.AnyMap{"expected": expected, "actual": actual})
			return actual, stop
		})
	},
	Lte: func(code string, expected int) ruleset.Rule[int] {
		return ruleset.New(code, func(actual int, context *engine.Context) (int, bool) {
			if actual <= expected {
				return actual, false
			}
			stop := context.AddIssue(code, Msg.Lte, types.AnyMap{"expected": expected, "actual": actual})
			return actual, stop
		})
	},
	Max: func(code string, max int) ruleset.Rule[int] {
		return ruleset.New(code, func(actual int, context *engine.Context) (int, bool) {
			if actual <= max {
				return actual, false
			}
			stop := context.AddIssue(code, Msg.Max, types.AnyMap{"max": max, "actual": actual})
			return actual, stop
		})
	},
	Min: func(code string, min int) ruleset.Rule[int] {
		return ruleset.New(code, func(actual int, context *engine.Context) (int, bool) {
			if actual >= min {
				return actual, false
			}
			stop := context.AddIssue(code, Msg.Min, types.AnyMap{"min": min, "actual": actual})
			return actual, stop
		})
	},
	Ne: func(code string, expected int) ruleset.Rule[int] {
		return ruleset.New(code, func(actual int, context *engine.Context) (int, bool) {
			if actual != expected {
				return actual, false
			}
			stop := context.AddIssue(code, Msg.Ne, types.AnyMap{"expected": expected, "actual": actual})
			return actual, stop
		})
	},
}

func normalizeKeyCode(code string) string {
	if code == "" {
		return CodeKeyInvalid
	}
	if strings.HasPrefix(code, KeyCodePrefix) {
		return code
	}

	lastDot := strings.LastIndexByte(code, '.')
	if lastDot >= 0 && lastDot+1 < len(code) {
		return KeyCodePrefix + code[lastDot+1:]
	}

	return KeyCodePrefix + code
}

func joinPaths(base string, child string) string {
	if base == "" {
		return child
	}
	if child == "" {
		return base
	}

	first := child[0]
	if first == '.' || first == '[' {
		return base + child
	}

	return base + "." + child
}

func keyPathSegment(key string) string {
	escaped := strings.ReplaceAll(key, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return `["` + escaped + `"]`
}

func makeRelative(fullPath string, basePath string) string {
	if basePath == "" {
		return fullPath
	}
	if fullPath == basePath {
		return ""
	}
	if strings.HasPrefix(fullPath, basePath) {
		return fullPath[len(basePath):]
	}
	return fullPath
}

type KeysFunc func(context *engine.Context, key string) bool

type ValuesFunc[V any] func(context *engine.Context, value ast.Value) (V, bool)

type Schema[V any] struct {
	required  bool
	isDefault bool

	unique bool

	keySchema schema.AnySchema
	keyFunc   KeysFunc

	valueSchema schema.AnySchema
	valueFunc   ValuesFunc[V]

	defaultProvider defaults.Provider[map[string]V]

	lengthRules *ruleset.Set[int]
}

type RecordSchema[V any] = Schema[V]

func New[V any]() *Schema[V] {
	return &Schema[V]{
		defaultProvider: defaults.None[map[string]V](),
		lengthRules:     ruleset.NewSet[int](),
	}
}

func (schemaValue *Schema[V]) Required() *Schema[V] {
	schemaValue.required = true
	return schemaValue
}

func (schemaValue *Schema[V]) IsDefault() *Schema[V] {
	schemaValue.isDefault = true
	return schemaValue
}

func (schemaValue *Schema[V]) Len(value int) *Schema[V] {
	schemaValue.lengthRules.Put(Rule.Len(CodeLen, value))
	return schemaValue
}

func (schemaValue *Schema[V]) Min(value int) *Schema[V] {
	schemaValue.lengthRules.Put(Rule.Min(CodeMin, value))
	return schemaValue
}

func (schemaValue *Schema[V]) Max(value int) *Schema[V] {
	schemaValue.lengthRules.Put(Rule.Max(CodeMax, value))
	return schemaValue
}

func (schemaValue *Schema[V]) Eq(value int) *Schema[V] {
	schemaValue.lengthRules.Put(Rule.Eq(CodeEq, value))
	return schemaValue
}

func (schemaValue *Schema[V]) Ne(value int) *Schema[V] {
	schemaValue.lengthRules.Put(Rule.Ne(CodeNe, value))
	return schemaValue
}

func (schemaValue *Schema[V]) Gt(value int) *Schema[V] {
	schemaValue.lengthRules.Put(Rule.Gt(CodeGt, value))
	return schemaValue
}

func (schemaValue *Schema[V]) Gte(value int) *Schema[V] {
	schemaValue.lengthRules.Put(Rule.Gte(CodeGte, value))
	return schemaValue
}

func (schemaValue *Schema[V]) Lt(value int) *Schema[V] {
	schemaValue.lengthRules.Put(Rule.Lt(CodeLt, value))
	return schemaValue
}

func (schemaValue *Schema[V]) Lte(value int) *Schema[V] {
	schemaValue.lengthRules.Put(Rule.Lte(CodeLte, value))
	return schemaValue
}

func (schemaValue *Schema[V]) Unique() *Schema[V] {
	schemaValue.unique = true
	return schemaValue
}

func (schemaValue *Schema[V]) Keys(schemaValueKey schema.AnySchema) *Schema[V] {
	schemaValue.keySchema = schemaValueKey
	schemaValue.keyFunc = nil
	return schemaValue
}

func (schemaValue *Schema[V]) KeysFunc(fn KeysFunc) *Schema[V] {
	schemaValue.keyFunc = fn
	schemaValue.keySchema = nil
	return schemaValue
}

func (schemaValue *Schema[V]) EndKeys() *Schema[V] {
	return schemaValue
}

func (schemaValue *Schema[V]) Dive(schemaValueItem schema.AnySchema) *Schema[V] {
	schemaValue.valueSchema = schemaValueItem
	schemaValue.valueFunc = nil
	return schemaValue
}

func (schemaValue *Schema[V]) Values(fn func(context *engine.Context, value ast.Value) (V, bool)) *Schema[V] {
	schemaValue.valueFunc = fn
	schemaValue.valueSchema = nil
	return schemaValue
}

func (schemaValue *Schema[V]) Default(value map[string]V) *Schema[V] {
	schemaValue.defaultProvider = defaults.Value(value)
	return schemaValue
}

func (schemaValue *Schema[V]) DefaultFunc(fn func() map[string]V) *Schema[V] {
	schemaValue.defaultProvider = defaults.Func(fn)
	return schemaValue
}

func (schemaValue *Schema[V]) Validate(input any, optionList ...schema.Option) (map[string]V, error) {
	options := schema.ApplyOptions(optionList...)
	return schemaValue.validateWithOptions(input, options)
}

func (schemaValue *Schema[V]) ValidateAny(input any, options schema.Options) (any, error) {
	return schemaValue.validateWithOptions(input, options)
}

func (schemaValue *Schema[V]) OutputType() reflect.Type {
	return reflect.TypeFor[map[string]V]()
}

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
					"expected": reflect.TypeFor[V]().String(),
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
	var validationError *issues.ValidationError
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
	var validationError *issues.ValidationError
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
