<p align="center">
  <img src="docs/assets/logo.png" width="300" alt="GoForj validation gopher with a checklist">
</p>

<p align="center">Typed fluent validation with reusable rules and useful field errors.</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/goforj/validation"><img src="https://pkg.go.dev/badge/github.com/goforj/validation.svg" alt="Go Reference"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="MIT License"></a>
  <a href="https://github.com/goforj/validation/actions/workflows/ci.yml"><img src="https://github.com/goforj/validation/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/goforj/validation/actions/workflows/security.yml"><img src="https://github.com/goforj/validation/actions/workflows/security.yml/badge.svg" alt="Security"></a>
  <img src="https://img.shields.io/badge/go-1.24%2B-blue.svg" alt="Go 1.24 or newer">
  <img src="https://img.shields.io/badge/status-unreleased-lightgrey.svg" alt="Unreleased">
  <a href="docs/quality.json"><img src="docs/assets/coverage.svg" alt="Measured unit coverage"></a>
  <a href="docs/quality.json"><img src="docs/assets/tests.svg" alt="Executed Go tests"></a>
</p>

Keep field names, values, and constraints together. Validate decoded requests, configuration, and domain values using ordinary Go code, with no runtime dependencies.

| Feature | What it provides |
| :--- | :--- |
| <img src="docs/assets/check.svg" width="22" height="22" alt=""> Typed chains | String, number, boolean, slice, and domain-value builders expose relevant operations. |
| <img src="docs/assets/check.svg" width="22" height="22" alt=""> Reusable rules | Plug in custom validator types or functions without global registration. |
| <img src="docs/assets/check.svg" width="22" height="22" alt=""> Format checks | Check email, URL, UUID, IP, JSON, and Base64 directly in the fluent chain. |
| <img src="docs/assets/check.svg" width="22" height="22" alt=""> Field errors | Ordered paths, stable codes, custom messages, and separate operational errors. |

## Quick start

Requires Go 1.24 or newer. This library is unreleased. From a source checkout:

```sh
git clone https://github.com/goforj/validation.git
cd validation
go run ./examples/quickstart
```

The complete [registration program](examples/quickstart/main.go) defines the Registration model and imports the library:

```go
req := Registration{
	Name:  "Ada",
	Email: "ada@example.com",
	Age:   36,
	Terms: true,
}
err := validation.Validate(
	validation.String("name", req.Name).Required().Min(2).Max(100),
	validation.String("email", req.Email).Required().IsEmail(),
	validation.Number("age", req.Age).Min(18),
	validation.Bool("terms", req.Terms).Accepted(),
)
fmt.Println(err)
// <nil>

```

## How it works

- **Values are explicit.** Zero and false are ordinary values. Rules evaluate empty strings unless you use a lazy condition to omit a field. Required checks run before content checks.
- **Errors stay useful.** Validation collects the first violation per field. Custom messages retain their machine code. Unexpected rule errors abort evaluation and remain operational errors.
- **Custom rules fit the same chain.** Implement Rule[T] with Validate(T) error, or use a function adapter. Return validation.Invalid for an input failure. Add RuleWithContext[T] for dependencies that need cancellation.
- **Evaluation is local.** Builders can be branched and reused. Input slices and application dependencies are borrowed. Coordinate concurrent mutation in your application.

See [validation contracts](docs/contracts.md) for presence, formats, paths, configuration errors, ownership, and cancellation.

## Runnable examples

Run these from the repository root. Tests execute every program and check its output.

```sh
go run ./examples/custom
go run ./examples/nested
go run ./examples/patch
```

Complete programs:

- [quickstart](examples/quickstart/main.go)
- [formats](examples/formats/main.go)
- [patch](examples/patch/main.go)
- [nested](examples/nested/main.go)
- [custom](examples/custom/main.go)
- [context](examples/context/main.go)
- [domain](examples/domain/main.go)
- [errors](examples/errors/main.go)
- [constraints](examples/constraints/main.go)

## API index

