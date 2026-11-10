# Validation contracts

## Values and presence

Validation accepts application values after decoding. It does not trim, coerce, bind a request, or infer presence from zero values. Numeric zero and boolean false are valid inputs to their respective rules. Number builders reject NaN and infinities. Comparisons preserve the field's numeric type, including full-width integers.

String lengths count Unicode code points. Required rejects only an empty string; NotBlank also rejects Unicode whitespace. Slice lengths count elements, with Required rejecting both nil and empty slices. RequiredIf(false) disables only that requirement. Other constraints still apply.

Len requires an exact rune or collection count. String Contains, NotContains, StartsWith, and EndsWith use literal case-sensitive text and reject an empty configured search. NotEqual and NotIn keep comparison values out of error details. Numeric GreaterThan and LessThan are exclusive. MultipleOf preserves integer precision; floating-point fields use exact binary remainder semantics, so use integer units for decimal quantities that require exact divisibility.

Slice membership uses reflect.DeepEqual, distinguishing nil and empty collections. Unique uses Go equality for comparable elements; non-comparable elements require UniqueBy with a comparable key. Uniqueness uses linear auxiliary storage and never includes duplicate values in messages. String-keyed maps support length checks, required keys, key allowlists, and Each callbacks receiving both key and value. Empty OnlyKeys permits only an empty map. Map children run in sorted key order, with dots and backslashes escaped in paths; an empty key creates an empty path segment, such as labels.

When constructs its field lazily. A pointer can distinguish a supplied value from nil, but standard JSON decoding maps both absent and null fields to nil pointers. Use Input[T] when an endpoint needs to preserve all three states.

## Rules and errors

Rule[T] implementations can be ordinary application or third-party types. Using accepts the builder's exact type. Format methods preserve defined string types. RuleFunc and ContextRuleFunc adapt functions. RuleWithContext takes precedence over Validate when both are available. Check attaches a context-aware function directly.

Return Invalid(code, message) for an input violation, using a message safe to expose. Other errors are operational and abort the entire evaluation, even if an earlier field failed. Error wrapping preserves errors.Is and errors.As. A wrapper around a Violation remains a violation. Joined errors remain operational if any child is operational; a join containing only violations contributes its first violation to the field.

Fields are visited in declaration order. Required checks run before other constraints. By default, each field stops at its first violation; separate fields and slice items continue. Failed parent slice constraints prevent child evaluation. Each runs only after all parent rules, regardless of where Each appears in the fluent chain.

Message customizes the immediately preceding rule. Label changes built-in prose without changing the path. Custom rule messages remain literal unless explicitly overridden. Messages and Issues return owned copies. Constraint parameters contain limits, not submitted values or secret comparison targets.

Nil fields, rules, regular expressions, and callbacks; empty In lists; negative length bounds; non-finite numeric bounds; and Message without a preceding rule are configuration errors. The evaluator checks a field's known malformed configuration before its input constraints. Custom validator configuration belongs to the validator constructor or its ordinary error return. Dependent validators are never silently skipped when miswired.

## Format policies

Format methods perform no network access. All except IsUTF8 reject empty strings. Their shared implementations are internal to the library.

| Method | Accepted format |
| --- | --- |
| IsEmail | One bare ASCII mailbox, at most 254 bytes overall and 64 bytes before the final @, with a dotted ASCII DNS domain. Display names, quoted local parts, address literals, control characters, and surrounding whitespace are rejected. No DNS or deliverability check. |
| IsURL | Absolute http or https URL with a valid ASCII DNS hostname or IP literal. IPv6 addresses require square brackets; IPv4 addresses must be unbracketed. Localhost and a trailing DNS root dot are accepted. Ports, when present, are 1 through 65535. Whitespace, zones, and invalid escapes (including in the query) are rejected. This is a syntax check, not an SSRF policy. |
| IsUUID | Canonical hyphenated UUID, case insensitive, versions 1 through 8 with the RFC variant, plus nil and max UUIDs. |
| IsUUIDv4 and IsUUIDv7 | Canonical spelling with the named version and RFC variant. |
| IsIP and IsIPv4 | Unzoned address literals without port numbers or CIDR prefixes. IPv4 excludes mapped IPv6 addresses. |
| IsIPv6 | Unzoned IPv6 literals, including IPv4-mapped IPv6 addresses. |
| IsJSON | Any valid JSON document, including scalar values and null. |
| IsBase64 | Nonempty standard padded Base64 with canonical trailing bits and no whitespace. Inputs whose byte length needs no padding remain valid without equals signs. |

## Paths and ownership

Names are literal labels, not expressions for traversing a struct. Group creates parent paths and Each adds numeric index segments. Dots and backslashes inside a label are escaped in rendered paths. An empty child label identifies the item itself. Numeric labels can resemble indexes; use descriptive field labels. Map adds literal string-key segments in sorted order.

Builders own their constraint slices and presentation configuration. Extending a chain does not mutate other branches. Evaluating a chain does not retain failures or cache results. Input slices and maps, compiled regular expressions, and dependencies inside custom rules are borrowed. Do not mutate them during evaluation. Custom rules must not mutate submitted values and must coordinate state shared across goroutines.

