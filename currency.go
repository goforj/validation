package validation

import "context"

// CurrencyCode identifies a currency without attaching precision or monetary behavior.
// Package constants provide ISO 4217 identifiers; acceptance requires an explicit allowlist.
type CurrencyCode string

// CurrencyField is an immutable fluent chain with one shared currency allowlist.
type CurrencyField[T ~string] struct {
	f            field[T]
	allowed      []CurrencyCode
	hasAllowlist bool
}

// Currency starts a deferred currency allowlist, configured with selectors such as USD and EUR.
// Select at least one currency before validation; string and defined-string inputs are supported.
func Currency[T ~string](path string, value T) CurrencyField[T] {
	return (CurrencyField[T]{f: field[T]{path: path, value: value}}).withAllowlist()
}

// Label changes human-readable prose without changing the currency field path.
func (b CurrencyField[T]) Label(label string) CurrencyField[T] { b.f.label = label; return b }

// Message overrides the preceding rule message while preserving its code.
func (b CurrencyField[T]) Message(message string) CurrencyField[T] {
	b.f = b.f.withMessage(message)
	return b
}

// Using attaches an application validator for the exact currency-code input type.
func (b CurrencyField[T]) Using(rule Rule[T]) CurrencyField[T] { b.f = b.f.using(rule); return b }

// Check attaches a context-aware application check after the currency allowlist.
func (b CurrencyField[T]) Check(check func(context.Context, T) error) CurrencyField[T] {
	return b.Using(ContextRuleFunc[T](check))
}

// Required rejects an empty currency code before allowlist and application checks.
func (b CurrencyField[T]) Required() CurrencyField[T] { return b.RequiredIf(true) }

// RequiredIf rejects an empty code when condition is true; the allowlist still applies.
func (b CurrencyField[T]) RequiredIf(condition bool) CurrencyField[T] {
	b.f = b.f.add(step[T]{required: true, rule: constraint("required", "is required", nil, func(value T) bool {
		return !condition || value != ""
	})})
	return b
}

// allow owns selection storage so extending one chain cannot change a sibling's currencies.
func (b CurrencyField[T]) allow(code CurrencyCode) CurrencyField[T] {
	b.allowed = append(append([]CurrencyCode(nil), b.allowed...), code)
	return b.withAllowlist()
}

// withAllowlist replaces one prerequisite while preserving its message and other chain steps.
func (b CurrencyField[T]) withAllowlist() CurrencyField[T] {
	allowed := b.allowed
	s := step[T]{bail: true, rule: constraint("currency", "must be one of the allowed currencies", nil, func(value T) bool {
		for _, code := range allowed {
			if string(value) == string(code) {
				return true
			}
		}
		return false
	})}
	if len(allowed) == 0 {
		s.problem = "currency allowlist requires at least one selection"
	}
	if b.hasAllowlist {
		s.message = b.f.steps[0].message
		b.f.steps = append([]step[T](nil), b.f.steps...)
		b.f.steps[0] = s
	} else {
		b.f.steps = append([]step[T]{s}, b.f.steps...)
		b.hasAllowlist = true
	}
	return b
}

// evaluate requires a configured allowlist even for an uninitialized builder value.
func (b CurrencyField[T]) evaluate(ctx context.Context, prefix []string, issues *[]Issue) error {
	if !b.hasAllowlist {
		b = b.withAllowlist()
	}
	_, err := b.f.run(ctx, prefix, issues)
	return err
}
