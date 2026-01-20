// schema/record/rules.go
package record

import "strings"

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
