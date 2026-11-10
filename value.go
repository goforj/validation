package validation

import (
	"context"
)

// ValueField is an immutable fluent chain for value values.
type ValueField[T any] struct{ f field[T] }

// Value starts a deferred chain for a borrowed application value.
func Value[T any](path string, value T) ValueField[T] {
	return ValueField[T]{f: field[T]{path: path, value: value}}
}

// Label sets the human-readable field name without changing its output path.
func (b ValueField[T]) Label(label string) ValueField[T] { b.f.label = label; return b }

// Message overrides the preceding rule message while preserving its code.
func (b ValueField[T]) Message(message string) ValueField[T] {
	b.f = b.f.withMessage(message)
	return b
}

// Using attaches a reusable validator for the exact value type.
func (b ValueField[T]) Using(rule Rule[T]) ValueField[T] { b.f = b.f.using(rule); return b }

// Check attaches a context-aware application check.
func (b ValueField[T]) Check(check func(context.Context, T) error) ValueField[T] {
	b.f = b.f.using(ContextRuleFunc[T](check))
	return b
}

// evaluate forwards the typed chain to the shared execution contract.
func (b ValueField[T]) evaluate(ctx context.Context, prefix []string, issues *[]Issue) error {
	_, err := b.f.run(ctx, prefix, issues)
	return err
}
