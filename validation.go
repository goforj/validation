package validation

import (
	"context"
	"errors"
	"reflect"
	"strings"

	"github.com/goforj/validation/internal/violation"
)

// Field is a deferred validation operation constructed by a typed builder or group.
type Field interface {
	evaluate(context.Context, []string, *[]Issue) error
}

// Validate evaluates fields sequentially with a background context.
func Validate(fields ...Field) error { return ValidateContext(context.Background(), fields...) }

// ValidateContext collects field violations while giving operational errors precedence.
func ValidateContext(ctx context.Context, fields ...Field) error {
	if ctx == nil {
		ctx = context.Background()
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

// evaluateFields centralizes cancellation and nil-field handling across nested evaluations.
func evaluateFields(ctx context.Context, prefix []string, fields []Field, issues *[]Issue) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	for _, field := range fields {
		if nilValue(field) {
			return &ConfigurationError{Path: renderPath(prefix), Problem: "nil field"}
		}
		if err := field.evaluate(ctx, prefix, issues); err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	return nil
}

// nilValue detects malformed interface values without hiding required dependencies inside rules.
func nilValue(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Pointer, reflect.Func, reflect.Map, reflect.Slice, reflect.Chan, reflect.Interface:
		return v.IsNil()
	default:
		return false
	}
}

// appendPath retains literal field segments until the output path is rendered.
func appendPath(prefix []string, name string) []string {
	result := append([]string(nil), prefix...)
	if name != "" {
		result = append(result, name)
	}
	return result
}

// renderPath escapes literal dots and backslashes rather than interpreting labels as expressions.
func renderPath(path []string) string {
	escaped := make([]string, len(path))
	for i, segment := range path {
		escaped[i] = strings.ReplaceAll(strings.ReplaceAll(segment, `\`, `\\`), ".", `\.`)
	}
	return strings.Join(escaped, ".")
}

// step carries one deferred constraint and its local presentation metadata.
type step[T any] struct {
	rule     Rule[T]
	required bool
	message  *string
	problem  string
}

// field stores immutable chain configuration and a borrowed application value.
type field[T any] struct {
	path    string
	label   string
	value   T
	steps   []step[T]
	problem string
}

// add isolates step storage so branches and concurrent evaluation never share mutable configuration.
func (f field[T]) add(s step[T]) field[T] {
	f.steps = append(append([]step[T](nil), f.steps...), s)
	return f
}

// using attaches a public rule while rejecting nil implementations as configuration errors.
func (f field[T]) using(rule Rule[T]) field[T] {
	s := step[T]{rule: rule}
	if nilValue(rule) {
		s.problem = "nil rule"
	} else if configured, ok := rule.(configuredRule); ok {
		s.problem = configured.configurationProblem()
	}
	return f.add(s)
}

// withMessage copies only presentation metadata and preserves stable rule codes.
func (f field[T]) withMessage(message string) field[T] {
	if len(f.steps) == 0 {
		f.problem = "Message requires a preceding rule"
		return f
	}
	f.steps = append([]step[T](nil), f.steps...)
	f.steps[len(f.steps)-1].message = &message
	return f
}

// run validates configuration before input, then prioritizes required constraints over content rules.
func (f field[T]) run(ctx context.Context, prefix []string, issues *[]Issue) (bool, error) {
	path := renderPath(appendPath(prefix, f.path))
	if f.problem != "" {
		return false, &ConfigurationError{Path: path, Problem: f.problem}
	}
	for _, s := range f.steps {
		if s.problem != "" {
			return false, &ConfigurationError{Path: path, Problem: s.problem}
		}
	}
	for _, required := range []bool{true, false} {
		for _, s := range f.steps {
			if s.required != required {
				continue
			}
			if err := ctx.Err(); err != nil {
				return false, err
			}
			var err error
			if rule, ok := s.rule.(RuleWithContext[T]); ok {
				err = rule.ValidateContext(ctx, f.value)
			} else {
				err = s.rule.Validate(f.value)
			}
			if cancellation := ctx.Err(); cancellation != nil {
				return false, cancellation
			}
			if err == nil {
				continue
			}
			failure := ruleViolation(err)
			if failure == nil {
				return false, err
			}
			label := f.label
			if label == "" {
				label = path
			}
			message := failure.FieldMessage(label)
			if s.message != nil {
				message = *s.message
			}
			*issues = append(*issues, Issue{Path: path, Code: failure.Code(), Message: message, Parameters: failure.Parameters()})
			return false, nil
		}
	}
	return true, nil
}

// ruleViolation preserves operational errors even when joined with input failures.
func ruleViolation(err error) *Violation {
	switch wrapped := err.(type) {
	case interface{ Unwrap() []error }:
		var first *Violation
		for _, child := range wrapped.Unwrap() {
			failure := ruleViolation(child)
			if failure == nil {
				return nil
			}
			if first == nil {
				first = failure
			}
		}
		return first
	case interface{ Unwrap() error }:
		return ruleViolation(wrapped.Unwrap())
	default:
		var failure *Violation
		errors.As(err, &failure)
		return failure
	}
}

// constraint creates a deferred built-in check with safe, value-free failure details.
func constraint[T any](code, message string, parameters map[string]string, valid func(T) bool) Rule[T] {
	return RuleFunc[T](func(value T) error {
		if valid(value) {
			return nil
		}
		return violation.Builtin(code, message, parameters)
	})
}