Cancellation is checked before evaluation, between fields and rules, and after callbacks. Validate uses a background context, and ValidateContext normalizes a nil context to a background context. Dependent rules must honor the supplied context during blocking work. Panics from application callbacks propagate normally.

Input-size limits remain application policy. Put Max before IsJSON, IsBase64, or a custom Check when an endpoint has a size bound, since content rules run in declaration order. With CollectAll, a failing Max still permits later content checks; use the default policy or StopOnFirstError when that bound must prevent further work. Base64 validation uses fixed scratch storage without retaining decoded data. Slice checks without Each inspect the parent constraints without traversing elements.

## Dates and optional inputs

`Time` compares instants using Go's `time.Time` comparison methods. `Before` and `After` are exclusive; `Min` and `Max` are inclusive. `Required` rejects the zero time. Bounds belong to the caller; the library does not read the clock. `IsDate` parses a Go time layout in UTC, and `IsDateIn` uses the supplied location for unzoned strings. These checks preserve Go's parser behavior and do not normalize the input.

`Optional` skips a nil pointer and invokes its callback lazily for an existing value. A nil callback is a configuration error, including when the value is absent.

`Input[T]` preserves three states: its zero value is absent, `Null[T]()` is explicit null, and `Provided(value)` is a supplied non-null value. JSON decoding changes an input only after successful parsing. Decode into a fresh request value for each request; as with ordinary Go struct fields, decoding an omitted property does not reset existing state. `Get` reports true for supplied zero values such as `0` and `false`. JSON marshaling emits null for absent inputs unless the enclosing property uses `omitzero`, supported by the minimum Go version.

`Presence` skips missing inputs by default, rejects explicit null, and validates supplied values lazily. `Nullable` permits explicit null. `Required` always requires a supplied, non-null value, including on a nullable chain. It permits supplied zero values; apply content rules inside the callback when those should fail. An empty child path inherits the presence path. `Label` and `Message` on the presence chain affect presence failures; child builders own their content messages. A nil callback is always a configuration error. Presence distinguishes supplied JSON values, not application-specific notions of emptiness inside arbitrary domain types. Provided does not inspect Go values for nil pointers or interfaces; use Null to represent an explicit JSON null.

## Additional format policies

All format checks are local syntax checks. `IsULID` accepts the case-insensitive Crockford alphabet and rejects 128-bit overflow, following the [ULID specification](https://github.com/ulid/spec). `IsHostname` accepts ASCII DNS labels, including single-label hosts and an optional trailing root dot; international names must already use their ASCII representation. `IsCIDR` accepts IPv4 or IPv6 prefixes, including nonzero host bits, without rewriting them. `IsMAC` accepts the standard library's spellings for 48-bit and 64-bit addresses.

`IsE164` requires `+`, a nonzero first digit, and 2 through 15 ASCII digits. It does not establish country-code assignment, number ownership, or reachability. `IsBase64URL` accepts canonical padded and unpadded URL-safe encodings, rejects whitespace and invalid trailing bits, and uses bounded scratch storage.

`IsAlpha`, `IsAlphanumeric`, and `IsDigits` require nonempty ASCII characters in their stated classes. `IsASCII` accepts nonempty ASCII text, including control bytes; use a narrower application rule when control bytes are forbidden. `IsLowercase` and `IsUppercase` require valid UTF-8, at least one cased Unicode letter, and no letters of the opposite case or titlecase. Uncased characters such as digits and ideographs are allowed alongside those letters. `IsUTF8` checks encoding only and accepts empty text. Apply `Required` separately where emptiness should fail.

## Composition and evaluation policy

`Any`, `All`, and `Not` accept reusable typed rules and own a copy of their rule list. `Any` stops at the first success; if every branch returns an input violation, it emits code `any`. `All` returns the first failed rule's original violation. `Not` turns a successful rule into code `not` and an input violation into success. All three preserve operational errors and cancellation. Unvisited branches do not execute. Empty compositions and nil rules, including nested compositions, are configuration errors checked before evaluation. Application rules retain responsibility for their own configuration and concurrency safety.

`Validate` and `ValidateContext` retain first-violation-per-field defaults. `New` creates an immutable `Validator` with those same defaults. `CollectAll` evaluates every content rule on a field and preserves order. A required failure or a nonfinite number remains a prerequisite failure and stops that field. Collection child callbacks run only after every parent constraint passes. Operational errors still abort evaluation and take precedence over collected input failures.

`StopOnFirstError` stops globally at the first input violation, including inside groups and collections. It takes precedence over `CollectAll`, and later lazy callbacks do not execute. Configuration in unvisited fields is not evaluated. Each call to a public validation entry point establishes its own policy, including recursive calls from custom rules.

`FormatMessages` receives an issue with its path, stable code, default message, and an owned copy of constraint parameters. Use it to select application translations. It never receives the submitted value. It applies to built-in and application violations; an explicit builder `Message` overrides it. A nil formatter is a configuration error. The application owns formatter state and must synchronize it if the validator is reused concurrently. Labels remain part of the default message; applications can map paths to localized labels in their formatter.
