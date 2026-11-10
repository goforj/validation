# Validation contracts

## Values and presence

Validation accepts application values after decoding. It does not trim, coerce, bind a request, or infer presence from zero values. Numeric zero and boolean false are valid inputs to their respective rules. Number builders reject NaN and infinities. Comparisons preserve the field's numeric type, including full-width integers.

String lengths count Unicode code points. Required rejects only an empty string; NotBlank also rejects Unicode whitespace. Slice lengths count elements, with Required rejecting both nil and empty slices. RequiredIf(false) disables only that requirement. Other constraints still apply.

When constructs its field lazily. A pointer can distinguish a supplied value from nil, but standard JSON decoding maps both absent and null fields to nil pointers. Preserve all three states with an application-owned decoder when an endpoint needs them.

## Rules and errors

Rule[T] implementations can be ordinary application or third-party types. Using accepts the builder's exact type. Format methods preserve defined string types. RuleFunc and ContextRuleFunc adapt functions. RuleWithContext takes precedence over Validate when both are available. Check attaches a context-aware function directly.

Return Invalid(code, message) for an input violation, using a message safe to expose. Other errors are operational and abort the entire evaluation, even if an earlier field failed. Error wrapping preserves errors.Is and errors.As. A wrapper around a Violation remains a violation. Joined errors remain operational if any child is operational; a join containing only violations contributes its first violation to the field.

Fields are visited in declaration order. Required checks run before other constraints. Each field stops at its first violation; separate fields and slice items continue. Failed parent slice constraints prevent child evaluation. Each runs only after all parent rules, regardless of where Each appears in the fluent chain.

Message customizes the immediately preceding rule. Label changes built-in prose without changing the path. Custom rule messages remain literal unless explicitly overridden. Messages and Issues return owned copies. Constraint parameters contain limits, not submitted values or secret comparison targets.

Nil fields, rules, regular expressions, and callbacks; empty In lists; negative length bounds; non-finite numeric bounds; and Message without a preceding rule are configuration errors. The evaluator checks a field's known malformed configuration before its input constraints. Custom validator configuration belongs to the validator constructor or its ordinary error return. Dependent validators are never silently skipped when miswired.

## Format policies

All format methods reject empty strings and perform no network access. Their shared implementations are internal to the library.

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

Names are literal labels, not expressions for traversing a struct. Group creates parent paths and Each adds numeric index segments. Dots and backslashes inside a label are escaped in rendered paths. An empty child label identifies the item itself. Numeric labels can resemble indexes; use descriptive field labels. Dynamic-map addressing is outside this API.

Builders own their constraint slices and presentation configuration. Extending a chain does not mutate other branches. Evaluating a chain does not retain failures or cache results. Input slices, compiled regular expressions, and dependencies inside custom rules are borrowed. Do not mutate them during evaluation. Custom rules must not mutate submitted values and must coordinate state shared across goroutines.

Cancellation is checked before evaluation, between fields and rules, and after callbacks. Validate uses a background context, and ValidateContext normalizes a nil context to a background context. Dependent rules must honor the supplied context during blocking work. Panics from application callbacks propagate normally.
