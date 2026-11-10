// Package violation shares coded failures between validation and format rules.
package violation

import "maps"

// Violation describes an input failure without retaining the submitted value.
type Violation struct {
	code       string
	message    string
	parameters map[string]string
	field      bool
}

// New creates a client-facing failure with a literal message.
func New(code, message string) *Violation {
	return &Violation{code: code, message: message}
}

// Builtin creates a failure whose message can be associated with a field label.
func Builtin(code, message string, parameters map[string]string) *Violation {
	return &Violation{code: code, message: message, parameters: maps.Clone(parameters), field: true}
}

// Error returns a safe message without including submitted values.
func (v *Violation) Error() string { return v.message }

// Code returns the stable machine-readable rule identifier.
func (v *Violation) Code() string { return v.code }

// Parameters returns an independent copy of the rule's constraint parameters.
func (v *Violation) Parameters() map[string]string { return maps.Clone(v.parameters) }

// FieldMessage associates built-in prose with a field while preserving literal custom messages.
func (v *Violation) FieldMessage(label string) string {
	if v.field {
		return "The " + label + " field " + v.message + "."
	}
	return v.message
}
