package validation

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

// Len requires exactly length Unicode code points.
func (b StringField[T]) Len(length int) StringField[T] {
	s := step[T]{rule: constraint[T]("len", "must contain exactly "+strconv.Itoa(length)+" characters", map[string]string{"len": strconv.Itoa(length)}, func(v T) bool { return utf8.RuneCountInString(string(v)) == length })}
	if length < 0 {
		s.problem = "negative exact length"
	}
	b.f = b.f.add(s)
	return b
}

// textMatch validates a nonempty configured substring without exposing it in failures.
func (b StringField[T]) textMatch(code, message, text string, match func(string, string) bool) StringField[T] {
	s := step[T]{rule: constraint[T](code, message, nil, func(v T) bool { return match(string(v), text) })}
	if text == "" {
		s.problem = "empty search text"
	}
	b.f = b.f.add(s)
	return b
}

// Contains requires the configured substring without modifying case or whitespace.
func (b StringField[T]) Contains(text string) StringField[T] {
	return b.textMatch("contains", "must contain the required text", text, strings.Contains)
}

// NotContains rejects the configured substring without exposing it in failures.
func (b StringField[T]) NotContains(text string) StringField[T] {
	return b.textMatch("not_contains", "must not contain the excluded text", text, func(v, text string) bool { return !strings.Contains(v, text) })
}

// StartsWith requires a literal prefix.
func (b StringField[T]) StartsWith(prefix string) StringField[T] {
	return b.textMatch("starts_with", "must start with the required prefix", prefix, strings.HasPrefix)
}

// EndsWith requires a literal suffix.
func (b StringField[T]) EndsWith(suffix string) StringField[T] {
	return b.textMatch("ends_with", "must end with the required suffix", suffix, strings.HasSuffix)
}

// NotEqual rejects the other value without retaining it in failure details.
func (b StringField[T]) NotEqual(other T) StringField[T] {
	b.f = b.f.using(constraint[T]("not_equal", "must differ from the excluded value", nil, func(v T) bool { return v != other }))
	return b
}

// NotIn rejects the listed values and owns a copy of their slice.
func (b StringField[T]) NotIn(values ...T) StringField[T] {
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
	b.f = b.f.add(s)
	return b
}
