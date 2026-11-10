package validation

import (
	"context"
	"errors"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"testing"
)

// failureAt requires an input failure and returns its public issues.
func failureAt(t *testing.T, field Field, code string) []Issue {
	t.Helper()
	var failures *Errors
	err := Validate(field)
	if !errors.As(err, &failures) {
		t.Fatalf("expected field failure, got %v", err)
	}
	issues := failures.Issues()
	if len(issues) != 1 || issues[0].Code != code {
		t.Fatalf("issues: %+v, want code %s", issues, code)
	}
	return issues
}

// TestRegistration checks valid input, zero-value behavior, and ordered field failures.
func TestRegistration(t *testing.T) {
	if err := Validate(
		String("name", "Ada").Required().NotBlank().Min(2).Max(100),
		String("email", "ada@example.com").Required().Email(),
		Number("age", 36).Min(18).Max(120),
		String("role", "member").In("member", "admin"),
		Bool("terms", true).Accepted(),
	); err != nil {
		t.Fatal(err)
	}
	err := Validate(String("name", "").Min(2).Required(), Number("age", 0).Min(18), Bool("terms", false).Accepted())
	var failures *Errors
	if !errors.As(err, &failures) {
		t.Fatal(err)
	}
	issues := failures.Issues()
	if len(issues) != 3 || issues[0].Code != "required" || issues[1].Path != "age" || issues[2].Path != "terms" {
		t.Fatalf("%+v", issues)
	}
	if failures.First("missing") != "" || failures.First("name") != "The name field is required." {
		t.Fatal(failures.Messages())
	}
	if !strings.Contains(failures.Error(), "; ") {
		t.Fatal(failures.Error())
	}
	if err := Validate(Bool("flag", false).Equal(false), Number("zero", 0).Min(0), String("space", " ").Required()); err != nil {
		t.Fatal(err)
	}
	if err := Validate(); err != nil {
		t.Fatal(err)
	}
	if err := ValidateContext(nil, String("name", "Ada")); err != nil {
		t.Fatal(err)
	}
}

// TestStringRules verifies Unicode lengths, comparison rules, and format shortcuts.
func TestStringRules(t *testing.T) {
	tests := []struct {
		name  string
		field Field
		code  string
	}{
		{"blank", String("name", " \t\u2003").NotBlank(), "not_blank"},
		{"min", String("name", "é").Min(2), "min"},
		{"max", String("name", "日本語").Max(2), "max"},
		{"equal", String("confirmation", "other").Equal("secret"), "equal"},
		{"in", String("role", "owner").In("member", "admin"), "in"},
		{"match", String("sku", "abc").Match(regexp.MustCompile(`^\d+$`)), "match"},
		{"email", String("email", "").Email(), "email"},
		{"url", String("url", "ftp://example.com").URL(), "url"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { failureAt(t, tt.field, tt.code) })
	}
	if err := Validate(String("name", "é日").NotBlank().Min(2).Max(2).Equal("é日"), String("role", "member").In("member"), String("sku", "123").Match(regexp.MustCompile(`^\d+$`)), String("site", "https://example.com").URL()); err != nil {
		t.Fatal(err)
	}
	issue := failureAt(t, String("email", "bad").Label("email address").Email(), "email")[0]
	if issue.Message != "The email address field must be a valid email address." {
		t.Fatal(issue)
	}
	issue = failureAt(t, String("confirmation", "other").Equal("secret"), "equal")[0]
	if strings.Contains(fmt.Sprint(issue), "secret") || strings.Contains(fmt.Sprint(issue), "other") {
		t.Fatal("submitted values leaked")
	}
	allowed := []string{"admin"}
	chain := String("role", "admin").In(allowed...)
	allowed[0] = "member"
	if err := Validate(chain); err != nil {
		t.Fatal(err)
	}
}

// TestNumericRules verifies precision at integer limits and rejects non-finite input.
func TestNumericRules(t *testing.T) {
	if err := Validate(Number("id", uint64(math.MaxUint64)).Min(math.MaxUint64).Max(math.MaxUint64).Equal(math.MaxUint64).In(math.MaxUint64), Number("signed", int64(math.MinInt64)).Min(math.MinInt64), Number("rate", 1.5).Min(1).Max(2)); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		field Field
		code  string
	}{
		{Number("n", 1).Max(0), "max"}, {Number("n", 1).Equal(2), "equal"},
		{Number("n", 1).In(2, 3), "in"}, {Number("n", 0).Min(1), "min"},
		{Number("n", math.NaN()), "finite"}, {Number("n", math.Inf(1)), "finite"},
		{Bool("b", true).Equal(false), "equal"},
	} {
		failureAt(t, tt.field, tt.code)
	}
	allowed := []int{1}
	chain := Number("n", 1).In(allowed...)
	allowed[0] = 2
	if err := Validate(chain); err != nil {
		t.Fatal(err)
	}
}

