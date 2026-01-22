// internal/issues/format.go
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
	if len(e.Issues) == 0 {
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
