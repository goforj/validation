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
| <img src="docs/assets/check.svg" width="22" height="22" alt="" align="absmiddle"> Typed chains | String, number, boolean, time, slice, map, and domain-value builders expose relevant operations. |
| <img src="docs/assets/check.svg" width="22" height="22" alt="" align="absmiddle"> Reusable rules | Compose reusable rules and plug in custom validator types or functions without global registration. |
| <img src="docs/assets/check.svg" width="22" height="22" alt="" align="absmiddle"> Format checks | Check email, URL, UUID, ULID, network values, encodings, dates, and character classes directly in the fluent chain. |
| <img src="docs/assets/check.svg" width="22" height="22" alt="" align="absmiddle"> Field errors | Ordered paths, stable codes, configurable collection, localized messages, and separate operational errors. |
| <img src="docs/assets/check.svg" width="22" height="22" alt="" align="absmiddle"> Optional inputs | Distinguish omitted properties, explicit null, and supplied zero values without eager callbacks. |

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
- **Errors stay useful.** Validation collects the first violation per field by default. Use `New().CollectAll()` for all content failures or `StopOnFirstError()` to stop globally. Custom messages retain their machine code. Unexpected rule errors abort evaluation and remain operational errors.
- **Custom rules fit the same chain.** Implement Rule[T] with Validate(T) error, or use a function adapter. Return validation.Invalid for an input failure. Add RuleWithContext[T] for dependencies that need cancellation.
- **Evaluation is local.** Builders can be branched and reused. Input slices and application dependencies are borrowed. Coordinate concurrent mutation in your application.

See [validation contracts](docs/contracts.md) for presence, formats, paths, configuration errors, ownership, and cancellation.

## Runnable examples

Run from the repository root with `go run ./examples/<name>`. Tests verify every program's output.