// TestConfigurationErrors ensures malformed chains cannot be hidden behind an input failure.
func TestConfigurationErrors(t *testing.T) {
	tests := []Field{
		String("s", "").Required().Min(-1), String("s", "").Max(-1), String("s", "").In(),
		String("s", "").Match(nil), String("s", "").Message("bad"), String("s", "").Using(nil),
		String("s", "").Is(nil), String("s", "").Check(nil),
		Number("n", 0.0).Min(math.NaN()), Number("n", 0.0).Max(math.Inf(1)), Number("n", 0.0).Equal(math.NaN()),
		Number("n", 0).In(), Number("n", 0.0).In(math.Inf(-1)),
		Slice[int]("s", nil).Min(-1), Slice[int]("s", nil).Max(-1), Slice[int]("s", nil).Each(nil),
		When(false, nil), When(true, func() Field { return nil }), Group("g", nil), nil,
		String("s", "").Using(RuleFunc[string](nil)),
		Value("v", 1).Message("bad"), Bool("b", true).Using(nil),
	}
	for i, field := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			var configuration *ConfigurationError
			err := Validate(field)
			if !errors.As(err, &configuration) {
				t.Fatalf("expected configuration error: %v", err)
			}
			if !strings.HasPrefix(configuration.Error(), "validation:") {
				t.Fatal(configuration)
			}
		})
	}
	var nilRule *domainRule
	if err := Validate(String("s", "").Using(nilRule)); err == nil {
		t.Fatal("typed nil rule accepted")
	}
	var nilField *StringField[string]
	if err := Validate(nilField); err == nil {
		t.Fatal("typed nil field accepted")
	}
}

// TestConditions exercises lazy optional input and required-only conditions.
func TestConditions(t *testing.T) {
	var email *string
	if err := Validate(When(email != nil, func() Field { return String("email", *email).Required().Email() }), String("company", "").RequiredIf(false).Max(200), Slice[int]("items", nil).RequiredIf(false)); err != nil {
		t.Fatal(err)
	}
	empty := ""
	email = &empty
	failureAt(t, When(email != nil, func() Field { return String("email", *email).Required().Email() }), "required")
	failureAt(t, String("company", "too long").RequiredIf(false).Max(2), "max")
	failureAt(t, String("company", "").Email().RequiredIf(true), "required")
	failureAt(t, Slice[int]("items", nil).RequiredIf(true), "required")
}

// TestCollections verifies parent bail-out, concrete paths, and per-element aggregation.
func TestCollections(t *testing.T) {
	type item struct {
		sku      string
		quantity int
	}
	chain := Slice("items", []item{{"", 0}, {"ok", 1}, {"", -1}}).Required().Min(1).Max(3).Each(func(v item) []Field {
		return []Field{String("sku", v.sku).Required(), Number("quantity", v.quantity).Min(1)}
	})
	var failures *Errors
	err := Validate(Group("order", chain, Group("address", String("city", "").Required())))
	if !errors.As(err, &failures) {
		t.Fatal(err)
	}
	var paths []string
	for _, issue := range failures.Issues() {
		paths = append(paths, issue.Path)
	}
	if !reflect.DeepEqual(paths, []string{"order.items.0.sku", "order.items.0.quantity", "order.items.2.sku", "order.items.2.quantity", "order.address.city"}) {
		t.Fatal(paths)
	}
	calls := 0
	failureAt(t, Slice("tags", []string{"a", "b"}).Max(1).Each(func(v string) []Field { calls++; return []Field{String("", v).Required()} }), "max")
	if calls != 0 {
		t.Fatal("child ran after parent failed")
	}
	failureAt(t, Slice[string]("tags", nil).Min(1), "min")
	failureAt(t, Slice[string]("tags", nil).Required(), "required")
	issue := failureAt(t, Slice("tags", []string{""}).Each(func(v string) []Field { return []Field{String("", v).Required()} }), "required")[0]
	if issue.Path != "tags.0" {
		t.Fatal(issue)
	}
	issue = failureAt(t, Group(`a.b\c`, String("x.y", "").Required()), "required")[0]
	if issue.Path != `a\.b\\c.x\.y` {
		t.Fatal(issue.Path)
	}
	if err := Validate(Group(""), Slice[int]("empty", nil)); err != nil {
		t.Fatal(err)
	}
	base := Slice("items", []int{1})
	one := base.Each(func(int) []Field { calls++; return nil })
	two := base.Each(func(int) []Field { calls += 2; return nil })
	if err := Validate(one, two); err != nil || calls != 3 {
		t.Fatalf("callbacks: %d, error: %v", calls, err)
	}
}

