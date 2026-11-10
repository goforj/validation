package validation

import "context"

// Validator is an immutable evaluation policy; its zero value uses package defaults.
// Rule and formatter callbacks retain application ownership and concurrency requirements.
type Validator struct {
	collectAll, stopFirst bool
	formatter             func(Issue) string
	problem               string
}

// New creates a validator with first-violation-per-field collection and default messages.
func New() Validator { return Validator{} }

// CollectAll collects all content failures per field; required failures still stop that field.
func (v Validator) CollectAll() Validator { v.collectAll = true; return v }

// StopOnFirstError stops the entire evaluation after its first input violation.
func (v Validator) StopOnFirstError() Validator { v.stopFirst = true; return v }

// FormatMessages formats violations by path, code, and owned parameters, unless Message overrides it.
// The callback receives no submitted value and may provide application-localized messages.
func (v Validator) FormatMessages(format func(Issue) string) Validator {
	v.formatter = format
	if format == nil {
		v.problem = "nil message formatter"
	}
	return v
}

// Validate evaluates fields with a background context and the configured policy.
func (v Validator) Validate(fields ...Field) error {
	return v.ValidateContext(context.Background(), fields...)
}

// policyKey isolates evaluation policy from caller context values.
type policyKey struct{}

// settings reads the policy shared by nested field evaluation.
func settings(ctx context.Context) Validator { v, _ := ctx.Value(policyKey{}).(Validator); return v }

// stopped prevents lazy callbacks from running after a global first failure.
func stopped(ctx context.Context, issues []Issue) bool {
	return len(issues) > 0 && settings(ctx).stopFirst
}

// ValidateContext evaluates with request cancellation and the configured collection policy.
func (v Validator) ValidateContext(ctx context.Context, fields ...Field) error {
	if v.problem != "" {
		return &ConfigurationError{Problem: v.problem}
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if v.collectAll || v.stopFirst || v.formatter != nil || ctx.Value(policyKey{}) != nil {
		ctx = context.WithValue(ctx, policyKey{}, v)
	}
	var issues []Issue
	if err := evaluateFields(ctx, nil, fields, &issues); err != nil {
		return err
	}
	if len(issues) > 0 {
		return &Errors{issues: issues}
	}
	return nil
}
