// internal/issues/issues.go
package issues

import (
	"encoding/json"
	"fmt"
	"strings"
)

type ValidationError struct {
	Issues    []Issue
	formatter func(path string, code string, message string, meta map[string]any) string
}

var _ error = (*ValidationError)(nil)
var _ interface{ Unwrap() []error } = (*ValidationError)(nil)

func NewValidationError(issues []Issue) *ValidationError {
	if len(issues) == 0 {
		return nil
	}
	return &ValidationError{Issues: append([]Issue(nil), issues...)}
}

func NewValidationErrorWithFormatter(
	issues []Issue,
	formatter func(path string, code string, message string, meta map[string]any) string,
) error {
	if len(issues) == 0 {
		return nil
	}
	return &ValidationError{
		Issues:    append([]Issue(nil), issues...),
		formatter: formatter,
	}
}

func (e *ValidationError) Unwrap() []error {
	if len(e.Issues) == 0 {
		return nil
	}
	errs := make([]error, len(e.Issues))
	for i, issue := range e.Issues {
		errs[i] = fmt.Errorf("%s", e.formatIssue(issue))
	}
	return errs
}

func (e *ValidationError) Error() string {
	if e == nil || len(e.Issues) == 0 {
		return ""
	}
	if len(e.Issues) == 1 {
		return e.formatIssue(e.Issues[0])
	}

	var builder strings.Builder
	for index, issue := range e.Issues {
		if index > 0 {
			builder.WriteByte('\n')
		}
		builder.WriteString(e.formatIssue(issue))
	}
	return builder.String()
}

func (e *ValidationError) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Issues)
}

func (e *ValidationError) formatIssue(issue Issue) string {
	if e.formatter != nil {
		formatted := e.formatter(issue.Path, issue.Code, issue.Message, issue.Meta)
		if formatted != "" {
			return formatted
		}
	}
	return FormatIssue(issue)
}

func FormatIssue(issue Issue) string {
	message := issue.Message
	if message == "" {
		if issue.Code != "" {
			message = issue.Code
		} else {
			message = "invalid"
		}
	}

	var builder strings.Builder
	if issue.Path != "" {
		builder.WriteString(issue.Path)
		builder.WriteString(": ")
	}
	builder.WriteString(message)

	if issue.Code != "" && issue.Message != "" {
		builder.WriteString(" (")
		builder.WriteString(issue.Code)
		builder.WriteByte(')')
	}

	return builder.String()
}

func FormatIssues(issues []Issue) string {
	if len(issues) == 0 {
		return ""
	}
	if len(issues) == 1 {
		return FormatIssue(issues[0])
	}

	var builder strings.Builder
	for index, issue := range issues {
		if index > 0 {
			builder.WriteByte('\n')
		}
		builder.WriteString(FormatIssue(issue))
	}
	return builder.String()
}

func Wrapf(base error, format string, args ...any) error {
	if base == nil {
		return fmt.Errorf(format, args...)
	}
	return fmt.Errorf("%w: %s", base, fmt.Sprintf(format, args...))
}

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

type List struct {
	items []Issue
}

func NewList() List {
	return List{items: make([]Issue, 0)}
}

func (list *List) Add(issue Issue) {
	if issue.IsZero() {
		return
	}
	list.items = append(list.items, issue)
}

func (list *List) AddWithLimit(issue Issue, maxIssues int) bool {
	if issue.IsZero() {
		return false
	}
	if maxIssues > 0 && len(list.items) >= maxIssues {
		return true
	}
	list.items = append(list.items, issue)
	if maxIssues > 0 && len(list.items) >= maxIssues {
		return true
	}
	return false
}

func (list *List) Merge(other List) {
	if len(other.items) == 0 {
		return
	}
	list.items = append(list.items, other.items...)
}

func (list *List) MergeWithLimit(other List, maxIssues int) bool {
	if len(other.items) == 0 {
		return false
	}
	if maxIssues <= 0 {
		list.items = append(list.items, other.items...)
		return false
	}

	for _, issue := range other.items {
		if len(list.items) >= maxIssues {
			return true
		}
		if issue.IsZero() {
			continue
		}
		list.items = append(list.items, issue)
	}

	return len(list.items) >= maxIssues
}

func (list *List) Items() []Issue {
	return append([]Issue(nil), list.items...)
}

func (list *List) Len() int {
	return len(list.items)
}

func (list *List) IsEmpty() bool {
	return len(list.items) == 0
}

func (list *List) Clear() {
	list.items = list.items[:0]
}