// domainRule is an example third-party validator requiring no library-owned base type.
type domainRule struct{ domain string }

// Validate returns an explicit input violation for an unsupported mailbox domain.
func (r domainRule) Validate(value string) error {
	if strings.HasSuffix(value, "@"+r.domain) {
		return nil
	}
	return Invalid("domain", "Use your company email.")
}

// TestCustomRules verifies configured structs, adapters, domain values, and operational errors.
func TestCustomRules(t *testing.T) {
	if err := Validate(String("email", "a@example.com").Email().Using(domainRule{"example.com"})); err != nil {
		t.Fatal(err)
	}
	issue := failureAt(t, String("email", "a@other.com").Using(domainRule{"example.com"}), "domain")[0]
	if issue.Message != "Use your company email." {
		t.Fatal(issue)
	}
	issue = failureAt(t, String("email", "bad").Using(domainRule{"example.com"}).Message("Custom message"), "domain")[0]
	if issue.Message != "Custom message" {
		t.Fatal(issue)
	}
	wrapped := RuleFunc[string](func(string) error { return fmt.Errorf("wrapped: %w", Invalid("wrapped", "Safe message")) })
	failureAt(t, String("s", "secret").Is(wrapped), "wrapped")
	operational := errors.New("database unavailable")
	err := Validate(String("first", "").Required(), String("second", "value").Using(RuleFunc[string](func(string) error { return operational })).Message("must not mask operational errors"))
	var failures *Errors
	if !errors.Is(err, operational) || errors.As(err, &failures) {
		t.Fatal(err)
	}
	type delivery struct{ days int }
	rule := RuleFunc[delivery](func(d delivery) error {
		if d.days > 30 {
			return Invalid("delivery", "Delivery is too far away.")
		}
		return nil
	})
	if err := Validate(Value("delivery", delivery{5}).Using(rule)); err != nil {
		t.Fatal(err)
	}
	failureAt(t, Value("delivery", delivery{40}).Using(rule), "delivery")
	if got := Invalid("test", "safe"); got.Code() != "test" || got.Error() != "safe" || got.FieldMessage("unused") != "safe" || got.Parameters() != nil {
		t.Fatal(got)
	}
}

// contextualRule ensures dispatch prefers the contextual method of a custom type.
type contextualRule struct {
	called   *bool
	expected context.Context
}

// Validate rejects accidental non-contextual dispatch during a context-aware evaluation.
func (r contextualRule) Validate(string) error { return errors.New("wrong dispatch") }

// ValidateContext confirms the original context reaches a dependent validator.
func (r contextualRule) ValidateContext(ctx context.Context, _ string) error {
	*r.called = true
	if ctx != r.expected {
		return errors.New("wrong context")
	}
	return nil
}

