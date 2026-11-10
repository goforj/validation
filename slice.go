package validation

import (
	"context"
	"strconv"
)

// SliceField is an immutable fluent chain for slice values.
type SliceField[T any] struct {
	f           field[[]T]
	each        []func(T) []Field
	eachProblem string
}

// Slice starts a deferred chain for a borrowed application value.
func Slice[T any](path string, value []T) SliceField[T] {
	return SliceField[T]{f: field[[]T]{path: path, value: value}}
}

// Label sets the human-readable field name without changing its output path.
func (b SliceField[T]) Label(label string) SliceField[T] { b.f.label = label; return b }

// Message overrides the preceding rule message while preserving its code.
func (b SliceField[T]) Message(message string) SliceField[T] {
	b.f = b.f.withMessage(message)
	return b
}

// Using attaches a reusable validator for the exact value type.
func (b SliceField[T]) Using(rule Rule[[]T]) SliceField[T] { b.f = b.f.using(rule); return b }

// Check attaches a context-aware application check.
func (b SliceField[T]) Check(check func(context.Context, []T) error) SliceField[T] {
	b.f = b.f.using(ContextRuleFunc[[]T](check))
	return b
}

// Required rejects empty and nil slices before evaluating child fields.
func (b SliceField[T]) Required() SliceField[T] { return b.RequiredIf(true) }

// RequiredIf applies a nonempty constraint only when condition is true.
func (b SliceField[T]) RequiredIf(condition bool) SliceField[T] {
	b.f = b.f.add(step[[]T]{rule: constraint[[]T]("required", "is required", nil, func(v []T) bool { return !condition || len(v) > 0 }), required: true})
	return b
}

// Min requires at least min elements.
func (b SliceField[T]) Min(min int) SliceField[T] {
	s := step[[]T]{rule: constraint[[]T]("min", "must contain at least "+strconv.Itoa(min)+" items", map[string]string{"min": strconv.Itoa(min)}, func(v []T) bool { return len(v) >= min })}
	if min < 0 {
		s.problem = "negative minimum length"
	}
	b.f = b.f.add(s)
	return b
}

// Max permits at most max elements.
func (b SliceField[T]) Max(max int) SliceField[T] {
	s := step[[]T]{rule: constraint[[]T]("max", "must contain at most "+strconv.Itoa(max)+" items", map[string]string{"max": strconv.Itoa(max)}, func(v []T) bool { return len(v) <= max })}
	if max < 0 {
		s.problem = "negative maximum length"
	}
	b.f = b.f.add(s)
	return b
}

// Each lazily constructs child fields for each element after parent checks pass.
func (b SliceField[T]) Each(fields func(T) []Field) SliceField[T] {
	if fields == nil {
		b.eachProblem = "nil Each callback"
		return b
	}
	b.each = append(append([]func(T) []Field(nil), b.each...), fields)
	return b
}

// evaluate collects element failures by index while preserving parent and callback ordering.
func (b SliceField[T]) evaluate(ctx context.Context, prefix []string, issues *[]Issue) error {
	if b.eachProblem != "" {
		return &ConfigurationError{Path: renderPath(appendPath(prefix, b.f.path)), Problem: b.eachProblem}
	}
	valid, err := b.f.run(ctx, prefix, issues)
	if err != nil || !valid {
		return err
	}
	if len(b.each) == 0 {
		return nil
	}
	for i, value := range b.f.value {
		for _, children := range b.each {
			if err := ctx.Err(); err != nil {
				return err
			}
			path := appendPath(appendPath(prefix, b.f.path), strconv.Itoa(i))
			if err := evaluateFields(ctx, path, children(value), issues); err != nil {
				return err
			}
		}
	}
	return nil
}
