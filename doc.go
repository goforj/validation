// Package validation provides typed, fluent validation of application values.
// Rules run only during Validate or ValidateContext. Evaluation collects the
// first violation per field by default, in declaration order. Validator configures
// broader collection, global stopping, and message formatting. Unexpected rule errors abort
// evaluation and remain operational errors rather than client-facing messages.
//
// Builders can be branched and reused without sharing evaluation state. Slices
// and application dependencies are borrowed; callers must not mutate inputs
// during evaluation and must coordinate any state used by custom rules.
package validation