[quickstart](examples/quickstart/main.go) · [formats](examples/formats/main.go) · [patch](examples/patch/main.go) · [nested](examples/nested/main.go) · [custom](examples/custom/main.go) · [context](examples/context/main.go) · [domain](examples/domain/main.go) · [errors](examples/errors/main.go) · [constraints](examples/constraints/main.go) · [advanced](examples/advanced/main.go) · [time](examples/time/main.go) · [presence](examples/presence/main.go) · [moreformats](examples/moreformats/main.go) · [policies](examples/policies/main.go)

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
| **Text constraints** | [Len](#len) · [Contains](#contains) · [NotContains](#notcontains) · [StartsWith](#startswith) · [EndsWith](#endswith) · [NotEqual](#notequal) · [NotIn](#notin) |
| **Numeric constraints** | [GreaterThan](#greaterthan) · [LessThan](#lessthan) · [MultipleOf](#multipleof) |
| **Collection constraints** | [Map](#map) · [Unique](#unique) · [UniqueBy](#uniqueby) · [HasKey](#haskey) · [OnlyKeys](#onlykeys) |
| **Dates and times** | [Time](#time) · [Before](#before) · [After](#after) · [IsDate](#isdate) · [IsDateIn](#isdatein) |
| **Optional inputs** | [Presence](#presence) · [Nullable](#nullable) · [Optional](#optional) · [Provided](#provided) · [Null](#null) · [IsPresent](#ispresent) · [IsNull](#isnull) · [Get](#get) |
| **More formats** | [IsULID](#isulid) · [IsHostname](#ishostname) · [IsCIDR](#iscidr) · [IsMAC](#ismac) · [IsE164](#ise164) · [IsBase64URL](#isbase64url) · [IsAlpha](#isalpha) · [IsAlphanumeric](#isalphanumeric) · [IsDigits](#isdigits) · [IsASCII](#isascii) · [IsLowercase](#islowercase) · [IsUppercase](#isuppercase) · [IsUTF8](#isutf8) |
| **Composition** | [Any](#any) · [All](#all) · [Not](#not) |
| **Evaluation policy** | [New](#new) · [CollectAll](#collectall) · [StopOnFirstError](#stoponfirsterror) · [FormatMessages](#formatmessages) |

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

ValidateContext evaluates with request cancellation and the configured collection policy.

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

## Text constraints

<a id="len"></a>

<a id="contains"></a>

<a id="notcontains"></a>

<a id="startswith"></a>

<a id="endswith"></a>

<a id="notequal"></a>

<a id="notin"></a>

### Len · Contains · NotContains · StartsWith · EndsWith · NotEqual · NotIn

Len requires exactly length Unicode code points.

```go
err := validation.Validate(
	validation.String("sku", "BOOK-42").Len(7).Contains("-").NotContains(" ").StartsWith("BOOK").EndsWith("42"),
	validation.String("username", "ada").NotEqual("root").NotIn("admin", "system"),
)
fmt.Println(err)
// <nil>

```

## Numeric constraints

<a id="greaterthan"></a>

<a id="lessthan"></a>

<a id="multipleof"></a>

### GreaterThan · LessThan · MultipleOf

GreaterThan requires a strict lower bound in the field's numeric type.

```go
err = validation.Validate(
	validation.Number("quantity", 12).GreaterThan(0).LessThan(100).MultipleOf(3),
)
fmt.Println(err)
// <nil>

```

## Collection constraints

<a id="map"></a>

<a id="unique"></a>

<a id="uniqueby"></a>

<a id="haskey"></a>

<a id="onlykeys"></a>

### Map · Unique · UniqueBy · HasKey · OnlyKeys

Map starts a deferred chain for a borrowed string-keyed map.

```go
err = validation.Validate(
	validation.Slice("tags", []string{"go", "books"}).Unique().Contains("go"),
	validation.Slice("products", []Product{{ID: 1}, {ID: 2}}).
		Using(validation.UniqueBy(func(product Product) int { return product.ID })),
	validation.Map("labels", map[string]string{"category": "books"}).
		Required().HasKey("category").OnlyKeys("category").
		Each(func(key, value string) []validation.Field {
			return []validation.Field{validation.String("", value).NotBlank()}
		}),
)
fmt.Println(err)
// <nil>

```

## Dates and times

<a id="time"></a>

<a id="before"></a>

<a id="after"></a>

<a id="isdate"></a>

<a id="isdatein"></a>

### Time · Before · After · IsDate · IsDateIn

Time starts a deferred chain using instant comparisons across time zones.

```go
starts := time.Date(2026, 11, 10, 10, 0, 0, 0, time.UTC)
ends := starts.Add(2 * time.Hour)
fmt.Println(validation.Validate(
	validation.Time("starts_at", starts).Required().Before(ends).Min(starts).Max(ends).Equal(starts),
	validation.Time("ends_at", ends).After(starts),
	validation.String("date", "2026-11-10").IsDate(time.DateOnly),
	validation.String("local_time", "2026-11-10 10:00").IsDateIn("2006-01-02 15:04", time.UTC),
))
// <nil>

```

## Optional inputs

<a id="presence"></a>

<a id="nullable"></a>

<a id="optional"></a>

<a id="provided"></a>

<a id="null"></a>

<a id="ispresent"></a>

<a id="isnull"></a>

<a id="get"></a>

### Presence · Nullable · Optional · Provided · Null · IsPresent · IsNull · Get

Presence defers child validation until input contains a non-null value.
Use an empty child path to retain path, or a named child to append a segment.

```go
// nonBlank shares a content rule across optional and presence-aware inputs.
func nonBlank(value string) validation.Field { return validation.String("", value).NotBlank() }

var patch struct {
	Seats validation.Input[int]    `json:"seats,omitzero"`
	Name  validation.Input[string] `json:"name,omitzero"`
}
fmt.Println(json.Unmarshal([]byte(`{"seats":0,"name":null}`), &patch))
// <nil>
fmt.Println(validation.Validate(
	validation.Presence("seats", patch.Seats, func(v int) validation.Field { return validation.Number("", v).Min(0) }).Required(),
	validation.Presence("name", patch.Name, nonBlank).Nullable(),
))
// <nil>
fmt.Println(patch.Seats.IsPresent(), patch.Name.IsNull())
// true true
value, ok := validation.Provided(false).Get()
fmt.Println(value, ok, validation.Null[string]().IsNull())
// false true true
var nickname *string
fmt.Println(validation.Validate(validation.Optional(nickname, nonBlank)))
// <nil>
name := "Ada"
fmt.Println(validation.Validate(validation.Optional(&name, nonBlank), validation.Presence("name", validation.Provided(name), nonBlank)))
// <nil>

```

## More formats

<a id="isulid"></a>

<a id="ishostname"></a>

<a id="iscidr"></a>

<a id="ismac"></a>

<a id="ise164"></a>

<a id="isbase64url"></a>

<a id="isalpha"></a>

<a id="isalphanumeric"></a>

<a id="isdigits"></a>

<a id="isascii"></a>

<a id="islowercase"></a>

<a id="isuppercase"></a>

<a id="isutf8"></a>

### IsULID · IsHostname · IsCIDR · IsMAC · IsE164 · IsBase64URL · IsAlpha · IsAlphanumeric · IsDigits · IsASCII · IsLowercase · IsUppercase · IsUTF8

IsULID requires a case-insensitive 128-bit ULID.

```go
fmt.Println(validation.Validate(
	validation.String("id", "01ARZ3NDEKTSV4RRFFQ69G5FAV").IsULID(),
	validation.String("host", "api.example.com").IsHostname(),
	validation.String("network", "192.0.2.0/24").IsCIDR(),
	validation.String("device", "00:11:22:33:44:55").IsMAC(),
	validation.String("phone", "+14155552671").IsE164(),
	validation.String("token", "__8").IsBase64URL(),
	validation.String("name", "Ada").IsAlpha(),
	validation.String("sku", "A123").IsAlphanumeric(),
	validation.String("pin", "012345").IsDigits(),
	validation.String("key", "header_name").IsASCII(),
	validation.String("slug", "café").IsLowercase(),
	validation.String("code", "ÉCOLE").IsUppercase(),
	validation.String("title", "こんにちは").IsUTF8(),
))
// <nil>

```

## Composition

<a id="any"></a>

<a id="all"></a>

<a id="not"></a>

### Any · All · Not

Any accepts a value when a rule succeeds; operational errors abort evaluation.

```go
startsA := validation.RuleFunc[string](func(v string) error {
	if strings.HasPrefix(v, "A") {
		return nil
	}
	return validation.Invalid("starts_a", "Must start with A")
})
endsZ := validation.ContextRuleFunc[string](func(_ context.Context, v string) error {
	if strings.HasSuffix(v, "Z") {
		return nil
	}
	return validation.Invalid("ends_z", "Must end with Z")
})
fmt.Println(validation.Validate(validation.String("code", "ABC").Using(validation.All(validation.Any[string](startsA, endsZ), validation.Not[string](endsZ)))))
// <nil>
fmt.Println(validation.Validate(validation.String("code", "XYZ").Using(validation.Any[string](startsA, endsZ))))
// <nil>

```

## Evaluation policy

<a id="new"></a>

<a id="collectall"></a>

<a id="stoponfirsterror"></a>

<a id="formatmessages"></a>

### New · CollectAll · StopOnFirstError · FormatMessages

New creates a validator with first-violation-per-field collection and default messages.

```go
validator := validation.New().CollectAll().FormatMessages(func(issue validation.Issue) string { return issue.Path + ": " + issue.Code })
var failures *validation.Errors
if err := validator.Validate(validation.String("name", "a").Min(3).Contains("z")); errors.As(err, &failures) {
	fmt.Println(failures.Messages()["name"])
	// [name: min name: contains]
}
first := validator.StopOnFirstError()
fmt.Println(first.ValidateContext(context.Background(), validation.String("name", "").Required(), validation.Number("age", 0).Min(18)))
// name: required

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
