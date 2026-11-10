package validation

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"
)

// TestDecimalRules verifies exact bounds, scale, signs, and written representations.
func TestDecimalRules(t *testing.T) {
	for _, f := range []Field{
		Decimal("amount", "0001.2300").Required().RequiredIf(false).Min("1.23").Max("1.23000").GreaterThan("1.22").LessThan("1.24").Equal("+1.23").MaxScale(4),
		Decimal("amount", "-0.00").Min("0").Max("0").Equal("+0").MaxScale(2),
		Decimal("amount", "9007199254740993.01").GreaterThan("9007199254740993.00"),
		Decimal("amount", "0").MaxScale(0),
		Decimal("amount", "-100").Min("-101").Max("-99"),
	} {
		if err := Validate(f); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		field Field
		code  string
	}{
		{Decimal("amount", ""), "decimal"}, {Decimal("amount", "").Required(), "required"},
		{Decimal("amount", "").RequiredIf(false), "decimal"}, {Decimal("amount", "1e3"), "decimal"},
		{Decimal("amount", "0.009").Min("0.01"), "min"}, {Decimal("amount", "100.01").Max("100"), "max"},
		{Decimal("amount", "1.0").GreaterThan("1"), "greater_than"}, {Decimal("amount", "1.0").LessThan("1"), "less_than"},
		{Decimal("amount", "1.00000000000000000001").Equal("1"), "equal"},
		{Decimal("amount", "1.230").MaxScale(2), "max_scale"}, {Decimal("amount", "1.0").MaxScale(0), "max_scale"},
	} {
		failureAt(t, tc.field, tc.code)
	}
	for _, f := range []Field{
		Decimal("amount", "").Required().Min("bad"), Decimal("amount", "1").Max("1e3"),
		Decimal("amount", "1").GreaterThan("NaN"), Decimal("amount", "1").LessThan(""),
		Decimal("amount", "1").Equal("Infinity"), Decimal("amount", "1").MaxScale(-1),
		Decimal("amount", "1").Using(nil), Decimal("amount", "1").Check(nil),
	} {
		var configuration *ConfigurationError
		if err := Validate(f); !errors.As(err, &configuration) {
			t.Fatal(err)
		}
	}
}

// TestDecimalExecution preserves custom-rule types, short circuits malformed input, and isolates branches.
func TestDecimalExecution(t *testing.T) {
	type Amount string
	value := Amount("1.230")
	ctx := context.WithValue(context.Background(), struct{}{}, "request")
	seen := 0
	base := Decimal("amount", value).Label("Amount").Using(RuleFunc[Amount](func(v Amount) error {
		if v != value {
			t.Fatal("custom rule received changed input")
		}
		return nil
	})).Check(func(received context.Context, v Amount) error {
		if received != ctx || v != value {
			t.Fatal("context or defined value type was lost")
		}
		seen++
		return nil
	})
	if err := ValidateContext(ctx, base.MaxScale(3)); err != nil || seen != 1 {
		t.Fatal(err, seen)
	}
	var failures *Errors
	if err := ValidateContext(ctx, base.MaxScale(2).Message("Use at most two fractional digits")); !errors.As(err, &failures) || failures.First("amount") != "Use at most two fractional digits" {
		t.Fatal(err)
	}
	if err := ValidateContext(ctx, base); err != nil {
		t.Fatal(err)
	}
	calls := 0
	for _, validator := range []Validator{New(), New().CollectAll()} {
		err := validator.Validate(Decimal("amount", "not-a-number").Min("0").MaxScale(2).Check(func(context.Context, string) error { calls++; return nil }))
		if !errors.As(err, &failures) || len(failures.Issues()) != 1 || failures.Issues()[0].Code != "decimal" || calls != 0 {
			t.Fatal(err, calls)
		}
	}
	op := errors.New("dependency unavailable")
	if err := Validate(Decimal("amount", "1").Check(func(context.Context, string) error { return op })); !errors.Is(err, op) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := ValidateContext(ctx, Decimal("amount", "1")); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := Validate(Decimal("amount", json.Number("18446744073709551615.01")).Min("18446744073709551615")); err != nil {
		t.Fatal(err)
	}
}

// TestDecimalFailures checks localization data and private equality targets.
func TestDecimalFailures(t *testing.T) {
	var failures *Errors
	err := New().CollectAll().Validate(Decimal("amount", "1234567.891").Min("1234568").MaxScale(2))
	if !errors.As(err, &failures) {
		t.Fatal(err)
	}
	issues := failures.Issues()
	if !reflect.DeepEqual(issues[0].Parameters, map[string]string{"min": "1234568"}) || !reflect.DeepEqual(issues[1].Parameters, map[string]string{"max_scale": "2"}) {
		t.Fatal(issues)
	}
	err = Validate(Decimal("amount", "1234567.891").Equal("9876543.210"))
	if !errors.As(err, &failures) {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(failures.Issues())
	if err != nil || strings.Contains(string(encoded), "1234567") || strings.Contains(string(encoded), "9876543") {
		t.Fatal(string(encoded), err)
	}
}

// TestDecimalConcurrentReuse checks that exact comparisons retain no per-evaluation state.
func TestDecimalConcurrentReuse(t *testing.T) {
	chain := Decimal("amount", "9007199254740993.01").Required().Min("9007199254740993").Max("9007199254740994").MaxScale(2)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				if err := Validate(chain); err != nil {
					t.Error(err)
				}
			}
		}()
	}
	wg.Wait()
}
