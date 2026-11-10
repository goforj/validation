package validation

import (
	"math"
	"reflect"
)

// comparison creates typed comparisons without narrowing integers or exposing target values.
func (b NumberField[T]) comparison(code, message string, other T, valid func(T) bool) NumberField[T] {
	s := step[T]{rule: constraint[T](code, message, nil, valid)}
	if !finite(other) {
		s.problem = "non-finite comparison value"
	}
	b.f = b.f.add(s)
	return b
}

// GreaterThan requires a strict lower bound in the field's numeric type.
func (b NumberField[T]) GreaterThan(other T) NumberField[T] {
	return b.comparison("greater_than", "must exceed the comparison value", other, func(v T) bool { return v > other })
}

// LessThan requires a strict upper bound in the field's numeric type.
func (b NumberField[T]) LessThan(other T) NumberField[T] {
	return b.comparison("less_than", "must be below the comparison value", other, func(v T) bool { return v < other })
}

// NotEqual rejects the other finite value without exposing it in errors.
func (b NumberField[T]) NotEqual(other T) NumberField[T] {
	return b.comparison("not_equal", "must differ from the excluded value", other, func(v T) bool { return v != other })
}

// NotIn rejects the finite listed values and owns its constraint slice.
func (b NumberField[T]) NotIn(values ...T) NumberField[T] {
	values = append([]T(nil), values...)
	s := step[T]{rule: constraint[T]("not_in", "must not be one of the excluded values", nil, func(v T) bool {
		for _, excluded := range values {
			if v == excluded {
				return false
			}
		}
		return true
	})}
	if len(values) == 0 {
		s.problem = "NotIn requires at least one value"
	}
	for _, value := range values {
		if !finite(value) {
			s.problem = "non-finite excluded value"
		}
	}
	b.f = b.f.add(s)
	return b
}

// MultipleOf requires an exact remainder of zero; floats use binary math.Mod semantics.
func (b NumberField[T]) MultipleOf(divisor T) NumberField[T] {
	s := step[T]{rule: constraint[T]("multiple_of", "must be a multiple of the configured divisor", nil, func(v T) bool {
		value, divisor := reflect.ValueOf(v), reflect.ValueOf(divisor)
		if value.Kind() >= reflect.Int && value.Kind() <= reflect.Int64 {
			return value.Int()%divisor.Int() == 0
		}
		if value.Kind() >= reflect.Uint && value.Kind() <= reflect.Uintptr {
			return value.Uint()%divisor.Uint() == 0
		}
		return math.Mod(value.Float(), divisor.Float()) == 0
	})}
	if !finite(divisor) || divisor == 0 {
		s.problem = "divisor must be finite and nonzero"
	}
	b.f = b.f.add(s)
	return b
}
