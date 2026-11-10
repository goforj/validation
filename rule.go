package validation

import "context"

// Rule validates a value using application-owned configuration and dependencies.
type Rule[T any] interface {
	// Validate returns nil, an explicit Violation, or an operational error.
	Validate(value T) error
}

// RuleWithContext propagates request cancellation to dependent operations.
type RuleWithContext[T any] interface {
	Rule[T]
	// ValidateContext uses the evaluation context for dependent operations.
	ValidateContext(ctx context.Context, value T) error
}

// RuleFunc adapts an ordinary function into a reusable typed rule.
type RuleFunc[T any] func(T) error

// Validate evaluates the adapted function.
func (f RuleFunc[T]) Validate(value T) error {
	if f == nil {
		return &ConfigurationError{Problem: "nil rule function"}
	}
	return f(value)
}

// ContextRuleFunc adapts a context-aware function into a reusable typed rule.
type ContextRuleFunc[T any] func(context.Context, T) error

// Validate evaluates the rule with a background context.
func (f ContextRuleFunc[T]) Validate(value T) error {
	return f.ValidateContext(context.Background(), value)
}

// ValidateContext forwards the caller's context to the adapted function.
func (f ContextRuleFunc[T]) ValidateContext(ctx context.Context, value T) error {
	if f == nil {
		return &ConfigurationError{Problem: "nil context rule function"}
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return f(ctx, value)
}
