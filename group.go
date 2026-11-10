package validation

import "context"

// group holds a literal parent label and independently owned child field storage.
type group struct {
	path   string
	fields []Field
}

// Group prefixes child paths with a literal parent label.
func Group(path string, fields ...Field) Field {
	return group{path: path, fields: append([]Field(nil), fields...)}
}

// evaluate preserves child ordering and propagates operational failures.
func (g group) evaluate(ctx context.Context, prefix []string, issues *[]Issue) error {
	return evaluateFields(ctx, appendPath(prefix, g.path), g.fields, issues)
}

// conditional defers constructing a field until its condition permits evaluation.
type conditional struct {
	condition bool
	makeField func() Field
}

// When evaluates the callback lazily only when condition is true.
func When(condition bool, makeField func() Field) Field { return conditional{condition, makeField} }

// evaluate avoids eager access to optional application values.
func (c conditional) evaluate(ctx context.Context, prefix []string, issues *[]Issue) error {
	if c.makeField == nil {
		return &ConfigurationError{Path: renderPath(prefix), Problem: "nil When callback"}
	}
	if !c.condition {
		return nil
	}
	return evaluateFields(ctx, prefix, []Field{c.makeField()}, issues)
}
