package validation

import (
	"context"
	"errors"
	"math"
	"reflect"
	"sync"
	"testing"
)

// TestComposition checks Boolean validity without hiding dependency or configuration errors.
func TestComposition(t *testing.T) {
	pass := RuleFunc[int](func(int) error { return nil })
	fail := RuleFunc[int](func(int) error { return Invalid("custom", "rejected") })
	op := errors.New("offline")
	broken := RuleFunc[int](func(int) error { return op })
	for _, rule := range []Rule[int]{Any(fail, pass), All(pass, pass), Not(fail)} {
		if err := rule.Validate(1); err != nil {
			t.Fatal(err)
		}
	}
	failureAt(t, Value("v", 1).Using(Any(fail, fail)), "any")
	failureAt(t, Value("v", 1).Using(All(pass, fail)), "custom")
	failureAt(t, Value("v", 1).Using(Not(pass)), "not")
	for _, rule := range []Rule[int]{Any(fail, broken, pass), All(pass, broken), Not(broken)} {
		if err := rule.Validate(1); !errors.Is(err, op) {
			t.Fatal(err)
		}
	}
	for _, rule := range []Rule[int]{Any[int](), All[int](), Not[int](nil), Any(pass, All[int](nil))} {
		var config *ConfigurationError
		if err := rule.Validate(1); !errors.As(err, &config) {
			t.Fatal(err)
		}
		if err := Validate(Value("v", 1).Using(rule)); !errors.As(err, &config) {
			t.Fatal(err)
		}
	}
	rules := []Rule[int]{pass}
	saved := All(rules...)
	rules[0] = fail
	if err := saved.Validate(1); err != nil {
		t.Fatal(err)
	}
	contextual := All[int](ContextRuleFunc[int](func(ctx context.Context, v int) error {
		if ctx == nil || v != 1 {
			t.Fatal("bad context or value")
		}
		return nil
	}))
	if err := contextual.(RuleWithContext[int]).ValidateContext(nil, 1); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := contextual.(RuleWithContext[int]).ValidateContext(ctx, 1); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	ctx, cancel = context.WithCancel(context.Background())
	cancelRule := Any(RuleFunc[int](func(int) error { cancel(); return nil }))
	if err := cancelRule.(RuleWithContext[int]).ValidateContext(ctx, 1); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

// TestCollectionPolicies verifies ordered failures, prerequisite short circuits, and lazy global stopping.
func TestCollectionPolicies(t *testing.T) {
	all := New().CollectAll()
	var failures *Errors
	err := all.Validate(String("name", "a").Min(2).Contains("z"), String("empty", "").Min(2).Required().IsEmail())
	if !errors.As(err, &failures) {
		t.Fatal(err)
	}
	var codes []string
	for _, issue := range failures.Issues() {
		codes = append(codes, issue.Path+":"+issue.Code)
	}
	if !reflect.DeepEqual(codes, []string{"name:min", "name:contains", "empty:required"}) {
		t.Fatal(codes)
	}
	if err := New().Validate(String("ok", "good").Min(2)); err != nil {
		t.Fatal(err)
	}
	if err := all.ValidateContext(nil); err != nil {
		t.Fatal(err)
	}
	calls := 0
	failureAt(t, Number("n", math.NaN()).Check(func(context.Context, float64) error { calls++; return nil }), "finite")
	if err := all.Validate(Number("n", math.NaN()).Check(func(context.Context, float64) error { calls++; return nil })); !errors.As(err, &failures) || len(failures.Issues()) != 1 || calls != 0 {
		t.Fatal(err, calls)
	}
	err = all.Validate(Slice("items", []int{1}).Min(2).Max(0).Each(func(int) []Field { t.Fatal("children after parent failure"); return nil }))
	if !errors.As(err, &failures) || len(failures.Issues()) != 2 {
		t.Fatal(err)
	}
	stop := all.StopOnFirstError()
	err = stop.Validate(Slice("items", []int{1, 2}).Each(func(v int) []Field { calls++; return []Field{Number("", v).Min(3).Max(0)} }), Value("later", 0).Check(func(context.Context, int) error { t.Fatal("later field ran"); return nil }))
	if !errors.As(err, &failures) || len(failures.Issues()) != 1 || calls != 1 {
		t.Fatal(err, calls)
	}
	calls = 0
	err = stop.Validate(Map("m", map[string]int{"b": 1, "a": 2}).Each(func(string, int) []Field { calls++; return []Field{String("", "").Required()} }))
	if !errors.As(err, &failures) || failures.Issues()[0].Path != "m.a" || calls != 1 {
		t.Fatal(err, calls)
	}
	op := errors.New("offline")
	if err := all.Validate(String("v", "a").Min(2).Check(func(context.Context, string) error { return op })); !errors.Is(err, op) {
		t.Fatal(err)
	}
}

// TestMessagePolicy preserves explicit messages, stable issue data, and policy scoping.
func TestMessagePolicy(t *testing.T) {
	formatter := New().CollectAll().FormatMessages(func(issue Issue) string {
		if issue.Parameters != nil {
			issue.Parameters["min"] = "changed"
		}
		return "localized " + issue.Path + " " + issue.Code
	})
	var failures *Errors
	err := formatter.Validate(String("name", "a").Min(2).Contains("z").Message("explicit"), Value("custom", 0).Using(RuleFunc[int](func(int) error { return Invalid("domain", "original") })))
	if !errors.As(err, &failures) {
		t.Fatal(err)
	}
	issues := failures.Issues()
	if issues[0].Message != "localized name min" || issues[0].Parameters["min"] != "2" || issues[1].Message != "explicit" || issues[2].Message != "localized custom domain" {
		t.Fatal(issues)
	}
	var config *ConfigurationError
	if err := New().FormatMessages(nil).Validate(); !errors.As(err, &config) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	if err := New().FormatMessages(func(Issue) string { cancel(); return "localized" }).ValidateContext(ctx, String("v", "").Required()); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := formatter.Validate(Value("nested", 0).Check(func(ctx context.Context, _ int) error {
		err := ValidateContext(ctx, String("v", "a").Min(2).Contains("z"))
		var nested *Errors
		if !errors.As(err, &nested) || len(nested.Issues()) != 1 || nested.Issues()[0].Message == "localized v min" {
			t.Fatalf("nested defaults: %v", err)
		}
		return nil
	})); err != nil {
		t.Fatal(err)
	}
}

// TestExpandedConcurrentReuse checks immutable policy branches and shared read-only inputs under the race detector.
func TestExpandedConcurrentReuse(t *testing.T) {
	base := New()
	expanded := base.CollectAll().FormatMessages(func(issue Issue) string { return issue.Code })
	fields := []Field{
		Map("m", map[string]string{"key": "value"}).OnlyKeys("key").Each(func(_, v string) []Field { return []Field{String("", v).Min(1)} }),
		Presence("count", Provided(0), func(v int) Field { return Number("", v).Min(1).Max(-1) }),
		Slice("ids", []int{1, 2, 3}).Using(UniqueBy(func(v int) int { return v })),
		Value("rule", 1).Using(All(RuleFunc[int](func(int) error { return nil }))),
	}
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				for _, test := range []struct {
					v     Validator
					count int
				}{{base, 1}, {expanded, 2}} {
					var failures *Errors
					if err := test.v.Validate(fields...); !errors.As(err, &failures) || len(failures.Issues()) != test.count {
						t.Errorf("policy leaked: %v", err)
					}
				}
			}
		}()
	}
	wg.Wait()
}
