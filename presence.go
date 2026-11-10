package validation

import (
	"bytes"
	"context"
	"encoding/json"
)

// Input preserves whether JSON supplied a value, null, or no property at all.
// Its zero value is absent; use json's omitzero option to omit absent properties.
// Decoding T uses json.Unmarshal, so enclosing decoder options do not propagate.
type Input[T any] struct {
	value         T
	present, null bool
}

// Provided constructs an input containing a supplied value, including zero values.
func Provided[T any](value T) Input[T] { return Input[T]{value: value, present: true} }

// Null constructs an explicitly null input.
func Null[T any]() Input[T] { return Input[T]{present: true, null: true} }

// IsPresent reports whether a property was supplied, including explicit null.
func (i Input[T]) IsPresent() bool { return i.present }

// IsNull reports whether a supplied property was explicitly null.
func (i Input[T]) IsNull() bool { return i.null }

// IsZero lets encoding/json omit absent properties with omitzero.
func (i Input[T]) IsZero() bool { return !i.present }

// Get returns the supplied value and reports whether it is present and non-null.
func (i Input[T]) Get() (T, bool) { return i.value, i.present && !i.null }

// UnmarshalJSON replaces the input only after a complete successful decode.
func (i *Input[T]) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.Trim(data, " \t\r\n"), []byte("null")) {
		*i = Null[T]()
		return nil
	}
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*i = Provided(value)
	return nil
}

// MarshalJSON emits null for absent or null inputs and otherwise encodes the value.
func (i Input[T]) MarshalJSON() ([]byte, error) {
	if !i.present || i.null {
		return []byte("null"), nil
	}
	return json.Marshal(i.value)
}

// optionalField defers construction until a pointer value exists.
type optionalField[T any] struct {
	value *T
	build func(T) Field
}

// Optional skips nil pointers and builds validation lazily for non-nil values.
func Optional[T any](value *T, build func(T) Field) Field {
	return optionalField[T]{value: value, build: build}
}

// evaluate rejects invalid configuration even when the optional value is absent.
func (b optionalField[T]) evaluate(ctx context.Context, prefix []string, issues *[]Issue) error {
	if b.build == nil {
		return &ConfigurationError{Path: renderPath(prefix), Problem: "nil optional field callback"}
	}
	if b.value == nil {
		return nil
	}
	return evaluateFields(ctx, prefix, []Field{b.build(*b.value)}, issues)
}

// PresenceField validates an Input while preserving omitted and explicit-null states.
type PresenceField[T any] struct {
	f        field[Input[T]]
	build    func(T) Field
	nullable bool
}

// Presence defers child validation until input contains a non-null value.
// Use an empty child path to retain path, or a named child to append a segment.
func Presence[T any](path string, input Input[T], build func(T) Field) PresenceField[T] {
	return PresenceField[T]{f: field[Input[T]]{path: path, value: input}, build: build}
}

// Label sets the human-readable name for presence failures.
func (b PresenceField[T]) Label(label string) PresenceField[T] { b.f.label = label; return b }

// Message overrides the preceding presence rule message while preserving its code.
func (b PresenceField[T]) Message(message string) PresenceField[T] {
	b.f = b.f.withMessage(message)
	return b
}

// Required requires a supplied, non-null value; zero values remain valid.
func (b PresenceField[T]) Required() PresenceField[T] { return b.RequiredIf(true) }

// RequiredIf requires a supplied, non-null value when condition is true.
func (b PresenceField[T]) RequiredIf(condition bool) PresenceField[T] {
	b.f = b.f.add(step[Input[T]]{required: true, rule: constraint("required", "is required", nil, func(i Input[T]) bool { return !condition || i.present && !i.null })})
	return b
}

// Nullable permits explicit null without constructing child validation.
func (b PresenceField[T]) Nullable() PresenceField[T] { b.nullable = true; return b }

// evaluate checks presence policies before invoking application callbacks.
func (b PresenceField[T]) evaluate(ctx context.Context, prefix []string, issues *[]Issue) error {
	if b.build == nil {
		return &ConfigurationError{Path: renderPath(appendPath(prefix, b.f.path)), Problem: "nil presence field callback"}
	}
	if !b.nullable {
		b.f = b.f.using(constraint("null", "must not be null", nil, func(i Input[T]) bool { return !i.null }))
	}
	valid, err := b.f.run(ctx, prefix, issues)
	if err != nil || !valid {
		return err
	}
	if !b.f.value.present || b.f.value.null {
		return nil
	}
	return evaluateFields(ctx, appendPath(prefix, b.f.path), []Field{b.build(b.f.value.value)}, issues)
}