| Group | Functions |
| ---: | :--- |
| **Values** | [String](#string) · [Number](#number) · [Bool](#bool) · [Validate](#validate) |
| **Formats** | [IsEmail](#isemail) · [IsURL](#isurl) · [IsUUID](#isuuid) · [IsUUIDv4](#isuuidv4) · [IsUUIDv7](#isuuidv7) · [IsIP](#isip) · [IsIPv4](#isipv4) · [IsIPv6](#isipv6) · [IsJSON](#isjson) · [IsBase64](#isbase64) |
| **Conditions** | [When](#when) |
| **Collections** | [Group](#group) · [Slice](#slice) · [Each](#each) |
| **Custom rules** | [Using](#using) · [Invalid](#invalid) |
| **Context** | [ValidateContext](#validatecontext) · [Check](#check) |
| **Domain values** | [Value](#value) · [RuleFunc](#rulefunc) · [ContextRuleFunc](#contextrulefunc) |
| **Errors** | [Label](#label) · [Message](#message) · [First](#first) · [Messages](#messages) · [Issues](#issues) |
| **Constraints** | [Required](#required) · [RequiredIf](#requiredif) · [NotBlank](#notblank) · [Min](#min) · [Max](#max) · [In](#in) · [Equal](#equal) · [Match](#match) · [Accepted](#accepted) |

## Values

<a id="string"></a>

<a id="number"></a>

<a id="bool"></a>

<a id="validate"></a>

### String · Number · Bool · Validate

String starts a deferred chain for a borrowed application value.

```go
req := Registration{
	Name:  "Ada",
	Email: "ada@example.com",
	Age:   36,
	Terms: true,
}
err := validation.Validate(
	validation.String("name", req.Name).Required().Min(2).Max(100),
	validation.String("email", req.Email).Required().IsEmail(),
	validation.Number("age", req.Age).Min(18),
	validation.Bool("terms", req.Terms).Accepted(),
)
fmt.Println(err)
// <nil>

```

## Formats

<a id="isemail"></a>

<a id="isurl"></a>

<a id="isuuid"></a>

<a id="isuuidv4"></a>

<a id="isuuidv7"></a>

<a id="isip"></a>

<a id="isipv4"></a>

<a id="isipv6"></a>

<a id="isjson"></a>

<a id="isbase64"></a>

### IsEmail · IsURL · IsUUID · IsUUIDv4 · IsUUIDv7 · IsIP · IsIPv4 · IsIPv6 · IsJSON · IsBase64

IsEmail requires a bare ASCII mailbox with a dotted domain.

```go
err := validation.Validate(
	validation.String("email", "ada@example.com").IsEmail(),
	validation.String("site", "https://goforj.dev").IsURL(),
	validation.String("id", "550e8400-e29b-41d4-a716-446655440000").IsUUID(),
	validation.String("trace_id", "550e8400-e29b-41d4-a716-446655440000").IsUUIDv4(),
	validation.String("request_id", "0190bde7-5ad0-7000-8000-000000000001").IsUUIDv7(),
	validation.String("host", "::1").IsIP(),
	validation.String("ip", "192.0.2.1").IsIPv4(),
	validation.String("ipv6", "2001:db8::1").IsIPv6(),
	validation.String("metadata", `{"source":"web"}`).IsJSON(),
	validation.String("cursor", "YWJj").IsBase64(),
)
fmt.Println(err)
// <nil>

```

## Conditions

<a id="when"></a>

### When

When evaluates the callback lazily only when condition is true.

```go
return validation.Validate(
	validation.When(req.Email != nil, func() validation.Field {
		return validation.String("email", *req.Email).Required().IsEmail()
	}),
)

```

## Collections

<a id="group"></a>

<a id="slice"></a>

<a id="each"></a>

### Group · Slice · Each

Group prefixes child paths with a literal parent label.

```go
items := []Item{
	{SKU: "book", Quantity: 0},
}
err := validation.Validate(
	validation.Group("address", validation.String("city", "").Required()),
	validation.Slice("items", items).Required().Max(100).
		Each(func(item Item) []validation.Field {
			return []validation.Field{
				validation.String("sku", item.SKU).Required(),
				validation.Number("quantity", item.Quantity).Min(1),
			}
		}),
)
failures := err.(*validation.Errors)
for _, issue := range failures.Issues() {
	fmt.Println(issue.Path, issue.Code)
}
// address.city required
// items.0.quantity min

```

## Custom rules

<a id="using"></a>

<a id="invalid"></a>

### Using · Invalid

Using attaches a reusable validator for the exact value type.

```go
// AllowedDomain restricts registrations to a configured company domain.
type AllowedDomain struct {
	Domain string
}

// Validate returns a safe input failure for addresses outside the company.
func (r AllowedDomain) Validate(value string) error {
	_, domain, ok := strings.Cut(value, "@")
	if !ok || !strings.EqualFold(domain, r.Domain) {
		return validation.Invalid("email_domain", "Use your company email address.")
	}
	return nil
}

err := validation.Validate(
	validation.String("email", "ada@other.example").Required().IsEmail().
		Using(AllowedDomain{Domain: "goforj.dev"}),
)
fmt.Println(err)
// Use your company email address.

```

## Context

<a id="validatecontext"></a>

<a id="check"></a>

### ValidateContext · Check

ValidateContext collects field violations while giving operational errors precedence.

```go
unavailable := errors.New("directory unavailable")
err := validation.ValidateContext(context.Background(),
	validation.String("email", "ada@example.com").IsEmail().
		Check(func(ctx context.Context, email string) error {
			// A real directory client would use ctx and email here.
			return unavailable
		}),
)
fmt.Println(errors.Is(err, unavailable))
// true

```

## Domain values

<a id="value"></a>

<a id="rulefunc"></a>

<a id="contextrulefunc"></a>

### Value · RuleFunc · ContextRuleFunc

Value starts a deferred chain for a borrowed application value.

```go
window := validation.RuleFunc[Delivery](func(value Delivery) error {
	if value.Days > 30 {
		return validation.Invalid("delivery_window", "Choose delivery within 30 days.")
	}
	return nil
})
available := validation.ContextRuleFunc[Delivery](func(ctx context.Context, value Delivery) error {
	return ctx.Err()
})
err := validation.Validate(
	validation.Value("delivery", Delivery{Days: 7}).Using(window).Using(available),
)
fmt.Println(err)
// <nil>

```

## Errors

<a id="label"></a>

<a id="message"></a>

<a id="first"></a>

<a id="messages"></a>

<a id="issues"></a>

### Label · Message · First · Messages · Issues

Label sets the human-readable field name without changing its output path.

```go
err := validation.Validate(
	validation.String("email", "bad").Label("email address").IsEmail().
		Message("Enter your work email address."),
)
var failures *validation.Errors
if errors.As(err, &failures) {
	fmt.Println(failures.First("email"))
	// Enter your work email address.
	fmt.Println(failures.Messages()["email"])
	// [Enter your work email address.]
	fmt.Println(failures.Issues()[0].Code)
	// email
}

```

## Constraints

<a id="required"></a>

<a id="requiredif"></a>

<a id="notblank"></a>

<a id="min"></a>

<a id="max"></a>

<a id="in"></a>

<a id="equal"></a>

<a id="match"></a>

<a id="accepted"></a>

### Required · RequiredIf · NotBlank · Min · Max · In · Equal · Match · Accepted

Required rejects an empty string before content rules run.

```go
err := validation.Validate(
	validation.String("name", "Ada").Required().NotBlank().Min(2).Max(100),
	validation.String("company", "").RequiredIf(false).Max(200),
	validation.String("role", "member").In("member", "admin"),
	validation.String("confirmation", "correct horse").Equal("correct horse"),
	validation.String("sku", "BOOK-42").Match(regexp.MustCompile(`^[A-Z]+-\d+$`)),
	validation.Number("age", 36).Min(18).Max(120),
	validation.Bool("terms", true).Accepted(),
)
fmt.Println(err)
// <nil>

```


## Development

```sh
make test
make test-race
make vet
make test-coverage
make generate
make check-generated
```

The root module contains the library and executable examples. The separate docs module contains documentation tooling. See [development](docs/development.md) for checks and [the design record](DESIGN.md) for the research behind the API.

The [initial release review](docs/release-readiness.md) records the review passes, fixes, and verification evidence.
