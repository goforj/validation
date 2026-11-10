package validation

import (
	"context"
	"time"
)

// TimeField is an immutable fluent chain for time instants.
type TimeField struct{ f field[time.Time] }

// Time starts a deferred chain using instant comparisons across time zones.
func Time(path string, value time.Time) TimeField {
	return TimeField{f: field[time.Time]{path: path, value: value}}
}

// Label sets the human-readable field name without changing its output path.
func (b TimeField) Label(label string) TimeField { b.f.label = label; return b }

// Message overrides the preceding rule message while preserving its code.
func (b TimeField) Message(message string) TimeField { b.f = b.f.withMessage(message); return b }

// Using attaches a reusable validator for time values.
func (b TimeField) Using(rule Rule[time.Time]) TimeField { b.f = b.f.using(rule); return b }

// Check attaches a context-aware application check.
func (b TimeField) Check(check func(context.Context, time.Time) error) TimeField {
	return b.Using(ContextRuleFunc[time.Time](check))
}

// Required rejects the zero time before content rules run.
func (b TimeField) Required() TimeField { return b.RequiredIf(true) }

// RequiredIf rejects the zero time when condition is true.
func (b TimeField) RequiredIf(condition bool) TimeField {
	b.f = b.f.add(step[time.Time]{required: true, rule: constraint("required", "is required", nil, func(v time.Time) bool { return !condition || !v.IsZero() })})
	return b
}

// compare records a bound without exposing the submitted instant.
func (b TimeField) compare(code, message string, other time.Time, valid func(time.Time, time.Time) bool) TimeField {
	return b.Using(constraint(code, message, map[string]string{"time": other.Format(time.RFC3339Nano)}, func(v time.Time) bool { return valid(v, other) }))
}

// Before requires an instant strictly earlier than other.
func (b TimeField) Before(other time.Time) TimeField {
	return b.compare("before", "must be before the specified time", other, time.Time.Before)
}

// After requires an instant strictly later than other.
func (b TimeField) After(other time.Time) TimeField {
	return b.compare("after", "must be after the specified time", other, time.Time.After)
}

// Min permits instants on or after min.
func (b TimeField) Min(min time.Time) TimeField {
	return b.compare("min", "must be on or after the specified time", min, func(v, bound time.Time) bool { return !v.Before(bound) })
}

// Max permits instants on or before max.
func (b TimeField) Max(max time.Time) TimeField {
	return b.compare("max", "must be on or before the specified time", max, func(v, bound time.Time) bool { return !v.After(bound) })
}

// Equal requires the same instant without exposing the expected value in failures.
func (b TimeField) Equal(other time.Time) TimeField {
	return b.Using(constraint("equal", "must match the expected time", nil, func(value time.Time) bool { return value.Equal(other) }))
}

// evaluate forwards the typed chain to the shared execution contract.
func (b TimeField) evaluate(ctx context.Context, prefix []string, issues *[]Issue) error {
	_, err := b.f.run(ctx, prefix, issues)
	return err
}

// IsDate requires a value parseable with a Go time layout, using UTC for unzoned input.
func (b StringField[T]) IsDate(layout string) StringField[T] { return b.IsDateIn(layout, time.UTC) }

// IsDateIn interprets unzoned input in location without changing the supplied string.
func (b StringField[T]) IsDateIn(layout string, location *time.Location) StringField[T] {
	s := step[T]{}
	if layout == "" {
		s.problem = "empty date layout"
	} else if location == nil {
		s.problem = "nil date location"
	} else {
		s.rule = constraint("date", "must match the specified date layout", map[string]string{"layout": layout}, func(v T) bool { _, err := time.ParseInLocation(layout, string(v), location); return err == nil })
	}
	b.f = b.f.add(s)
	return b
}