// TestContext verifies dispatch, cancellation precedence, and context adapters.
func TestContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	called := false
	rule := contextualRule{&called, ctx}
	if err := ValidateContext(ctx, String("s", "value").Using(rule), String("s", "value").Is(rule)); err != nil || !called {
		t.Fatalf("called %v, err %v", called, err)
	}
	fn := ContextRuleFunc[string](func(ctx context.Context, _ string) error { return ctx.Err() })
	if err := fn.Validate("ok"); err != nil {
		t.Fatal(err)
	}
	if err := fn.ValidateContext(nil, "ok"); err != nil {
		t.Fatal(err)
	}
	if err := ContextRuleFunc[string](nil).Validate("bad"); err == nil {
		t.Fatal("nil context function accepted")
	}
	if err := RuleFunc[string](nil).Validate("bad"); err == nil {
		t.Fatal("nil function accepted")
	}
	cancel()
	if err := ValidateContext(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	ctx, cancel = context.WithCancel(context.Background())
	err := ValidateContext(ctx, String("s", "value").Check(func(context.Context, string) error { cancel(); return Invalid("bad", "bad") }))
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	ctx, cancel = context.WithCancel(context.Background())
	err = ValidateContext(ctx, When(true, func() Field { cancel(); return String("s", "") }))
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	ctx, cancel = context.WithCancel(context.Background())
	err = ValidateContext(ctx, Slice("items", []int{1, 2}).Each(func(int) []Field { cancel(); return nil }))
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

// TestOwnership verifies field branches, returned copies, and deterministic concurrent reuse.
func TestOwnership(t *testing.T) {
	base := String("name", "x").Min(2)
	left := base.Message("Left")
	right := base.Message("Right")
	if failureAt(t, left, "min")[0].Message != "Left" || failureAt(t, right, "min")[0].Message != "Right" || failureAt(t, base, "min")[0].Message == "Right" {
		t.Fatal("branches share state")
	}
	var failures *Errors
	errors.As(Validate(base), &failures)
	issues := failures.Issues()
	issues[0].Parameters["min"] = "999"
	issues[0].Message = "mutated"
	messages := failures.Messages()
	messages["name"][0] = "mutated"
	if failures.Issues()[0].Parameters["min"] != "2" || failures.First("name") == "mutated" {
		t.Fatal("returned errors share storage")
	}
	source := []Field{String("name", "ok").Required()}
	group := Group("user", source...)
	source[0] = nil
	if err := Validate(group); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := Validate(String("name", "Ada").Required(), Number("age", 30).Min(18)); err != nil {
				t.Error(err)
			}
			if err := Validate(left); err.Error() != "Left" {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
}

// TestDefinedTypes verifies exact custom-rule typing and every builder's common operations.
func TestDefinedTypes(t *testing.T) {
	type name string
	type age uint64
	type accepted bool
	if err := Validate(String("name", name("Ada")).Email().Message("not email")); err == nil {
		t.Fatal("bad defined-string email accepted")
	}
	if err := Validate(
		String("name", name("Ada")).Using(RuleFunc[name](func(name) error { return nil })),
		Number("age", age(30)).Min(18).Using(RuleFunc[age](func(age) error { return nil })).Check(func(context.Context, age) error { return nil }).Label("age").Message("bad age"),
		Bool("terms", accepted(true)).Using(RuleFunc[accepted](func(accepted) error { return nil })).Check(func(context.Context, accepted) error { return nil }).Label("terms").Message("bad terms"),
		Slice("items", []int{1}).Using(RuleFunc[[]int](func([]int) error { return nil })).Check(func(context.Context, []int) error { return nil }).Label("items").Message("bad items"),
		Value("value", struct{}{}).Check(func(context.Context, struct{}) error { return nil }).Label("value").Message("bad value"),
	); err != nil {
		t.Fatal(err)
	}
	if !nilValue(nil) || !nilValue((chan int)(nil)) || !nilValue((map[string]int)(nil)) || !nilValue(([]int)(nil)) || nilValue(1) {
		t.Fatal("nil classification")
	}
}

// cancelField models cancellation occurring as a deferred field finishes.
type cancelField struct{ cancel context.CancelFunc }

// evaluate cancels without returning an error so the enclosing executor must notice.
func (f cancelField) evaluate(context.Context, []string, *[]Issue) error { f.cancel(); return nil }

// TestNestedCancellation verifies cancellation at each execution boundary.
func TestNestedCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	if err := ValidateContext(ctx, cancelField{cancel}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	var issues []Issue
	if err := String("s", "value").Required().evaluate(ctx, nil, &issues); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := Slice("s", []int{1}).Each(func(int) []Field { t.Fatal("callback after cancellation"); return nil }).evaluate(ctx, nil, &issues); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

// TestJoinedErrors prevents a client failure from masking an unavailable dependency.
func TestJoinedErrors(t *testing.T) {
	unavailable := errors.New("directory unavailable")
	joined := fmt.Errorf("lookup: %w", errors.Join(Invalid("domain", "Unsupported domain."), unavailable))
	err := Validate(String("email", "ada@example.com").Using(RuleFunc[string](func(string) error { return joined })))
	var failures *Errors
	if !errors.Is(err, unavailable) || errors.As(err, &failures) {
		t.Fatal(err)
	}
	failureAt(t, String("email", "ada@example.com").Using(RuleFunc[string](func(string) error {
		return errors.Join(Invalid("first", "First failure."), Invalid("second", "Second failure."))
	})), "first")
}
