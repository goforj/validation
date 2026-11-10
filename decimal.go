package validation

import (
	"context"
	"strconv"
)

// DecimalField is an immutable fluent chain for exact plain-decimal text.
// T may be an application-defined string type; submitted text is never rounded or rewritten.
type DecimalField[T ~string] struct{ f field[T] }

// Decimal starts a deferred chain for signed decimal text without exponent notation.
// Values require ASCII integer digits and may include a fractional part with digits after the point.
func Decimal[T ~string](path string, value T) DecimalField[T] {
	b := DecimalField[T]{f: field[T]{path: path, value: value}}
	b.f = b.f.add(step[T]{bail: true, rule: constraint("decimal", "must be a plain decimal number", nil, func(value T) bool {
		_, ok := parseDecimal(string(value))
		return ok
	})})
	return b
}

// Label sets the human-readable field name without changing its output path.
func (b DecimalField[T]) Label(label string) DecimalField[T] { b.f.label = label; return b }

// Message overrides the preceding rule message while preserving its code.
func (b DecimalField[T]) Message(message string) DecimalField[T] {
	b.f = b.f.withMessage(message)
	return b
}

// Using attaches a reusable validator for the exact decimal-text type.
func (b DecimalField[T]) Using(rule Rule[T]) DecimalField[T] { b.f = b.f.using(rule); return b }

// Check attaches a context-aware application check without converting the submitted text.
func (b DecimalField[T]) Check(check func(context.Context, T) error) DecimalField[T] {
	return b.Using(ContextRuleFunc[T](check))
}

// Required rejects empty text before decimal syntax and content checks run.
func (b DecimalField[T]) Required() DecimalField[T] { return b.RequiredIf(true) }

// RequiredIf rejects empty text when condition is true; decimal syntax checks still apply.
func (b DecimalField[T]) RequiredIf(condition bool) DecimalField[T] {
	b.f = b.f.add(step[T]{required: true, rule: constraint("required", "is required", nil, func(value T) bool {
		return !condition || value != ""
	})})
	return b
}

// comparison validates configured bounds before input and compares only exact decimal values.
func (b DecimalField[T]) comparison(code, message string, bound T, parameters map[string]string, valid func(int) bool) DecimalField[T] {
	parsed, ok := parseDecimal(string(bound))
	s := step[T]{rule: constraint(code, message, parameters, func(value T) bool {
		input, ok := parseDecimal(string(value))
		return ok && valid(input.compare(parsed))
	})}
	if !ok {
		s.problem = "invalid decimal comparison value"
	}
	b.f = b.f.add(s)
	return b
}

// Min requires an exact inclusive lower bound expressed as plain decimal text.
func (b DecimalField[T]) Min(min T) DecimalField[T] {
	return b.comparison("min", "must be at least "+string(min), min, map[string]string{"min": string(min)}, func(order int) bool { return order >= 0 })
}

// Max requires an exact inclusive upper bound expressed as plain decimal text.
func (b DecimalField[T]) Max(max T) DecimalField[T] {
	return b.comparison("max", "must be at most "+string(max), max, map[string]string{"max": string(max)}, func(order int) bool { return order <= 0 })
}

// GreaterThan requires a strict lower bound without exposing the comparison value in failures.
func (b DecimalField[T]) GreaterThan(other T) DecimalField[T] {
	return b.comparison("greater_than", "must exceed the comparison value", other, nil, func(order int) bool { return order > 0 })
}

// LessThan requires a strict upper bound without exposing the comparison value in failures.
func (b DecimalField[T]) LessThan(other T) DecimalField[T] {
	return b.comparison("less_than", "must be below the comparison value", other, nil, func(order int) bool { return order < 0 })
}

// Equal compares numeric value exactly without exposing its target; written scale may differ.
func (b DecimalField[T]) Equal(other T) DecimalField[T] {
	return b.comparison("equal", "must match the expected value", other, nil, func(order int) bool { return order == 0 })
}

// MaxScale permits at most max written fractional digits, counting trailing zeros.
func (b DecimalField[T]) MaxScale(max int) DecimalField[T] {
	s := step[T]{rule: constraint("max_scale", "must have at most "+strconv.Itoa(max)+" fractional digits", map[string]string{"max_scale": strconv.Itoa(max)}, func(value T) bool {
		parsed, ok := parseDecimal(string(value))
		return ok && parsed.scale <= max
	})}
	if max < 0 {
		s.problem = "negative maximum decimal scale"
	}
	b.f = b.f.add(s)
	return b
}

// evaluate forwards the typed chain to the shared execution contract.
func (b DecimalField[T]) evaluate(ctx context.Context, prefix []string, issues *[]Issue) error {
	_, err := b.f.run(ctx, prefix, issues)
	return err
}
