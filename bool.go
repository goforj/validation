package validation

import (
	"context"
)

// BoolField is an immutable fluent chain for bool values.
type BoolField[T ~bool] struct{ f field[T] }

// Bool starts a deferred chain for a borrowed application value.
func Bool[T ~bool](path string, value T) BoolField[T] {
	return BoolField[T]{f: field[T]{path: path, value: value}}
}

// Label sets the human-readable field name without changing its output path.
func (b BoolField[T]) Label(label string) BoolField[T] { b.f.label = label; return b }

// Message overrides the preceding rule message while preserving its code.
func (b BoolField[T]) Message(message string) BoolField[T] { b.f = b.f.withMessage(message); return b }

// Using attaches a reusable validator for the exact value type.
func (b BoolField[T]) Using(rule Rule[T]) BoolField[T] { b.f = b.f.using(rule); return b }

// Check attaches a context-aware application check.
func (b BoolField[T]) Check(check func(context.Context, T) error) BoolField[T] {
	b.f = b.f.using(ContextRuleFunc[T](check))
	return b
}

// evaluate forwards the typed chain to the shared execution contract.
func (b BoolField[T]) evaluate(ctx context.Context, prefix []string, issues *[]Issue) error {
	_, err := b.f.run(ctx, prefix, issues)
	return err
}

// Accepted requires true while ordinary boolean checks continue to accept false.
func (b BoolField[T]) Accepted() BoolField[T] {
	b.f = b.f.using(constraint[T]("accepted", "must be accepted", nil, func(v T) bool { return bool(v) }))
	return b
}

// Equal requires the supplied boolean value.
func (b BoolField[T]) Equal(other T) BoolField[T] {
	b.f = b.f.using(constraint[T]("equal", "must match the expected value", nil, func(v T) bool { return v == other }))
	return b
}
