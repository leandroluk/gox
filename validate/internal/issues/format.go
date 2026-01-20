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

func NewValidationError(issues []Issue) error {
	if len(issues) == 0 {
		return nil
	}
	return ValidationError{Issues: append([]Issue(nil), issues...)}
}

func NewValidationErrorWithFormatter(issues []Issue, formatter func(path string, code string, message string, meta map[string]any) string) error {
	if len(issues) == 0 {
		return nil
	}
	return ValidationError{
		Issues:    append([]Issue(nil), issues...),
		formatter: formatter,
	}
}

func (err ValidationError) Unwrap() []error {
	if len(err.Issues) == 0 {
		return nil
	}
	errs := make([]error, len(err.Issues))
	for i, issue := range err.Issues {
		errs[i] = fmt.Errorf("%s", err.formatIssue(issue))
	}
	return errs
}

func (err ValidationError) Error() string {
	if len(err.Issues) == 0 {
		return ""
	}
	if len(err.Issues) == 1 {
		return err.formatIssue(err.Issues[0])
	}

	var builder strings.Builder
	for index, issue := range err.Issues {
		if index > 0 {
			builder.WriteByte('\n')
		}
		builder.WriteString(err.formatIssue(issue))
	}
	return builder.String()
}

func (err ValidationError) MarshalJSON() ([]byte, error) {
	return json.Marshal(err.Issues)
}

func (err ValidationError) formatIssue(issue Issue) string {
	if err.formatter != nil {
		formatted := err.formatter(issue.Path, issue.Code, issue.Message, issue.Meta)
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
