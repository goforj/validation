package validation

import (
	"context"
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"
)

// TestCommonConstraints verifies useful success and failure contracts without implicit normalization.
func TestCommonConstraints(t *testing.T) {
	fields := []Field{
		String("name", "é日").Len(2).Contains("日").NotContains("x").StartsWith("é").EndsWith("日").NotEqual("other").NotIn("one", "two"),
		Number("count", 9).GreaterThan(8).LessThan(10).NotEqual(0).NotIn(1, 2).MultipleOf(3),
		Number("unsigned", uint64(math.MaxUint64)).GreaterThan(math.MaxUint64 - 1).MultipleOf(3),
		Number("float", 1.5).MultipleOf(0.5),
		Slice("tags", []string{"a", "b"}).Len(2).Unique().Contains("b").NotContains("c"),
	}
	if err := Validate(fields...); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		field Field
		code  string
	}{
		{String("s", "a").Len(2), "len"}, {String("s", "a").Contains("b"), "contains"}, {String("s", "a").NotContains("a"), "not_contains"},
		{String("s", "a").StartsWith("b"), "starts_with"}, {String("s", "a").EndsWith("b"), "ends_with"}, {String("s", "a").NotEqual("a"), "not_equal"}, {String("s", "a").NotIn("a"), "not_in"},
		{Number("n", 1).GreaterThan(1), "greater_than"}, {Number("n", 1).LessThan(1), "less_than"}, {Number("n", 1).NotEqual(1), "not_equal"}, {Number("n", 1).NotIn(1), "not_in"}, {Number("n", 4).MultipleOf(3), "multiple_of"},
		{Slice("s", []int{1}).Len(2), "len"}, {Slice("s", []int{1, 1}).Unique(), "unique"}, {Slice("s", []int{1}).Contains(2), "contains"}, {Slice("s", []int{1}).NotContains(1), "not_contains"},
	} {
		failureAt(t, test.field, test.code)
	}
	type item struct {
		ID     int
		Values []int
	}
	if err := Validate(Slice("items", []item{{ID: 1}, {ID: 2}}).Using(UniqueBy(func(v item) int { return v.ID }))); err != nil {
		t.Fatal(err)
	}
	failureAt(t, Slice("items", []item{{ID: 1}, {ID: 1}}).Using(UniqueBy(func(v item) int { return v.ID })), "unique")
	if err := Validate(Slice("nil", []any{nil, 1}).Unique(), Slice("values", []map[string]int{{"a": 1}}).Contains(map[string]int{"a": 1})); err != nil {
		t.Fatal(err)
	}
	values := []string{"x"}
	field := String("s", "ok").NotIn(values...)
	values[0] = "ok"
	if err := Validate(field); err != nil {
		t.Fatal(err)
	}
	excluded := []int{1}
	number := Number("n", 2).NotIn(excluded...)
	excluded[0] = 2
	if err := Validate(number); err != nil {
		t.Fatal(err)
	}
}

// TestConstraintConfiguration rejects malformed rules before earlier content failures can hide them.
func TestConstraintConfiguration(t *testing.T) {
	for _, field := range []Field{
		String("s", "").Required().Len(-1), String("s", "x").Contains(""), String("s", "x").NotIn(),
		Number("n", 0.0).GreaterThan(math.NaN()), Number("n", 0.0).LessThan(math.Inf(1)), Number("n", 0.0).NotEqual(math.NaN()),
		Number("n", 0).NotIn(), Number("n", 0.0).NotIn(math.Inf(1)), Number("n", 0).MultipleOf(0), Number("n", 0.0).MultipleOf(math.NaN()),
		Slice[int]("s", nil).Len(-1), Slice("s", []map[string]int{{"x": 1}}).Unique(), Slice[int]("s", nil).Required().Using(UniqueBy[int, int](nil)),
	} {
		var configuration *ConfigurationError
		if err := Validate(field); !errors.As(err, &configuration) {
			t.Fatalf("configuration accepted: %v", err)
		}
	}
	if err := UniqueBy[int, int](nil).Validate(nil); err == nil {
		t.Fatal("missing key function accepted")
	}
}

