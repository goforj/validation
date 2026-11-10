package validation

import (
	"context"
	"fmt"
)

// Numeric permits all Go integer and floating-point types, including defined types.
type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64
}

// NumberField is an immutable fluent chain for number values.
type NumberField[T Numeric] struct{ f field[T] }

// Number starts a deferred chain for a borrowed application value.
func Number[T Numeric](path string, value T) NumberField[T] {
	f := NumberField[T]{f: field[T]{path: path, value: value}}
	f.f = f.f.using(constraint[T]("finite", "must be finite", nil, finite[T]))
	return f
}

// Label sets the human-readable field name without changing its output path.
func (b NumberField[T]) Label(label string) NumberField[T] { b.f.label = label; return b }

// Message overrides the preceding rule message while preserving its code.
func (b NumberField[T]) Message(message string) NumberField[T] {
	b.f = b.f.withMessage(message)
	return b
}

// Using attaches a reusable validator for the exact value type.
func (b NumberField[T]) Using(rule Rule[T]) NumberField[T] { b.f = b.f.using(rule); return b }

// Check attaches a context-aware application check.
func (b NumberField[T]) Check(check func(context.Context, T) error) NumberField[T] {
	b.f = b.f.using(ContextRuleFunc[T](check))
	return b
}

// evaluate forwards the typed chain to the shared execution contract.
func (b NumberField[T]) evaluate(ctx context.Context, prefix []string, issues *[]Issue) error {
	_, err := b.f.run(ctx, prefix, issues)
	return err
}

// finite rejects NaN and infinities without narrowing integer values through float64.
func finite[T Numeric](value T) bool { return value == value && value-value == 0 }

// Min requires an inclusive lower bound in the field's exact numeric type.
func (b NumberField[T]) Min(min T) NumberField[T] {
	s := step[T]{rule: constraint[T]("min", "must be at least "+fmt.Sprint(min), map[string]string{"min": fmt.Sprint(min)}, func(v T) bool { return v >= min })}
	if !finite(min) {
		s.problem = "non-finite minimum"
	}
	b.f = b.f.add(s)
	return b
}

// Max requires an inclusive upper bound in the field's exact numeric type.
func (b NumberField[T]) Max(max T) NumberField[T] {
	s := step[T]{rule: constraint[T]("max", "must be at most "+fmt.Sprint(max), map[string]string{"max": fmt.Sprint(max)}, func(v T) bool { return v <= max })}
	if !finite(max) {
		s.problem = "non-finite maximum"
	}
	b.f = b.f.add(s)
	return b
}

// Equal requires the other value without exposing it in error messages.
func (b NumberField[T]) Equal(other T) NumberField[T] {
	s := step[T]{rule: constraint[T]("equal", "must match the expected value", nil, func(v T) bool { return v == other })}
	if !finite(other) {
		s.problem = "non-finite expected value"
	}
	b.f = b.f.add(s)
	return b
}

// In permits only finite listed values and owns its constraint slice.
func (b NumberField[T]) In(values ...T) NumberField[T] {
	values = append([]T(nil), values...)
	s := step[T]{rule: constraint[T]("in", "must be one of the allowed values", nil, func(v T) bool {
		for _, allowed := range values {
			if v == allowed {
				return true
			}
		}
		return false
	})}
	if len(values) == 0 {
		s.problem = "In requires at least one value"
	}
	for _, value := range values {
		if !finite(value) {
			s.problem = "non-finite allowed value"
		}
	}
	b.f = b.f.add(s)
	return b
}
