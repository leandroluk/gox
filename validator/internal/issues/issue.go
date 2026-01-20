// internal/issues/issue.go
package issues

type Issue struct {
	Path    string
	Code    string
	Message string
	Meta    map[string]any
}

func NewIssue(path string, code string, message string) Issue {
	return Issue{
		Path:    path,
		Code:    code,
		Message: message,
	}
}

func (issue Issue) WithMeta(key string, value any) Issue {
	if issue.Meta == nil {
		issue.Meta = make(map[string]any, 1)
	}
	issue.Meta[key] = value
	return issue
}

func (issue Issue) WithMetaMap(meta map[string]any) Issue {
	if len(meta) == 0 {
		return issue
	}
	if issue.Meta == nil {
		issue.Meta = make(map[string]any, len(meta))
	}
	for key, value := range meta {
		issue.Meta[key] = value
	}
	return issue
}

func (issue Issue) IsZero() bool {
	return issue.Path == "" && issue.Code == "" && issue.Message == "" && len(issue.Meta) == 0
}
