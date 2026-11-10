# GoForj validation API proposal

Research date: November 9, 2026. Status: design record for the initial implementation. The current behavior and supported scope are documented in [validation contracts](docs/contracts.md).

Build a small standalone library around typed fluent field checks, ordinary Go control flow, and useful structured errors. The common case should require one validation call and one chain per field. Start with typed application data; keep decoding and HTTP responses in the application.

## Research findings

### Laravel

[Laravel validation](https://laravel.com/docs/13.x/validation) groups rules by input field. It supports compact rule strings, rule objects and fluent builders, reusable Form Requests, conditional rules, nested paths and wildcards, custom rules, and customized messages. Its message bag provides field lookup. `validated()` and `safe()` expose validated input. `bail` stops a field after its first failure; stopping the whole validation is a separate option.

Laravel distinguishes `required`, `present`, `nullable`, and `sometimes`. Its `required` rejects missing or empty input, while `sometimes` validates only supplied input. Those distinctions are valuable, but Go structs lose input presence unless decoding preserves it. Laravel also integrates validation with HTTP errors and database checks.

Borrow field-oriented declarations, readable constraints, conditional validation, and helpful errors. Keep HTTP handling and database access outside the core. Go errors and typed values should determine the Go API rather than reproducing PHP's rule syntax.

### Go libraries

The capabilities below come from the linked projects. The last column is a design assessment for GoForj, rather than a claim that an approach is universally better.

| Library | Documented approach | Implication for this design |
| --- | --- | --- |
| [go-playground/validator](https://github.com/go-playground/validator) | Struct tags, cross-field checks, collection diving, custom validators, translation support, and `error` returns. | Tags keep simple declarations short. Complex rules embedded in strings offer less compiler and editor assistance. Keep ordinary error returns. |
| [ozzo-validation](https://github.com/go-ozzo/ozzo-validation) | Rules in Go code, field pointers, ordered checks, conditional rules, custom functions, context support, and internal error distinctions. | Closest baseline for ordinary Go applications. Fluent chains can reduce repeated package qualifiers and field-pointer ceremony. |
| [gookit/validate](https://github.com/gookit/validate) | Struct, map, and request validation; tags and programmatic rules; filtering; scenes; custom messages. | Broad input coverage is useful, but a smaller core can give each operation clearer semantics. |
| [thedevsaddam/govalidator](https://github.com/thedevsaddam/govalidator) | Explicitly Laravel-inspired, with field maps, string rules, custom messages, and request validation. | Reproducing Laravel's syntax in Go already exists. GoForj needs a stronger reason to exist than familiar rule strings. |
| [GoZod](https://github.com/kaptinlin/gozod) | Fluent schemas, parsing, refinements, transforms, struct support, and JSON Schema interoperability. | Fluent discovery works well. A validation-only API can stay smaller than a parsing and schema system. |

These are capability comparisons, not performance rankings. No comparative benchmarks were run. Maintenance and version decisions should be checked again if choosing a dependency.

## Recommended API

```go
import "github.com/goforj/validation"

err := validation.Validate(
    validation.String("name", req.Name).Required().Min(2).Max(100),
    validation.String("email", req.Email).Required().IsEmail(),
    validation.Number("age", req.Age).Min(18).Max(120),
    validation.String("role", req.Role).In("member", "admin"),
    validation.Bool("terms", req.Terms).Accepted(),
)
if err != nil {
    return err
}
```

`String`, `Number`, and `Bool` select the operations available in the chain. `Number` is generic over Go integer and floating-point types, including defined types. `String` and `Bool` should also preserve defined types. A string chain exposes `IsEmail`; a number chain does not. The path is an output label, not a reflective lookup into the request.

The proposed core signatures are:

```go
// Field represents a deferred validation operation.
type Field interface {
    // Unexported implementation method omitted from this sketch.
}

// Validate evaluates fields using a background context.
func Validate(fields ...Field) error

// ValidateContext propagates cancellation to custom checks.
func ValidateContext(ctx context.Context, fields ...Field) error
```

The empty interface body above is illustrative. The implementation must seal `Field` with an unexported method. Custom checks use public rule interfaces and typed callbacks rather than requiring users to implement this interface.

Constructing a chain must not execute validation or perform I/O. Evaluation happens only inside `Validate` or `ValidateContext`. Return an actual nil `error` on success to avoid the typed-nil interface trap.

Chain methods return independent builder values, with copy-on-write storage where needed, so extending one chain cannot alter another. Evaluation keeps its state local. Reusing a field must not retain failures from an earlier call. Referenced slices remain borrowed inputs and must not be mutated during validation; custom callbacks are responsible for their own concurrency safety.

### Why this shape

| Candidate | Assessment |
| --- | --- |
| `validate:"required,email"` | Short for static rules, but conditions and custom parameters become string conventions. |
| `Field("email", value, Required(), Email())` | Composable, but repeats constructors and package names for every rule. |
| `Field("email", value).Required().IsEmail()` | Very concise, but a universal field builder exposes methods that cannot apply to its value type. |
| `String("email", value).Required().IsEmail()` | Recommended. Small extra type cue, useful editor completion, and invalid rule families fail to compile. |
| `Schema(...).Parse(input)` | Useful for dynamic input and schema export, but adds a second data model for applications already using Go structs. |

Explicit paths duplicate JSON field names. Accept that tradeoff initially: rules can validate local variables, computed values, configuration, and domain objects without reflecting over a parent struct. Renaming a Go field remains compiler checked; renaming its external label needs a test.

## Fluent format checks

Format checks belong directly on string builders, with no extra import:

```go
err := validation.Validate(
    validation.String("email", req.Email).Required().IsEmail(),
    validation.String("request_id", req.RequestID).Required().IsUUIDv7(),
    validation.String("ip", req.IP).IsIPv4(),
    validation.String("metadata", req.Metadata).IsJSON(),
)
```

The initial methods are IsEmail, IsURL, IsUUID, IsUUIDv4, IsUUIDv7, IsIP, IsIPv4, IsIPv6, IsJSON, and IsBase64. Each participates in ordinary rule ordering, codes, labels, and message overrides. Defined string types work without casts.

[Ozzo's is package](https://pkg.go.dev/github.com/go-ozzo/ozzo-validation/v4/is) inspired the useful format coverage. The accepted GoForj API puts those operations directly in the fluent chain. Custom validators continue to attach through Using; there is no public is package or generic Is attachment method.

The format contracts are documented in [validation contracts](docs/contracts.md). Character classes, country/currency code lists, and specialized application formats remain follow-up work driven by concrete needs.

## Conditions and related fields

Use typed values directly for comparisons:

```go
err := validation.Validate(
    validation.String("password", req.Password).Required().Min(12),
    validation.String("password_confirmation", req.PasswordConfirmation).
        Required().Equal(req.Password),
    validation.String("company", req.Company).
        RequiredIf(req.AccountType == "business").Max(200),
)
```

`RequiredIf(false)` disables only the required check. Other checks still run, so a supplied company exceeding 200 runes fails regardless of account type. A blank company passes `Max(200)`. Conditional requirements are independent of their position in a chain and run before content checks.

Use a lazy field helper when a whole check is conditional:

```go
err := validation.Validate(
    validation.When(req.ChangePassword, func() validation.Field {
        return validation.String("password", req.Password).Required().Min(12)
    }),
)
```

The callback runs during evaluation only when the condition is true. This supports guarded pointer access without eager dereferencing.

## Empty values and input presence

All supplied values run through their declared rules. Optionality must be explicit. In particular, `Number("age", 0).Min(18)` fails and `String("email", "").IsEmail()` fails.

| Input | Proposed behavior |
| --- | --- |
| Empty string with `Required()` | Fails with code `required`. |
| Whitespace with `Required()` | Passes; whitespace is data. Add `NotBlank()` to reject strings containing only Unicode whitespace. |
| Empty string with `Min(2)` | Fails with code `min`. |
| Zero with a numeric rule | Evaluates normally. Zero is not missing. |
| False with a boolean rule | Evaluates normally. Use `Accepted()` when true is required. |
| Empty or nil slice with `Required()` | Fails. Both contain zero elements. |
| Pointer or nullable wrapper | Must be handled explicitly; scalar constructors accept values. |

`Required()` belongs on strings and slices, where it means nonempty. Numeric and boolean value types cannot establish whether the client supplied a field, so their builders should not advertise `Required()`. Use presence-aware decoding when zero or false must be explicitly supplied.

For a PATCH struct where nil means omitted:

```go
err := validation.Validate(
    validation.When(req.Email != nil, func() validation.Field {
        return validation.String("email", *req.Email).Required().IsEmail()
    }),
)
```

An omitted value skips validation; an explicitly supplied empty string fails. Standard JSON decoding into `*string` treats both missing and null as nil. Endpoints that distinguish missing, null, and value need a presence-aware input wrapper or decoder. Do not claim that pointers solve all three states. A core presence abstraction should wait for a concrete decoder integration.

An `OmitEmpty()` convenience would make form fields shorter, but defer it until there is a clear need. Automatically skipping zero values makes otherwise simple rules surprising.

## Errors and messages

`Validate` returns `nil`, a `*validation.Errors` containing field failures, or an ordinary operational/configuration error. `Errors` keeps issues in declaration order and exposes:

```go
var failures *validation.Errors
if errors.As(err, &failures) {
    fmt.Println(failures.First("email"))
    // The email field must be a valid email address.

    messages := failures.Messages()
    // messages has type map[string][]string.
    _ = messages
}
```

The displayed message assumes an email-format failure. `First(path)` returns an empty string for an unmatched path. `Messages()` returns an independent map and slices. `Issues()` returns an ordered copy of structured issues containing `Path`, `Code`, `Message`, and rule parameters such as the minimum length. Raw submitted values must not appear in issues, JSON, or default messages.

Collect failures across fields and stop at the first failure within each field by default. For slices, collect failures from different elements, stopping within each element's field chain. A failed parent constraint prevents child checks. This keeps forms useful and avoids calling an expensive custom check after an earlier rule has already rejected its input.

Customize a field label and an individual rule message:

```go
validation.String("email", req.Email).
    Label("email address").
    Required().
    Email().Message("Enter a valid email address.")
```

`Message` overrides the immediately preceding rule and leaves its machine code unchanged. A misplaced `Message` produces a configuration error. Labels affect prose, not output paths. Use stable codes such as `required`, `email`, `min`, `max`, `equal`, and `in`. Keep localization possible through codes and parameters; defer translation bundles and mutable global settings.

HTTP handlers can use `Messages()` in their chosen response envelope. The core does not choose status codes or write a response.

## Nested data and slices

Use explicit field groups for nested structs and typed callbacks for collections:

```go
err := validation.Validate(
    validation.Group("address",
        validation.String("city", req.Address.City).Required(),
        validation.String("postal_code", req.Address.PostalCode).Required(),
    ),
    validation.Slice("items", req.Items).Required().Max(100).
        Each(func(item Item) []validation.Field {
            return []validation.Field{
                validation.String("sku", item.SKU).Required(),
                validation.Number("quantity", item.Quantity).Min(1),
            }
        }),
)
```

Failures use concrete paths such as `address.city` and `items.2.quantity`. `Each` runs lazily and visits elements by index. No string expression parser is needed. For scalar slices, an empty child label means the item itself, producing paths such as `tags.2`.

Internally retain path segments. Escape literal dots and backslashes in named segments when rendering dotted paths, and reserve decimal segments for slice indexes. A future dynamic-map adapter will need an unambiguous map-key convention. Do not infer nested traversal from the user-supplied label.

## Custom checks and application dependencies

### Pluggable validator types

Make rule extension a first-class public contract. A reusable validator implements one small interface:

```go
// Rule validates a value without retaining execution state.
type Rule[T any] interface {
    // Validate returns a coded violation or an operational error.
    Validate(value T) error
}
```

Attach it to the appropriate builder with `.Using(rule)`:

```go
// AllowedDomain restricts email addresses to an application's domain.
type AllowedDomain struct {
    Domain string
}

// Validate rejects email addresses outside the configured domain.
func (r AllowedDomain) Validate(value string) error {
    _, domain, ok := strings.Cut(value, "@")
    if !ok || !strings.EqualFold(domain, r.Domain) {
        return validation.Invalid("email_domain", "Use your company email address.")
    }
    return nil
}

err := validation.Validate(
    validation.String("email", req.Email).Required().IsEmail().
        Using(AllowedDomain{Domain: "goforj.dev"}).
        Message("Register with your GoForj email address."),
)
```

Rules own their configuration and dependencies. They need no global registration, reflection, field-name knowledge, or package-specific base type. Ordinary structs in an application or a third-party package can implement `Rule[T]`. Builders supply field paths, labels, ordering, and message overrides. A rule can call another validation library and translate its input failures into `Invalid`; preserve unexpected errors as operational failures.

Provide `RuleFunc[T]` as an adapter for `func(T) error`, and `ContextRuleFunc[T]` for `func(context.Context, T) error`. Both implement `Rule[T]`, so functions and structs use the same attachment method. The context adapter's standalone `Validate` uses a background context.

### Context-aware rules

Rules that need cancellation may also implement:

```go
// RuleWithContext allows evaluation to propagate request cancellation.
type RuleWithContext[T any] interface {
    Rule[T]
    // ValidateContext uses the caller's context for dependent operations.
    ValidateContext(ctx context.Context, value T) error
}
```

During evaluation, prefer `ValidateContext` when implemented; otherwise call `Validate`. A service-backed type can implement `Validate` by forwarding to `ValidateContext` with a background context. `ValidateContext` on the library forwards the supplied context, and plain `Validate` supplies a background context. Check cancellation between fields and rules as well. Callbacks must honor context cancellation during blocking work.

Typed `Check` callbacks remain the shortest way to write a one-off context-aware rule:

```go
err := validation.ValidateContext(ctx,
    validation.String("email", req.Email).Required().IsEmail().
        Check(func(ctx context.Context, email string) error {
            exists, err := users.EmailExists(ctx, email)
            if err != nil {
                return err
            }
            if exists {
                return validation.Invalid("unique", "That email is already registered.")
            }
            return nil
        }),
)
```

`Invalid(code, message)` creates an explicit, client-facing validation failure. Any other callback error aborts validation and is returned as an operational error, preserving `errors.Is` and `errors.As`. It must not be folded into the field message bag. An operational failure takes precedence over collected field failures, so callers cannot accidentally return database errors as ordinary invalid input. Cancellation also aborts evaluation.

`Check` preserves the builder's value type, including defined types. `Invalid` messages must be safe to show to the caller. A uniqueness check improves feedback, but a database unique constraint must still protect the write from races. Do not add ORM-specific `Unique` or `Exists` helpers to the core.

### Built-in and custom rule parity

Built-in format methods use the same execution and error classification as application rules. `.Using(rule)` requires `Rule[T]` for the builder's exact value type, retaining domain type safety. `.Check(fn)` adapts to that same execution path.

Add `Value[T](path, value)` for arbitrary domain types. Its builder exposes `.Using`, `.Check`, `.Label`, and `.Message` without scalar-specific methods:

```go
validation.Value("delivery", req.Delivery).Using(DeliveryWindow{})
```

Here `DeliveryWindow` implements `Rule[Delivery]`, where `req.Delivery` has type `Delivery`. This lets applications validate money, date ranges, coordinates, and other domain values without adding specialized core constructors. Attaching a rule for the wrong type must fail to compile.

Shared format implementations stay internal, so no public subpackage or import cycle is needed.

Malformed custom-rule configuration should be rejected by the rule's constructor or returned as an ordinary configuration error. Do not silently skip a nil rule. Rules must not mutate submitted values; concurrency safety of shared dependencies belongs to the rule implementation. Test reusable structs, function adapters, context dispatch, wrapped violations, message overrides, defined types, domain values, and operational errors through every attachment route.

## Reusing rules

Use helper functions to reuse an entire field chain, and `Rule[T]` implementations to reuse an individual validation:

```go
// emailField keeps email constraints consistent across request types.
func emailField(path, value string) validation.Field {
    return validation.String(path, value).Required().IsEmail().Max(254)
}
```

Applications can expose `Validate() error` methods on request types that assemble these fields. No automatic recursive method invocation is needed. Defer reusable schema objects until ordinary helpers prove insufficient.

## Rule contracts

| Family | Proposed initial methods | Contract |
| --- | --- | --- |
| String | `Required`, `RequiredIf`, `NotBlank`, `Min`, `Max`, `IsEmail`, `IsURL`, `IsUUID`, `IsUUIDv4`, `IsUUIDv7`, `IsIP`, `IsIPv4`, `IsIPv6`, `IsJSON`, `IsBase64`, `In`, `Equal`, `Match` | Length counts Unicode code points. No trimming or coercion. `Match` takes a compiled regular expression. |
| Number | `Min`, `Max`, `In`, `Equal` | Inclusive bounds; preserve signed, unsigned, and floating types without converting integers to float64. Reject NaN and infinities. |
| Bool | `Accepted`, `Equal` | `Accepted` requires true. False is otherwise valid. |
| Slice | `Required`, `RequiredIf`, `Min`, `Max`, `Each` | Length counts elements. Preserve the element type. |
| Value | `Using`, `Check`, `Label`, `Message` | Accept arbitrary domain types and matching typed rules. |
| Common | `Label`, `Message`, typed `Using` and `Check` | Metadata and checks remain local to the chain. |

`IsEmail` accepts a single bare ASCII mailbox with a domain, reject display names and surrounding whitespace, and perform no DNS calls. Use an explicitly documented application email policy rather than claiming full RFC or deliverability verification. `IsURL` requires an absolute HTTP or HTTPS URL with a host. Additional schemes can use a custom check. Unicode mailbox policy, IDNA, and URL edge cases need a focused contract and test pass before release.

Reject malformed rule configuration, such as negative length bounds or a nil regex/callback, as configuration errors during evaluation. Empty `In()` is a configuration error. Incompatible rule methods should fail to compile. For numeric `Min`/`Max`, negative bounds are valid for signed types. A bound outside the field type must fail to compile rather than be silently narrowed.

## First release and follow-up work

Implement the constructors and rules above, the public rule interfaces and function adapters, the format methods, `Validate`, `ValidateContext`, `When`, `Group`, structured errors, and `Invalid`. Aim for standard-library-only dependencies. No selected feature requires generic methods or a recent Go language increase; set the actual minimum Go version after checking GoForj support policy and compiling with the chosen baseline.

Before freezing the API, build executable examples for registration, PATCH, nested orders, a reusable field helper, and a dependency failure through a context-aware callback. Evaluation stays sequential. Test empty values, defined primitive types, integer limits, Unicode lengths, stable ordering, lazy callbacks, cancellation, operational error precedence, independent builder branches, repeated evaluation, and path escaping. Test every validation and configuration failure branch directly. Add compile-failure fixtures for invalid method families and value/bound type mismatches.

Measure a successful request, a request with multiple failures, and a slice-heavy request, including allocations. Compare equivalent work only if a performance comparison is needed. Do not promise zero allocations or superiority before measurement.

Defer tags, string-rule parsing, input coercion, request binding, wildcard lookup, schema export, automatic struct traversal, files/images, translation bundles, and database adapters. Dynamic input validation may merit a separate adapter after the typed API is proven.

Wrapping an existing validator remains an option if delivery speed becomes the main requirement. The proposed contracts for typed chains, presence, ordered errors, and operational failures differ enough that a thin wrapper needs a prototype before it can be assumed to save work. The recommended starting point is a small independent core with a deliberately limited rule set.