// TestCollectionContexts verifies reusable uniqueness checks and membership callbacks honor cancellation.
func TestCollectionContexts(t *testing.T) {
	rule := UniqueBy(func(v int) int { return v })
	if err := rule.(RuleWithContext[[]int]).ValidateContext(nil, []int{1}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := rule.(RuleWithContext[[]int]).ValidateContext(ctx, []int{1}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	ctx, cancel = context.WithCancel(context.Background())
	keyed := UniqueBy(func(v int) int { cancel(); return v })
	if err := keyed.(RuleWithContext[[]int]).ValidateContext(ctx, []int{1}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	var issues []Issue
	if err := Slice("s", []int{1}).Contains(1).evaluate(ctx, nil, &issues); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := Slice("s", []int{1}).NotContains(2).evaluate(ctx, nil, &issues); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	// Directly invoking the stored contextual callbacks exercises cancellation during a rule boundary.
	for _, chain := range []SliceField[int]{Slice("s", []int{1}).Contains(1), Slice("s", []int{1}).NotContains(2)} {
		if err := chain.f.steps[0].rule.(RuleWithContext[[]int]).ValidateContext(ctx, []int{1}); !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	}
}

// TestMapValidation checks key allowlists, deterministic paths, and independent callbacks.
func TestMapValidation(t *testing.T) {
	values := map[string]string{"z": "", "a.b": "", "": ""}
	field := Map("settings", values).Required().RequiredIf(false).Min(1).Max(3).Len(3).HasKey("z").OnlyKeys("z", "a.b", "").Each(func(key, value string) []Field { return []Field{String("", value).Required()} })
	var failures *Errors
	if err := Validate(field); !errors.As(err, &failures) {
		t.Fatal(err)
	}
	var paths []string
	for _, issue := range failures.Issues() {
		paths = append(paths, issue.Path)
	}
	if !reflect.DeepEqual(paths, []string{"settings.", `settings.a\.b`, "settings.z"}) {
		t.Fatal(paths)
	}
	for _, test := range []struct {
		field Field
		code  string
	}{
		{Map[int]("m", nil).Required(), "required"}, {Map[int]("m", nil).Min(1), "min"}, {Map("m", map[string]int{"x": 1}).Max(0), "max"},
		{Map[int]("m", nil).Len(1), "len"}, {Map[int]("m", nil).HasKey("x"), "has_key"}, {Map("m", map[string]int{"x": 1}).OnlyKeys(), "only_keys"},
	} {
		failureAt(t, test.field, test.code)
	}
	for _, field := range []Field{Map[int]("m", nil).Min(-1), Map[int]("m", nil).Max(-1), Map[int]("m", nil).Len(-1), Map[int]("m", nil).Each(nil), Map[int]("m", nil).Message("bad"), Map[int]("m", nil).Using(nil)} {
		var configuration *ConfigurationError
		if err := Validate(field); !errors.As(err, &configuration) {
			t.Fatal(err)
		}
	}
	good := Map("m", map[string]int{"x": 1}).Label("map").Check(func(context.Context, map[string]int) error { return nil }).Message("bad").Using(RuleFunc[map[string]int](func(map[string]int) error { return nil }))
	if err := Validate(good, Map[int]("empty", nil).OnlyKeys()); err != nil {
		t.Fatal(err)
	}
	calls := 0
	base := Map("m", map[string]int{"x": 1})
	one := base.Each(func(string, int) []Field { calls++; return nil })
	two := base.Each(func(string, int) []Field { calls += 2; return nil })
	if err := Validate(one, two); err != nil || calls != 3 {
		t.Fatalf("%d: %v", calls, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var issues []Issue
	if err := one.evaluate(ctx, nil, &issues); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	op := errors.New("dependency down")
	err := Validate(base.Each(func(string, int) []Field {
		return []Field{Value("", 0).Check(func(context.Context, int) error { return op })}
	}))
	if !errors.Is(err, op) {
		t.Fatal(err)
	}
	if !strings.Contains(failures.Error(), "required") {
		t.Fatal(failures)
	}
}
