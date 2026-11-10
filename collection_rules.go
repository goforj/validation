package validation

import (
	"context"
	"reflect"
	"strconv"

	"github.com/goforj/validation/internal/violation"
)

// configuredRule exposes only library-owned construction errors to field preflight.
type configuredRule interface{ configurationProblem() string }

// uniqueRule retains an application-owned key function and no execution state.
type uniqueRule[T any, K comparable] struct{ key func(T) K }

// UniqueBy requires distinct comparable keys without retaining keys in failures.
func UniqueBy[T any, K comparable](key func(T) K) Rule[[]T] { return uniqueRule[T, K]{key: key} }

// configurationProblem rejects a missing key function before field constraints execute.
func (r uniqueRule[T, K]) configurationProblem() string {
	if r.key == nil {
		return "nil uniqueness key function"
	}
	return ""
}

// Validate evaluates uniqueness with a background context.
func (r uniqueRule[T, K]) Validate(values []T) error {
	return r.ValidateContext(context.Background(), values)
}

// ValidateContext checks cancellation between borrowed elements and extracted keys.
func (r uniqueRule[T, K]) ValidateContext(ctx context.Context, values []T) error {
	if problem := r.configurationProblem(); problem != "" {
		return &ConfigurationError{Problem: problem}
	}
	if ctx == nil {
		ctx = context.Background()
	}
	var seen map[K]struct{}
	for _, value := range values {
		if err := ctx.Err(); err != nil {
			return err
		}
		if seen == nil {
			seen = make(map[K]struct{}, len(values))
		}
		key := r.key(value)
		if err := ctx.Err(); err != nil {
			return err
		}
		if dynamic := reflect.ValueOf(key); dynamic.IsValid() && !dynamic.Comparable() {
			return &ConfigurationError{Problem: "uniqueness key must be comparable"}
		}
		if _, exists := seen[key]; exists {
			return violation.Builtin("unique", "must contain distinct values", nil)
		}
		seen[key] = struct{}{}
	}
	return ctx.Err()
}

// Unique requires distinct Go-comparable elements; use UniqueBy for structs with slices or maps.
func (b SliceField[T]) Unique() SliceField[T] {
	return b.Using(UniqueBy(func(value T) any { return value }))
}

// Len requires exactly length elements.
func (b SliceField[T]) Len(length int) SliceField[T] {
	s := step[[]T]{rule: constraint[[]T]("len", "must contain exactly "+strconv.Itoa(length)+" items", map[string]string{"len": strconv.Itoa(length)}, func(v []T) bool { return len(v) == length })}
	if length < 0 {
		s.problem = "negative exact length"
	}
	b.f = b.f.add(s)
	return b
}

// Contains requires an element equal under reflect.DeepEqual; the target value is borrowed.
func (b SliceField[T]) Contains(target T) SliceField[T] {
	return b.Check(func(ctx context.Context, values []T) error {
		for _, value := range values {
			if err := ctx.Err(); err != nil {
				return err
			}
			if reflect.DeepEqual(value, target) {
				return nil
			}
		}
		return violation.Builtin("contains", "must contain the required element", nil)
	})
}

// NotContains rejects an element equal under reflect.DeepEqual; the target value is borrowed.
func (b SliceField[T]) NotContains(target T) SliceField[T] {
	return b.Check(func(ctx context.Context, values []T) error {
		for _, value := range values {
			if err := ctx.Err(); err != nil {
				return err
			}
			if reflect.DeepEqual(value, target) {
				return violation.Builtin("not_contains", "must not contain the excluded element", nil)
			}
		}
		return nil
	})
}
