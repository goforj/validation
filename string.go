package validation

import (
	"context"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/goforj/validation/internal/formats"
)

// StringField is an immutable fluent chain for string values.
type StringField[T ~string] struct{ f field[T] }

// String starts a deferred chain for a borrowed application value.
func String[T ~string](path string, value T) StringField[T] {
	return StringField[T]{f: field[T]{path: path, value: value}}
}

// Label sets the human-readable field name without changing its output path.
func (b StringField[T]) Label(label string) StringField[T] { b.f.label = label; return b }

// Message overrides the preceding rule message while preserving its code.
func (b StringField[T]) Message(message string) StringField[T] {
	b.f = b.f.withMessage(message)
	return b
}

// Using attaches a reusable validator for the exact value type.
func (b StringField[T]) Using(rule Rule[T]) StringField[T] { b.f = b.f.using(rule); return b }

// Check attaches a context-aware application check.
func (b StringField[T]) Check(check func(context.Context, T) error) StringField[T] {
	b.f = b.f.using(ContextRuleFunc[T](check))
	return b
}

// evaluate forwards the typed chain to the shared execution contract.
func (b StringField[T]) evaluate(ctx context.Context, prefix []string, issues *[]Issue) error {
	_, err := b.f.run(ctx, prefix, issues)
	return err
}

// Required rejects an empty string before content rules run.
func (b StringField[T]) Required() StringField[T] { return b.RequiredIf(true) }

// RequiredIf applies an emptiness check only when condition is true.
func (b StringField[T]) RequiredIf(condition bool) StringField[T] {
	b.f = b.f.add(step[T]{rule: constraint[T]("required", "is required", nil, func(v T) bool { return !condition || v != "" }), required: true})
	return b
}

// NotBlank rejects empty strings and strings containing only Unicode whitespace.
func (b StringField[T]) NotBlank() StringField[T] {
	b.f = b.f.using(constraint[T]("not_blank", "must not be blank", nil, func(v T) bool { return strings.TrimSpace(string(v)) != "" }))
	return b
}

// Min requires at least min Unicode code points.
func (b StringField[T]) Min(min int) StringField[T] {
	s := step[T]{rule: constraint[T]("min", "must contain at least "+strconv.Itoa(min)+" characters", map[string]string{"min": strconv.Itoa(min)}, func(v T) bool { return utf8.RuneCountInString(string(v)) >= min })}
	if min < 0 {
		s.problem = "negative minimum length"
	}
	b.f = b.f.add(s)
	return b
}

// Max permits at most max Unicode code points.
func (b StringField[T]) Max(max int) StringField[T] {
	s := step[T]{rule: constraint[T]("max", "must contain at most "+strconv.Itoa(max)+" characters", map[string]string{"max": strconv.Itoa(max)}, func(v T) bool { return utf8.RuneCountInString(string(v)) <= max })}
	if max < 0 {
		s.problem = "negative maximum length"
	}
	b.f = b.f.add(s)
	return b
}

// format adapts shared format checks while preserving application-defined string types.
func (b StringField[T]) format(rule formats.Rule) StringField[T] {
	return b.Check(func(_ context.Context, value T) error { return rule.Validate(string(value)) })
}

// IsEmail requires a bare ASCII mailbox with a dotted domain.
func (b StringField[T]) IsEmail() StringField[T] { return b.format(formats.Email) }

// IsURL requires an absolute HTTP or HTTPS URL with a host.
func (b StringField[T]) IsURL() StringField[T] { return b.format(formats.URL) }

// IsUUID requires a canonical UUID, including nil and max UUIDs.
func (b StringField[T]) IsUUID() StringField[T] { return b.format(formats.UUID) }

// IsUUIDv4 requires a canonical version 4 UUID.
func (b StringField[T]) IsUUIDv4() StringField[T] { return b.format(formats.UUIDv4) }

// IsUUIDv7 requires a canonical version 7 UUID.
func (b StringField[T]) IsUUIDv7() StringField[T] { return b.format(formats.UUIDv7) }

// IsIP requires an unzoned IP address literal.
func (b StringField[T]) IsIP() StringField[T] { return b.format(formats.IP) }

// IsIPv4 requires an IPv4 address literal.
func (b StringField[T]) IsIPv4() StringField[T] { return b.format(formats.IPv4) }

// IsIPv6 requires an unzoned IPv6 address literal.
func (b StringField[T]) IsIPv6() StringField[T] { return b.format(formats.IPv6) }

// IsJSON requires any valid JSON document, including scalar values.
func (b StringField[T]) IsJSON() StringField[T] { return b.format(formats.JSON) }

// IsBase64 requires strict standard padded Base64 without whitespace.
func (b StringField[T]) IsBase64() StringField[T] { return b.format(formats.Base64) }

// Equal requires the other value without retaining it in failure details.
func (b StringField[T]) Equal(other T) StringField[T] {
	b.f = b.f.using(constraint[T]("equal", "must match the expected value", nil, func(v T) bool { return v == other }))
	return b
}

// In permits only the given values and owns a copy of their slice.
func (b StringField[T]) In(values ...T) StringField[T] {
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
	b.f = b.f.add(s)
	return b
}

// Match requires a match against an already compiled, caller-owned expression.
func (b StringField[T]) Match(pattern *regexp.Regexp) StringField[T] {
	if pattern == nil {
		b.f = b.f.add(step[T]{problem: "nil regular expression"})
		return b
	}
	b.f = b.f.using(constraint[T]("match", "has an invalid format", nil, func(v T) bool { return pattern.MatchString(string(v)) }))
	return b
}
