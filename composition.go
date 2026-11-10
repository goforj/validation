package validation

import (
	"context"

	"github.com/goforj/validation/internal/violation"
)

// compositeRule owns its rule list while borrowing each rule's configuration.
type compositeRule[T any] struct {
	mode  string
	rules []Rule[T]
}

// Any accepts a value when a rule succeeds; operational errors abort evaluation.
func Any[T any](rules ...Rule[T]) Rule[T] {
	return compositeRule[T]{mode: "any", rules: append([]Rule[T](nil), rules...)}
}

// All requires every rule, returning the first violation or operational error.
func All[T any](rules ...Rule[T]) Rule[T] {
	return compositeRule[T]{mode: "all", rules: append([]Rule[T](nil), rules...)}
}

// Not inverts a rule's input-validity result while preserving operational errors.
func Not[T any](rule Rule[T]) Rule[T] { return compositeRule[T]{mode: "not", rules: []Rule[T]{rule}} }

// configurationProblem validates the complete rule list before any branch can succeed.
func (r compositeRule[T]) configurationProblem() string {
	if len(r.rules) == 0 {
		return "rule composition requires at least one rule"
	}
	for _, rule := range r.rules {
		if nilValue(rule) {
			return "nil composed rule"
		}
		if configured, ok := rule.(configuredRule); ok {
			if problem := configured.configurationProblem(); problem != "" {
				return problem
			}
		}
	}
	return ""
}

// Validate evaluates the composition with a background context.
func (r compositeRule[T]) Validate(value T) error {
	return r.ValidateContext(context.Background(), value)
}

// ValidateContext preserves cancellation and separates input failures from operational failures.
func (r compositeRule[T]) ValidateContext(ctx context.Context, value T) error {
	if problem := r.configurationProblem(); problem != "" {
		return &ConfigurationError{Problem: problem}
	}
	if ctx == nil {
		ctx = context.Background()
	}
	for _, rule := range r.rules {
		if err := ctx.Err(); err != nil {
			return err
		}
		var err error
		if contextual, ok := rule.(RuleWithContext[T]); ok {
			err = contextual.ValidateContext(ctx, value)
		} else {
			err = rule.Validate(value)
		}
		if cancellation := ctx.Err(); cancellation != nil {
			return cancellation
		}
		if err != nil && ruleViolation(err) == nil {
			return err
		}
		switch r.mode {
		case "any":
			if err == nil {
				return nil
			}
		case "all":
			if err != nil {
				return err
			}
		case "not":
			if err != nil {
				return nil
			}
			return violation.Builtin("not", "must not satisfy the excluded rule", nil)
		}
	}
	if r.mode == "any" {
		return violation.Builtin("any", "must satisfy at least one allowed rule", nil)
	}
	return nil
}
