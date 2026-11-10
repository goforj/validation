package validation

import (
	"context"
	"sort"
	"strconv"
)

// MapField validates a borrowed string-keyed map with deterministic child paths.
type MapField[T any] struct {
	f       field[map[string]T]
	each    []func(string, T) []Field
	problem string
}

// Map starts a deferred chain for a borrowed string-keyed map.
func Map[T any](path string, value map[string]T) MapField[T] {
	return MapField[T]{f: field[map[string]T]{path: path, value: value}}
}

// Label changes human-readable prose without changing map paths.
func (b MapField[T]) Label(label string) MapField[T] { b.f.label = label; return b }

// Message customizes the preceding parent constraint while preserving its code.
func (b MapField[T]) Message(message string) MapField[T] { b.f = b.f.withMessage(message); return b }

// Using attaches an application validator for the complete map value.
func (b MapField[T]) Using(rule Rule[map[string]T]) MapField[T] { b.f = b.f.using(rule); return b }

// Check attaches a context-aware application validator for the map.
func (b MapField[T]) Check(check func(context.Context, map[string]T) error) MapField[T] {
	b.f = b.f.using(ContextRuleFunc[map[string]T](check))
	return b
}

// Required rejects nil and empty maps before content checks.
func (b MapField[T]) Required() MapField[T] { return b.RequiredIf(true) }

// RequiredIf enforces a nonempty map only when condition is true.
func (b MapField[T]) RequiredIf(condition bool) MapField[T] {
	b.f = b.f.add(step[map[string]T]{rule: constraint[map[string]T]("required", "is required", nil, func(v map[string]T) bool { return !condition || len(v) > 0 }), required: true})
	return b
}

// size configures entry counts without converting the map or exposing its values.
func (b MapField[T]) size(code, message string, length int, valid func(int) bool) MapField[T] {
	s := step[map[string]T]{rule: constraint[map[string]T](code, message, map[string]string{code: strconv.Itoa(length)}, func(v map[string]T) bool { return valid(len(v)) })}
	if length < 0 {
		s.problem = "negative map length"
	}
	b.f = b.f.add(s)
	return b
}

// Min requires at least min entries.
func (b MapField[T]) Min(min int) MapField[T] {
	return b.size("min", "must contain at least "+strconv.Itoa(min)+" entries", min, func(n int) bool { return n >= min })
}

// Max permits at most max entries.
func (b MapField[T]) Max(max int) MapField[T] {
	return b.size("max", "must contain at most "+strconv.Itoa(max)+" entries", max, func(n int) bool { return n <= max })
}

// Len requires exactly length entries.
func (b MapField[T]) Len(length int) MapField[T] {
	return b.size("len", "must contain exactly "+strconv.Itoa(length)+" entries", length, func(n int) bool { return n == length })
}

// HasKey requires a literal key without exposing it in failure messages.
func (b MapField[T]) HasKey(key string) MapField[T] {
	b.f = b.f.using(constraint[map[string]T]("has_key", "must contain the required key", nil, func(v map[string]T) bool { _, ok := v[key]; return ok }))
	return b
}

// OnlyKeys permits only the listed keys, copying configuration independently of the caller.
func (b MapField[T]) OnlyKeys(keys ...string) MapField[T] {
	allowed := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		allowed[key] = struct{}{}
	}
	b.f = b.f.using(constraint[map[string]T]("only_keys", "contains an unexpected key", nil, func(v map[string]T) bool {
		for key := range v {
			if _, ok := allowed[key]; !ok {
				return false
			}
		}
		return true
	}))
	return b
}

// Each constructs key and value checks in sorted key order after parent constraints pass.
func (b MapField[T]) Each(fields func(string, T) []Field) MapField[T] {
	if fields == nil {
		b.problem = "nil Each callback"
		return b
	}
	b.each = append(append([]func(string, T) []Field(nil), b.each...), fields)
	return b
}

// evaluate preserves literal map keys, including empty keys, in nested issue paths.
func (b MapField[T]) evaluate(ctx context.Context, prefix []string, issues *[]Issue) error {
	if b.problem != "" {
		return &ConfigurationError{Path: renderPath(appendPath(prefix, b.f.path)), Problem: b.problem}
	}
	valid, err := b.f.run(ctx, prefix, issues)
	if err != nil || !valid {
		return err
	}
	if len(b.each) == 0 {
		return nil
	}
	keys := make([]string, 0, len(b.f.value))
	for key := range b.f.value {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		for _, children := range b.each {
			if stopped(ctx, *issues) {
				return nil
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			path := append(appendPath(prefix, b.f.path), key)
			if err := evaluateFields(ctx, path, children(key, b.f.value[key]), issues); err != nil {
				return err
			}
		}
	}
	return nil
}
