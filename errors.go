package validation

import (
	"fmt"
	"maps"
	"strings"

	"github.com/goforj/validation/internal/violation"
)

// Violation is a coded, client-facing failure returned by a rule.
type Violation = violation.Violation

// Invalid creates a violation with a stable code and a message safe for clients.
func Invalid(code, message string) *Violation { return violation.New(code, message) }

// Issue describes a failed field without retaining its submitted value.
type Issue struct {
	Path       string            `json:"path"`
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Parameters map[string]string `json:"parameters,omitempty"`
}

// Errors contains ordered field failures and supports errors.As.
type Errors struct{ issues []Issue }

// Error returns field messages in declaration order.
func (e *Errors) Error() string {
	messages := make([]string, len(e.issues))
	for i, issue := range e.issues {
		messages[i] = issue.Message
	}
	return strings.Join(messages, "; ")
}

// First returns the first message at path, or an empty string if none exists.
func (e *Errors) First(path string) string {
	for _, issue := range e.issues {
		if issue.Path == path {
			return issue.Message
		}
	}
	return ""
}

// Messages returns an independent field-to-messages map for form or API responses.
func (e *Errors) Messages() map[string][]string {
	result := make(map[string][]string)
	for _, issue := range e.issues {
		result[issue.Path] = append(result[issue.Path], issue.Message)
	}
	return result
}

// Issues returns ordered copies of issues and their constraint parameters.
func (e *Errors) Issues() []Issue {
	result := make([]Issue, len(e.issues))
	for i, issue := range e.issues {
		result[i] = issue
		result[i].Parameters = maps.Clone(issue.Parameters)
	}
	return result
}

// ConfigurationError identifies a malformed chain or rule rather than invalid input.
type ConfigurationError struct {
	Path    string
	Problem string
}

// Error describes the invalid rule configuration without printing submitted values.
func (e *ConfigurationError) Error() string {
	return fmt.Sprintf("validation: %s: %s", e.Path, e.Problem)
}
